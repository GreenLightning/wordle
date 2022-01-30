package main

type Hints struct {
	Fixed  []Hint
	Moving []Hint
	Bad    string
}

type Hint struct {
	Letter byte
	Index  byte
}

// Calculates a reprentation that can be used as a map key.
func (hints *Hints) key() string {
	key := make([]byte, 0, 2*len(hints.Fixed)+1+2*len(hints.Moving)+1+len(hints.Bad))
	for _, h := range hints.Fixed {
		key = append(key, h.Letter)
		key = append(key, h.Index)
	}
	key = append(key, '+')
	for _, h := range hints.Moving {
		key = append(key, h.Letter)
		key = append(key, h.Index)
	}
	key = append(key, '-')
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
		if h.Index != 255 && word[h.Index] == h.Letter {
			return false
		}
		found := false
		for j := 0; j < len(word); j++ {
			if word[j] == h.Letter {
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
			if word[j] == hints.Bad[i] {
				return false
			}
		}
	}

	return true
}
