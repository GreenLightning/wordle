package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

type Hints struct {
	Base string
	Good string
	Bad  string
}

const (
	sourceDir   = "dicts"
	filteredDir = "filtered"
)

var filterFlag = flag.Bool("filter", false, "filter source dictionaries")
var dictFlag = flag.String("dict", "small.txt", "dictionary to use")
var distFlag = flag.String("dist", "", "calculate distribution for one word")
var evalFlag = flag.Bool("eval", false, "find best starting word")

func main() {
	flag.Parse()

	if *filterFlag {
		filter()
	}

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

func filter() {
	regex := regexp.MustCompile(`^[A-Za-z]{5}$`)

	entries, err := os.ReadDir(sourceDir)
	check(err)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		words := readLines(filepath.Join(sourceDir, name))

		// Strip header.
		for i, word := range words {
			if word == "---" {
				words = words[i+1:]
				break
			}
		}

		// Find five-character words, convert to upper case and remove duplicates.
		filtered := make(map[string]bool)
		for _, word := range words {
			if regex.MatchString(word) {
				word = strings.ToUpper(word)
				filtered[word] = true
			}
		}

		words = words[:0]
		for word, _ := range filtered {
			words = append(words, word)
		}

		sort.Strings(words)

		fmt.Printf("%10s %6d\n", name, len(words))

		writeLines(filepath.Join(filteredDir, name), words)
	}
}

func lookup(args []string) {
	words := loadDict()
	argRegex := regexp.MustCompile(`^([_A-Za-z]{5})(?:\+([A-Za-z]+))?(?:-([A-Za-z]+))?$`)

	for i, arg := range args {
		matches := argRegex.FindStringSubmatch(arg)
		if matches == nil {
			fmt.Printf("invalid argument %q\n", arg)
			continue
		}

		hints := Hints{
			Base: strings.ToUpper(matches[1]),
			Good: strings.ToUpper(matches[2]),
			Bad:  strings.ToUpper(matches[3]),
		}

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
	cache := make(map[Hints]int64)
	dist := make(map[int64]int64)
	for _, target := range words {
		hints := calculateHints(target, word)
		if count, ok := cache[hints]; ok {
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
		cache[hints] = count
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
			cache := make(map[Hints]int64)
			for record := range inputs {
				for _, target := range words {
					hints := calculateHints(target, record.Word)
					if count, ok := cache[hints]; ok {
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
					cache[hints] = count
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
	base := []byte(word)
	for i := range base {
		if base[i] != target[i] {
			base[i] = '_'
		}
	}

	var good []byte
	for i := range word {
		exists := false
		for j := range good {
			if good[j] == word[i] {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		for j := range base {
			if base[j] == word[i] {
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
			good = append(good, word[i])
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
		Base: string(base),
		Good: string(good),
		Bad:  string(bad),
	}
}

func matchesHints(word string, hints Hints) bool {
	if len(word) != len(hints.Base) {
		return false
	}

	for i := 0; i < len(hints.Base); i++ {
		if hints.Base[i] != '_' && word[i] != hints.Base[i] {
			return false
		}
	}

	for i := 0; i < len(hints.Good); i++ {
		found := false
		for j := 0; j < len(word); j++ {
			if word[j] == hints.Good[i] {
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
	// Here we assume that the file is well-formed and each line
	// contains one word consisting of five uppercase letters.
	return readLines(filepath.Join(filteredDir, *dictFlag))
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
