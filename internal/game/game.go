package game

import (
	"sync"
	"math/rand"
)

type WordleGame struct {
	PossibleWords []string
	LastGuess     string
	IsActive      bool
	Mode          string
	Attempts      int
}

type PlayWordleGame struct {
	HiddenWord 		string
	LettersFlags    []string
	WordGame 		*WordleGame
}

var (
	gamesMu   sync.RWMutex
	userGames = make(map[int64]*WordleGame)
	userPlayGames = make(map[int64]*PlayWordleGame)
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
	var letters []string
	for char := 'A'; char <= 'Z'; char++ {
    letters = append(letters, string(char))
}
	if mode == "PLAY" {
		hidden := wordlist[rand.Intn(14849)]
		userPlayGames[chatID] = &PlayWordleGame{
			HiddenWord : hidden,
			LettersFlags: letters,
			WordGame: &WordleGame{
				PossibleWords: wordlist,
				IsActive: true,
				Mode: mode,
				Attempts: attempts,
			},
		}
		return nil
	}
	userGames[chatID] = &WordleGame{
		PossibleWords: wordlist,
		IsActive: true,
		Mode: mode,
		Attempts: attempts,
	}
	return nil
}

func (pg *PlayWordleGame) GetHiddenWord() string {
	return pg.HiddenWord
}

func (pg *PlayWordleGame) GetLettersFlags() []string {
	return pg.LettersFlags
}

func (wg *WordleGame) IncrementAttempts(count int) {
	wg.Attempts += count
}

func (wg *WordleGame) UpdateLastGuess(guess string) {
	wg.LastGuess = guess
}

func (wg *WordleGame) UpdateGameState(words []string, guess string) {
	wg.PossibleWords = words
	wg.LastGuess = guess
	if wg.Mode != "HELP" {
		wg.Attempts++
	}
}

func EndGame(chatID int64) error {
	gamesMu.Lock()
	defer gamesMu.Unlock()
	if _, exists := userGames[chatID]; exists {
		delete(userGames, chatID)
		return nil
	}
	if _, exists := userPlayGames[chatID]; exists {
		delete(userPlayGames, chatID)
		return nil
	}
	return ErrGameNotFound
}

func GetWGame(chatID int64) (*WordleGame, bool) {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	game, exists := userGames[chatID]
	return game, exists
}

func GetPGame(chatID int64) (*PlayWordleGame, bool) {
	gamesMu.RLock()
	defer gamesMu.RUnlock()
	game, exists := userPlayGames[chatID]
	return game, exists
  }

func (wg *WordleGame)GetPossibleWords() []string {
	return wg.PossibleWords
}

func (wg *WordleGame) GetMode() string {
	return wg.Mode
}

func (wg *WordleGame) GetAttempts() int {
	return wg.Attempts
}

func (wg *WordleGame) GetState() bool {
	return wg.IsActive
}

func (wg *WordleGame) FilteredOutLastGuess() {
	wg.PossibleWords = filteredOut(wg.PossibleWords, wg.LastGuess)
	wg.Attempts--
}

func (wg *WordleGame) FilterSingleWord(feedback string) []string {
	return filterWords(wg.PossibleWords, wg.LastGuess ,feedback)
}

func (wg *WordleGame) FilterWords(inputs [][]string) []string {
	filtered := wg.PossibleWords
	for _, input := range inputs {
		word, feedback := input[0], input[1]
		filtered = filterWords(filtered, word, feedback)
	}
	return filtered
}

