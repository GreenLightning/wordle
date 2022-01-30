package main

const (
	ABSENT  = 0
	PRESENT = 1
	CORRECT = 2
)

// Evaluation function extracted from Wordle JavaScript.
// Not used right now, for reference only.
func evaluate(target, guess string) []int {
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
