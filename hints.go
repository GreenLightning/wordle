package main

import (
	"fmt"
	"regexp"
	"strings"
)

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

func (hints Hints) String() string {
	var s []byte
	s = append(s, '!')
	for _, h := range hints.Fixed {
		s = append(s, h.Letter)
		s = append(s, '1'+h.Index)
	}
	s = append(s, '?')
	for _, h := range hints.Moving {
		s = append(s, h.Letter)
		s = append(s, '1'+h.Index)
	}
	s = append(s, '+')
	s = append(s, hints.Required...)
	s = append(s, '-')
	s = append(s, hints.Bad...)
	return string(s)
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

var hintRegex = regexp.MustCompile(`^([_A-Za-z]{5})(?:\+((?:[A-Za-z][1-5]*)+))?(?:-([A-Za-z]+))?$`)

func parseHints(text string) (hints Hints, err error) {
	matches := hintRegex.FindStringSubmatch(text)
	if matches == nil {
		err = fmt.Errorf("invalid argument %q\n", text)
		return
	}

	fixed := strings.ToUpper(matches[1])
	for i := range fixed {
		if fixed[i] != '_' {
			hints.Fixed = append(hints.Fixed, Hint{
				Letter: fixed[i],
				Index:  byte(i),
			})
		}
	}

	good := strings.ToUpper(matches[2])
	for i := 0; i < len(good); i++ {
		hints.Required += good[i : i+1]
		for letter := good[i]; i+1 < len(good) && good[i+1] >= '0' && good[i+1] <= '9'; i++ {
			hints.Moving = append(hints.Moving, Hint{
				Letter: letter,
				Index:  good[i+1] - '1',
			})
		}
	}

	bad := strings.ToUpper(matches[3])
	for i := 0; i < len(bad); i++ {
		if !strings.Contains(hints.Bad, bad[i:i+1]) {
			hints.Bad += bad[i : i+1]
		}
	}

	return
}

func calculateHints(target, word string) (hints Hints) {
	var (
		good [5]bool // refers to word
		used [5]bool // refers to target
	)

	for i := 0; i < 5; i++ {
		if word[i] == target[i] {
			good[i] = true
			used[i] = true
			hints.Fixed = append(hints.Fixed, Hint{
				Letter: word[i],
				Index:  byte(i),
			})
		}
	}

	var required [5]byte
	var requiredCount int
	for i := 0; i < 5; i++ {
		if good[i] {
			continue
		}
		for j := 0; j < 5; j++ {
			if !used[j] && word[i] == target[j] {
				good[i] = true
				used[j] = true
				required[requiredCount] = word[i]
				requiredCount++
				hints.Moving = append(hints.Moving, Hint{
					Letter: word[i],
					Index:  byte(i),
				})
				break
			}
		}
	}

	var bad [5]byte
	var badCount int
badLoop:
	for i := 0; i < 5; i++ {
		if good[i] {
			continue
		}
		char := word[i]
		for j := 0; j < badCount; j++ {
			if bad[j] == char {
				continue badLoop
			}
		}
		bad[badCount] = char
		badCount++
	}

	hints.Required = string(required[:requiredCount])
	hints.Bad = string(bad[:badCount])
	return
}

func matchesHints(word string, hints Hints) bool {
	for i := 0; i < len(hints.Fixed); i++ {
		if word[hints.Fixed[i].Index] != hints.Fixed[i].Letter {
			return false
		}
	}

	for i := 0; i < len(hints.Moving); i++ {
		if word[hints.Moving[i].Index] == hints.Moving[i].Letter {
			return false
		}
	}

	var used [5]bool
	for i := 0; i < len(hints.Fixed); i++ {
		used[hints.Fixed[i].Index] = true
	}

requiredLoop:
	for i := 0; i < len(hints.Required); i++ {
		for j := 0; j < 5; j++ {
			if !used[j] && word[j] == hints.Required[i] {
				used[j] = true
				continue requiredLoop
			}
		}
		return false
	}

	for i := 0; i < len(hints.Bad); i++ {
		for j := 0; j < 5; j++ {
			if !used[j] && word[j] == hints.Bad[i] {
				return false
			}
		}
	}

	return true
}

func mergeHints(a Hints, b Hints) (r Hints) {
	var fixed [5]Hint
	for i := 0; i < len(a.Fixed); i++ {
		fixed[a.Fixed[i].Index] = a.Fixed[i]
	}
	for i := 0; i < len(b.Fixed); i++ {
		fixed[b.Fixed[i].Index] = b.Fixed[i]
	}

	var fixedCount int
	for i := 0; i < 5; i++ {
		if fixed[i].Letter != 0 {
			fixedCount++
		}
	}

	r.Fixed = make([]Hint, 0, fixedCount)
	for i := 0; i < 5; i++ {
		if fixed[i].Letter != 0 {
			r.Fixed = append(r.Fixed, fixed[i])
		}
	}

	r.Moving = make([]Hint, len(a.Moving))
	copy(r.Moving, a.Moving)
movingLoop:
	for i := 0; i < len(b.Moving); i++ {
		for j := 0; j < len(a.Moving); j++ {
			if a.Moving[j] == b.Moving[i] {
				continue movingLoop
			}
		}
		r.Moving = append(r.Moving, b.Moving[i])
	}

	{
		moving := r.Moving
		less := func(i, j int) bool {
			if moving[i].Index != moving[j].Index {
				return moving[i].Index < moving[j].Index
			}
			return moving[i].Letter < moving[j].Letter
		}
		for i := 0; i < len(moving); i++ {
			best := i
			for j := i + 1; j < len(moving); j++ {
				if less(j, best) {
					best = j
				}
			}
			if best != i {
				moving[i], moving[best] = moving[best], moving[i]
			}
		}
		r.Moving = moving
	}

	var letters ['Z'+1]bool
	for i := 0; i < len(a.Required); i++ {
		letters[a.Required[i]] = true
	}
	for i := 0; i < len(b.Required); i++ {
		letters[b.Required[i]] = true
	}

	var required []byte
	for letter := byte('A'); letter <= byte('Z'); letter++ {
		if !letters[letter] {
			continue
		}

		var aCount int
		for i := 0; i < len(a.Fixed); i++ {
			if a.Fixed[i].Letter == letter {
				aCount++
			}
		}
		for i := 0; i < len(a.Required); i++ {
			if a.Required[i] == letter {
				aCount++
			}
		}

		var bCount int
		for i := 0; i < len(b.Fixed); i++ {
			if b.Fixed[i].Letter == letter {
				bCount++
			}
		}
		for i := 0; i < len(b.Required); i++ {
			if b.Required[i] == letter {
				bCount++
			}
		}

		var count = aCount
		if bCount > aCount {
			count = bCount
		}

		for i := 0; i < len(r.Fixed); i++ {
			if r.Fixed[i].Letter == letter {
				count--
			}
		}

		for i := 0; i < count; i++ {
			required = append(required, letter)
		}
	}

	bad := []byte(a.Bad)
badLoop:
	for i := 0; i < len(b.Bad); i++ {
		for j := 0; j < len(a.Bad); j++ {
			if a.Bad[j] == b.Bad[i] {
				continue badLoop
			}
		}
		bad = append(bad, b.Bad[i])
	}

	{
		for i := 0; i < len(bad); i++ {
			best := i
			for j := i + 1; j < len(bad); j++ {
				if bad[j] < bad[best] {
					best = j
				}
			}
			if best != i {
				bad[i], bad[best] = bad[best], bad[i]
			}
		}
	}

	r.Required = string(required)
	r.Bad = string(bad)
	return
}
