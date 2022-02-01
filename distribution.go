package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var wordRegex = regexp.MustCompile(`^[A-Za-z]{5}$`)

func parseWord(text string) (string, error) {
	if !wordRegex.MatchString(text) {
		return "", fmt.Errorf("invalid argument %q\n", text)
	}
	return strings.ToUpper(text), nil
}

func calculateDistribution(input string) error {
	word, err := parseWord(input)
	if err != nil {
		return err
	}

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

	return nil
}
