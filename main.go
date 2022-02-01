package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var distFlag = flag.String("dist", "", "calculate distribution for one word")
var bestFlag = flag.Bool("best", false, "find best starting word")
var best2Flag = flag.Bool("best2", false, "find best two starting words")
var forFlag = flag.String("for", "", "find best word for hints")

//go:generate go run words_organized_generator.go
func main() {
	flag.Parse()

	if *distFlag != "" {
		err := calculateDistribution(*distFlag)
		if err != nil {
			fmt.Println(err)
		}
	}

	if *bestFlag {
		findBest()
	}

	if *best2Flag {
		findBest2()
	}

	if *forFlag != "" {
		hints, err := parseHints(*forFlag)
		if err != nil {
			fmt.Println(err)
		} else {
			findBestFor(hints)
		}
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

func lookup(arg string) error {
	hints, err := parseHints(arg)
	if err != nil {
		return err
	}

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
