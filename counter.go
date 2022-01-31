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

	list := small
	for i := 0; i < len(hints.Fixed); i++ {
		sub := smallWordsByFixedHint[hints.Fixed[i]]
		if len(sub) < len(list) {
			list = sub
		}
	}

	var count int64
	for i := 0; i < len(list); i++ {
		if matchesHints(list[i], hints) {
			count++
		}
	}

	counter.cache[key] = count
	return count
}
