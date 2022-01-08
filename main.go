package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

type Hints struct {
	Fixed  []Hint
	Moving []Hint
	Bad    string
}

type Hint struct {
	Letter byte
	Index  byte
}

func (hints *Hints) key() string {
	key := make([]byte, 0, 2*len(hints.Fixed)+1+2*len(hints.Moving)+1+len(hints.Bad))
	for _, h := range hints.Fixed {
		key = append(key, h.Letter)
		key = append(key, h.Index)
	}
	key = append(key, '+')
	for _, h := range hints.Moving {
		key = append(key, h.Letter)
		key = append(key, h.Index)
	}
	key = append(key, '-')
	key = append(key, hints.Bad...)
	return string(key)
}

var distFlag = flag.String("dist", "", "calculate distribution for one word")
var evalFlag = flag.Bool("eval", false, "find best starting word")

func main() {
	flag.Parse()

	if *distFlag != "" {
		distribution(*distFlag)
		return
	}

	if *evalFlag {
		evaluate()
		return
	}

	if flag.NArg() != 0 {
		lookup(flag.Args())
	}
}

func lookup(args []string) {
	words := loadDict()
	argRegex := regexp.MustCompile(`^([_A-Za-z]{5})(?:\+((?:[A-Za-z][1-5])+))?(?:-([A-Za-z]+))?$`)

	for i, arg := range args {
		matches := argRegex.FindStringSubmatch(arg)
		if matches == nil {
			fmt.Printf("invalid argument %q\n", arg)
			continue
		}

		var hints Hints
		fixed := strings.ToUpper(matches[1])
		for i := range fixed {
			if fixed[i] != '_' {
				hints.Fixed = append(hints.Fixed, Hint{
					Letter: fixed[i],
					Index:  byte(i),
				})
			}
		}
		moving := strings.ToUpper(matches[2])
		for i := 0; i < len(moving); i += 2 {
			hints.Moving = append(hints.Moving, Hint{
				Letter: moving[i],
				Index:  moving[i+1] - '1',
			})
		}
		hints.Bad = strings.ToUpper(matches[3])

		for _, word := range words {
			if matchesHints(word, hints) {
				fmt.Println(word)
			}
		}

		if i+1 < flag.NArg() {
			fmt.Println(strings.Repeat("=", 20))
		}
	}
}

func distribution(word string) {
	words := loadDict()
	cache := make(map[string]int64)
	dist := make(map[int64]int64)
	for _, target := range words {
		hints := calculateHints(target, word)
		key := hints.key()
		if count, ok := cache[key]; ok {
			dist[count]++
			continue
		}
		var count int64
		for _, word := range words {
			if matchesHints(word, hints) {
				count++
			}
		}
		dist[count]++
		cache[key] = count
	}

	keys := make([]int64, 0, len(dist))
	for key := range dist {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	const numBuckets = 10

	key := keys[len(keys)-1]
	bucketSize, increment := int64(1), int64(1)
	for key >= numBuckets*bucketSize {
		bucketSize += increment
		if (bucketSize/increment)%10 == 0 {
			increment *= 10
		}
	}

	buckets := make([]int64, numBuckets)
	for key, count := range dist {
		index := key / bucketSize
		buckets[index] += count
	}

	var maxCount int64
	for _, count := range buckets {
		if count > maxCount {
			maxCount = count
		}
	}

	formatLength := len(fmt.Sprintf("%d ", (numBuckets-1)*bucketSize))
	graphLength := int64(100 - formatLength)
	bucketScale := (maxCount + graphLength - 1) / graphLength

	for i, count := range buckets {
		barLength := int(count / bucketScale)
		fmt.Printf("%*d %s\n", formatLength-1, int64(i)*bucketSize, strings.Repeat("*", barLength))
	}
}

type Record struct {
	Index int
	Word  string
	Score int64
}

func evaluate() {
	words := loadDict()

	inputs := make(chan Record, 64)
	outputs := make(chan Record, 64)

	go func() {
		for i, word := range words {
			inputs <- Record{
				Index: i,
				Word:  word,
			}
		}
		close(inputs)
	}()

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			cache := make(map[string]int64)
			for record := range inputs {
				for _, target := range words {
					hints := calculateHints(target, record.Word)
					key := hints.key()
					if count, ok := cache[key]; ok {
						record.Score += count
						continue
					}
					var count int64
					for _, word := range words {
						if matchesHints(word, hints) {
							count++
						}
					}
					record.Score += count
					cache[key] = count
				}
				outputs <- record
			}
		}()
	}

	records := make([]Record, len(words))
	for i := range records {
		record := <-outputs
		records[record.Index] = record
		fmt.Println(i+1, "/", len(words))
	}

	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Score < records[j].Score
	})

	for i := 0; i < len(records) && i < 20; i++ {
		record := records[i]
		fmt.Printf("%s %.3f%%\n", record.Word, float64(100*record.Score)/float64(len(words)*len(words)))
	}
}

func calculateHints(target, word string) Hints {
	var fixed []Hint
	for i := range target {
		if word[i] == target[i] {
			fixed = append(fixed, Hint{
				Letter: word[i],
				Index:  byte(i),
			})
		}
	}

	var moving []Hint
	for i := range word {
		exists := false
		for _, p := range fixed {
			if p.Letter == word[i] {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		for j := range target {
			if target[j] == word[i] {
				exists = true
				break
			}
		}
		if exists {
			moving = append(moving, Hint{
				Letter: word[i],
				Index:  byte(i),
			})
		}
	}

	var bad []byte
	for i := range word {
		exists := false
		for j := range bad {
			if bad[j] == word[i] {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		for j := range target {
			if target[j] == word[i] {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		bad = append(bad, word[i])
	}

	return Hints{
		Fixed:  fixed,
		Moving: moving,
		Bad:    string(bad),
	}
}

func matchesHints(word string, hints Hints) bool {
	for _, p := range hints.Fixed {
		if word[p.Index] != p.Letter {
			return false
		}
	}

	for _, p := range hints.Moving {
		if word[p.Index] == p.Letter {
			return false
		}
		found := false
		for j := 0; j < len(word); j++ {
			if word[j] == p.Letter {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for i := 0; i < len(hints.Bad); i++ {
		for j := 0; j < len(word); j++ {
			if word[j] == hints.Bad[i] {
				return false
			}
		}
	}

	return true
}

func loadDict() []string {
	return small
}

func readLines(filename string) []string {
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func writeLines(filename string, lines []string) {
	file, err := os.Create(filename)
	check(err)
	defer file.Close()

	for _, line := range lines {
		file.WriteString(line)
		file.WriteString("\n")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
