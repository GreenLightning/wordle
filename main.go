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

const (
	ABSENT  = 0
	PRESENT = 1
	CORRECT = 2
)

func assess(target, guess string) []int {
	results := make([]int, len(target))
	used := make([]bool, len(target))
	for i := range results {
		if guess[i] == target[i] {
			results[i] = CORRECT
			used[i] = true
		}
	}
	for i, current := range results {
		if current == CORRECT {
			continue
		}
		for j := range target {
			if !used[j] && guess[i] == target[j] {
				results[i] = PRESENT
				used[j] = true
				break
			}
		}
	}
	return results
}

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

func calculateHints(target, word string) (hints Hints) {
	correct := make([]bool, len(target))
	used := make([]bool, len(target))

	for i := range target {
		if word[i] == target[i] {
			correct[i] = true
			used[i] = true
			hints.Fixed = append(hints.Fixed, Hint{
				Letter: word[i],
				Index:  byte(i),
			})
		}
	}

	for i := range target {
		if correct[i] {
			continue
		}
		for j := range target {
			if !used[j] && word[i] == target[j] {
				used[j] = true
				hints.Moving = append(hints.Moving, Hint{
					Letter: word[i],
					Index:  byte(i),
				})
				break
			}
		}
	}

	var bad []byte
	for i := range word {
		exists := false
		for _, h := range hints.Fixed {
			if word[i] == h.Letter {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		for _, h := range hints.Moving {
			if word[i] == h.Letter {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		bad = append(bad, word[i])
	}
	hints.Bad = string(bad)
	return
}

func matchesHints(word string, hints Hints) bool {
	for _, h := range hints.Fixed {
		if word[h.Index] != h.Letter {
			return false
		}
	}

	for _, h := range hints.Moving {
		if word[h.Index] == h.Letter {
			return false
		}
		found := false
		for j := 0; j < len(word); j++ {
			if word[j] == h.Letter {
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

type Counter struct {
	cache map[string]int64
}

func MakeCounter() Counter {
	return Counter{
		cache: make(map[string]int64),
	}
}

func (counter *Counter) CountMatches(hints Hints) int64 {
	key := hints.key()
	if count, ok := counter.cache[key]; ok {
		return count
	}
	var count int64
	for _, word := range small {
		if matchesHints(word, hints) {
			count++
		}
	}
	counter.cache[key] = count
	return count
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
	words := small
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
	dist := make(map[int64]int64)

	counter := MakeCounter()
	for _, target := range small {
		hints := calculateHints(target, word)
		count := counter.CountMatches(hints)
		dist[count]++
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
	inputs := make(chan Record, 64)
	outputs := make(chan Record, 64)

	go func() {
		for i, word := range big {
			inputs <- Record{
				Index: i,
				Word:  word,
			}
		}
		close(inputs)
	}()

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			counter := MakeCounter()
			for record := range inputs {
				for _, target := range small {
					hints := calculateHints(target, record.Word)
					count := counter.CountMatches(hints)
					record.Score += count
				}
				outputs <- record
			}
		}()
	}

	records := make([]Record, len(big))
	for i := range records {
		record := <-outputs
		records[record.Index] = record
		fmt.Println(i+1, "/", len(big))
	}

	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Score < records[j].Score
	})

	for i := 0; i < len(records) && i < 20; i++ {
		record := records[i]
		fmt.Printf("%s %.3f%%\n", record.Word, float64(100*record.Score)/float64(len(small)*len(small)))
	}
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
