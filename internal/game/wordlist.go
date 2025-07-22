package game

import (
	"bufio"
	"os"
	"errors"
	"log"
	"math/rand"
	"strings"
)

var (
	wordlist []string
	ErrEmptyWordList = errors.New("wordlist is empty")
	ErrGameNotFound  = errors.New("game not found")
)

func init() {
	wordlist = loadWordList("assets/wordle.txt")
}

func GetWordList() []string {
	return wordlist
}

func GetOptimalFirstWord() string {
	return optimalFirstWords[rand.Intn(len(optimalFirstWords))]
}

func IsValidWord(word string) bool {
	if len(word) != 5 {
		return false
	}
	for _, c := range word {
		if !('A' <= c && c <= 'Z') {
			return false
		}
	}
	return true
}

func IsValidFeedBack(feedback string) bool {
	if len(feedback) != 5 {
		return false
	}
	for _, c := range feedback {
		if c != 'G' && c != 'Y' && c != 'B' {
			return false
		}
	}
	return true
}

func filteredOut(words []string, exclude string) []string {
	filtered := []string{}
	for _, w := range words {
		if w != exclude {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

func filterWords(words []string, guess, feedback string) []string {
	var result []string
	for _, word := range words {
		valid := true
		letterCount := make(map[byte]int)
		for i := 0; i < 5; i++ {
			if feedback[i] == 'G' || feedback[i] == 'Y' {
				letterCount[guess[i]]++
			}
		}

		for i := 0; i < 5; i++ {
			g := guess[i]
			w := word[i]
			fb := feedback[i]

			switch fb {
			case 'G':
				if w != g {
					valid = false
				}
			case 'Y':
				if w == g || !strings.Contains(word, string(g)) {
					valid = false
				}
			case 'B':
				if strings.Count(word, string(g)) > letterCount[g] {
					valid = false
				}
			}
			if !valid {
				break
			}
		}
		if valid {
			result = append(result, word)
		}
	}
	return result
}

func scoreWords(words []string) map[string]int {
	freq := make(map[rune]int)
	for _, word := range words {
		seen := make(map[rune]bool)
		for _, c := range word {
			if !seen[c] {
				freq[c]++
				seen[c] = true
			}
		}
	}

	scores := make(map[string]int)
	for _, word := range words {
		score := 0
		seen := make(map[rune]bool)
		for _, c := range word {
			if !seen[c] {
				score += freq[c]
				seen[c] = true
			}
		}
		scores[word] = score
	}
	return scores
}

func ChooseNext(words []string) string {
	scores := scoreWords(words)
	best := ""
	maxScore := -1
	for word, score := range scores {
		if score > maxScore {
			best = word
			maxScore = score
		}
	}
	return best
}

func loadWordList(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Ошибка загрузки словаря: %v", err)
	}
	defer func(){
		if err := file.Close(); err != nil {
			log.Printf("Failed to close wordlist file %s: %v", filename, err)
		}
	}()
	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, strings.ToUpper(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading wordlist file %s: %v", filename, err)
	}

	return words
}