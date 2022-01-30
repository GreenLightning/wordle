package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var deckFlag = flag.Bool("deck", false, "generate Anki deck for learning words")
var distFlag = flag.String("dist", "", "calculate distribution for one word")
var bestFlag = flag.Bool("best", false, "find best starting word")

func main() {
	flag.Parse()

	if *deckFlag {
		generateDeck()
	}

	if *distFlag != "" {
		calculateDistribution(*distFlag)
	}

	if *bestFlag {
		findBest()
	}

	for i, arg := range flag.Args() {
		err := lookup(arg)
		if err != nil {
			fmt.Println(err)
		}

		if i+1 < flag.NArg() {
			fmt.Println(strings.Repeat("=", 20))
		}
	}
}

var argRegex = regexp.MustCompile(`^([_A-Za-z]{5})(?:\+((?:[A-Za-z][1-5]?)+))?(?:-([A-Za-z]+))?$`)

func lookup(arg string) error {
	matches := argRegex.FindStringSubmatch(arg)
	if matches == nil {
		return fmt.Errorf("invalid argument %q\n", arg)
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
	for i := 0; i < len(moving); i++ {
		hint := Hint{Letter: moving[i]}
		if i+1 < len(moving) && moving[i+1] >= '0' && moving[i+1] <= '9' {
			hint.Index = moving[i+1] - '1'
			i++
		} else {
			hint.Index = 255
		}
		hints.Moving = append(hints.Moving, hint)
	}
	hints.Bad = strings.ToUpper(matches[3])

	for _, word := range small {
		if matchesHints(word, hints) {
			fmt.Println(word)
		}
	}

	return nil
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
