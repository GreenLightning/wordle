package main

// Hints is a different representation of the hints from one or more guesses.
//
// A green letter generates a fixed hint.
// A yellow letter generates a moving hint and a required letter.
// A gray letter generates a bad letter.
//
// Consider the following example:
// First  guess is TOAST, the A is yellow.
// Second guess is PASTA, the first A is yellow, the second A is gray.
//
// This will result in Moving = [A3, A2], Required = A and Bad = A.
// (Using 1-based indices like the command line for the example, the actual code
// uses 0-based indices.)
//
// This is because the target word cannot have an A at position 2 or 3
// (otherwise one of the As would have been marked green).
// However, from the PASTA guess, we know that the target word contains exactly
// one A (otherwise the second A would not have been marked gray).
//
// If a letter appears only in Required, the word must have at least that many
// instances of the letter. If it appears in both Required and Bad, then it must
// have exactly that many instances.
type Hints struct {
	// The word must have the given letter at the given position.
	Fixed []Hint

	// The word must not have the given letter at the given position.
	// (Otherwise the letter would have been marked green instead of yellow.)
	Moving []Hint

	// The word must contain these letters (ignoring fixed letters in the word).
	Required string

	// The word must not contain these letters (ignoring fixed and required letters in the word).
	Bad string
}

type Hint struct {
	Letter byte
	Index  byte
}

// Calculates a reprentation that can be used as a map key.
func (hints *Hints) key() string {
	key := make([]byte, 0, 2*len(hints.Fixed)+1+2*len(hints.Moving)+1+len(hints.Required)+1+len(hints.Bad))
	for _, h := range hints.Fixed {
		key = append(key, h.Letter)
		key = append(key, h.Index)
	}
	key = append(key, '|')
	for _, h := range hints.Moving {
		key = append(key, h.Letter)
		key = append(key, h.Index)
	}
	key = append(key, '|')
	key = append(key, hints.Required...)
	key = append(key, '|')
	key = append(key, hints.Bad...)
	return string(key)
}

func calculateHints(target, word string) (hints Hints) {
	correct := make([]bool, len(target))
	used := make([]bool, len(target))

	for i := range target {
		if word[i] == target[i] {
			correct[i] = true
			used[i] = true
			hints.Fixed = append(hints.Fixed, Hint{
				Letter: word[i],
				Index:  byte(i),
			})
		}
	}

	var required []byte
	for i := range target {
		if correct[i] {
			continue
		}
		for j := range target {
			if !used[j] && word[i] == target[j] {
				used[j] = true
				hints.Moving = append(hints.Moving, Hint{
					Letter: word[i],
					Index:  byte(i),
				})
				required = append(required, word[i])
				break
			}
		}
	}

	var bad []byte
	for i := range word {
		exists := false
		for _, h := range hints.Fixed {
			if word[i] == h.Letter {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		for _, h := range hints.Moving {
			if word[i] == h.Letter {
				exists = true
				break
			}
		}
		if exists {
			continue
		}
		bad = append(bad, word[i])
	}

	hints.Required = string(required)
	hints.Bad = string(bad)
	return
}

func matchesHints(word string, hints Hints) bool {
	for _, h := range hints.Fixed {
		if word[h.Index] != h.Letter {
			return false
		}
	}

	for _, h := range hints.Moving {
		if word[h.Index] == h.Letter {
			return false
		}
	}

	used := make([]bool, len(word))
	for _, h := range hints.Fixed {
		used[h.Index] = true
	}

	for i := 0; i < len(hints.Required); i++ {
		found := false
		for j := 0; j < len(word); j++ {
			if !used[j] && word[j] == hints.Required[i] {
				used[j] = true
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
			if !used[j] && word[j] == hints.Bad[i] {
				return false
			}
		}
	}

	return true
}
