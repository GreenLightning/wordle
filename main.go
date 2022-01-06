package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
var dictFlag = flag.String("dict", "huge.txt", "dictionary to use")

func main() {
	flag.Parse()

	if *filterFlag {
		filter()
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
	words := readLines(filepath.Join(filteredDir, *dictFlag))
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
