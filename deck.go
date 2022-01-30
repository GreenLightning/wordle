package main

import (
	"fmt"
	"strings"
)

func generateDeck() {
	type Card struct {
		Front string
		Back  string
	}

	var cards []Card

	// Generate cards with 2, 3 or 4 fixed letters on the front
	// and a list of matching words on the back.

	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			buckets := make(map[string][]string)
			for _, word := range small {
				key := word[i:i+1] + word[j:j+1]
				buckets[key] = append(buckets[key], word)
			}
			for key, words := range buckets {
				if len(words) > 5 {
					continue
				}
				front := []byte("_____")
				front[i] = key[0]
				front[j] = key[1]
				cards = append(cards, Card{string(front), strings.Join(words, ", ")})
			}
		}
	}

	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			for k := j + 1; k < 5; k++ {
				buckets := make(map[string][]string)
				for _, word := range small {
					key := word[i:i+1] + word[j:j+1] + word[k:k+1]
					buckets[key] = append(buckets[key], word)
				}
				for key, words := range buckets {
					if len(words) > 5 {
						continue
					}
					front := []byte("_____")
					front[i] = key[0]
					front[j] = key[1]
					front[k] = key[2]
					cards = append(cards, Card{string(front), strings.Join(words, ", ")})
				}
			}
		}
	}

	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			for k := j + 1; k < 5; k++ {
				for l := k + 1; l < 5; l++ {
					buckets := make(map[string][]string)
					for _, word := range small {
						key := word[i:i+1] + word[j:j+1] + word[k:k+1] + word[l:l+1]
						buckets[key] = append(buckets[key], word)
					}
					for key, words := range buckets {
						front := []byte("_____")
						front[i] = key[0]
						front[j] = key[1]
						front[k] = key[2]
						front[l] = key[3]
						cards = append(cards, Card{string(front), strings.Join(words, ", ")})
					}
				}
			}
		}
	}

	// If multiple cards have the same word list, mark the later ones (with
	// more hints on the front) for removal.
	indicesByWords := make(map[string][]int)
	remove := make([]bool, len(cards))
	for i, card := range cards {
		indicesByWords[card.Back] = append(indicesByWords[card.Back], i)
	}
	for _, indices := range indicesByWords {
		for i := 1; i < len(indices); i++ {
			remove[indices[i]] = true
		}
	}

	var lines []string
	for i, card := range cards {
		if !remove[i] {
			lines = append(lines, fmt.Sprintf("%s\t%s", card.Front, card.Back))
		}
	}

	writeLines("wordle.txt", lines)
}
