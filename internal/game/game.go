package game

import (
	"sync"
)

type WordleGame struct {
	PossibleWords []string
	LastGuess     string
	IsActive      bool
	Mode          string
	Attempts      int
}

var (
	gamesMu   sync.RWMutex
	userGames = make(map[int64]*WordleGame)
)

var optimalFirstWords = []string{
	"CRANE", "SLATE", "ADIEU", "AUDIO", "RAISE",
	"ROATE", "CRATE", "TRACE", "LEAST", "STARE",
}

func StartGame(chatID int64, mode string) error {
	gamesMu.Lock()
	defer gamesMu.Unlock()

	wordlist := GetWordList()
	if len(wordlist) == 0 {
		return ErrEmptyWordList
	}

	attempts := 0
	if mode == "SOLVE" {
		attempts = 1
	}

	userGames[chatID] = &WordleGame{
		PossibleWords: wordlist,
		IsActive: true,
		Mode: mode,
		Attempts: attempts,
	}
	return nil
}

func IncrementAttempts(chatID int64, count int) error {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	if game, exists := userGames[chatID]; exists {
		game.Attempts += count
		return nil
	}
	return ErrGameNotFound
}

func UpdateLastGuess(chatID int64, guess string) error {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	if game, exists := userGames[chatID]; exists {
		game.LastGuess = guess
		return nil
	}
	return ErrGameNotFound
}

func UpdateGameState(chatID int64, words []string, guess string) error {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	if game, exists := userGames[chatID]; exists {
		game.PossibleWords = words
		game.LastGuess = guess
		game.Attempts ++
		return nil
	}
	return ErrGameNotFound
}

func EndGame(chatID int64) error {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	if _, exists := userGames[chatID]; exists {
		delete(userGames, chatID)
		return nil
	}
	return ErrGameNotFound
}

func GetGame(chatID int64) (*WordleGame, bool) {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	game, exists := userGames[chatID]
	return game, exists
}

func GetPossibleWords(chatID int64) []string {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	if game, exists := userGames[chatID]; exists {
		return game.PossibleWords
	}
	return nil
}

func GetMode(chatID int64) (string, error) {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	if game, exists := userGames[chatID]; exists {
		return game.Mode, nil
	}
	return "", ErrGameNotFound
}

func GetAttempts(chatID int64) (int, error) {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	if game, exists := userGames[chatID]; exists {
		return game.Attempts, nil
	}
	return 0, ErrGameNotFound
}

func FilteredOutLastGuess(chatID int64) error {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	if game, exists := userGames[chatID]; exists {
		game.PossibleWords = filteredOut(game.PossibleWords, game.LastGuess)
		game.Attempts --
		return nil
	}
	return ErrGameNotFound
}

func FilterSingleWord(chatID int64, feedback string) ([]string, error) {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	if game, exists := userGames[chatID]; exists {
		return filterWords(game.PossibleWords, game.LastGuess ,feedback), nil
	}
	return nil, ErrGameNotFound
}

func FilterWords(chatID int64, inputs [][]string) ([]string, error) {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	if game, exists := userGames[chatID]; exists {
		filtered := game.PossibleWords
		for _, input := range inputs {
			word, feedback := input[0], input[1]
			filtered = filterWords(filtered, word, feedback)
		}
		return filtered, nil
	}
	return nil, ErrGameNotFound
}

