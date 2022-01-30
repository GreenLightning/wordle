package main

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
