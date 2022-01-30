package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

type Record struct {
	Index int
	Word  string
	Score int64
}

func findBest() {
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
		fmt.Printf("\r%d/%d", i+1, len(big))
	}

	length := len(fmt.Sprint(len(big)))
	fmt.Printf("\r%s\r", strings.Repeat(" ", 2*length+1))

	sort.SliceStable(records, func(i, j int) bool {
		return records[i].Score < records[j].Score
	})

	for i := 0; i < len(records) && i < 20; i++ {
		record := records[i]
		fmt.Printf("%s %.3f%%\n", record.Word, float64(100*record.Score)/float64(len(small)*len(small)))
	}
}
