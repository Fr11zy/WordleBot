package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type WordleGame struct {
	PossibleWords []string
	LastGuess     string
	IsActive      bool
}

var (
	gamesMu   sync.RWMutex
	userGames = make(map[int64]*WordleGame)
	wordlist  = loadWordList("wordle.txt")
)

var optimalFirstWords = []string{
	"CRANE", "SLATE", "ADIEU", "AUDIO", "RAISE",
	"ROATE", "CRATE", "TRACE", "LEAST", "STARE",
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	os.Setenv("TOKEN", "7549155657:AAHw170x4VCtmlamgso7h-YoW6tDQt6GW_Q")
	botToken := os.Getenv("TOKEN")

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)
	bh, _ := th.NewBotHandler(bot, updates)
	defer func() { _ = bh.Stop() }()

	bh.Handle(handleSolve, th.CommandEqual("solve"))
	bh.Handle(handleStart, th.CommandEqual("start"))
	bh.Handle(handleHelp, th.CommandEqual("help"))
	bh.Handle(handleFeedBack)
	_ = bh.Start()
}

func handleStart(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		"Привет, я твой помощник в решении ежедевных Wordle от New York Times, и не только.\n"+
		"Я буду давать тебе новые слова, анализируя твой фидбэк от прошлого слова.\n"+
		"Чтобы начать решение Wordle, используй команду /solve.",
	))
	return nil
}

func handleHelp(ctx *th.Context, update telego.Update) error {
	return nil
}

func handleSolve(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	gamesMu.Lock()
	userGames[chatID] = &WordleGame{
		PossibleWords: wordlist,
		IsActive:      true,
	}
	gamesMu.Unlock()

	firstGuess := getOptimalFirstWord()

	gamesMu.Lock()
	userGames[chatID].LastGuess = firstGuess
	gamesMu.Unlock()

	ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		fmt.Sprintf("Начинаем Wordle! Мой первый вариант: **%s**\n\n"+
			"Отправляй мне фидбэк по моим вариантам в формате `GYBBG`:\n"+
			"🟩 (G) — буква на месте\n"+
			"🟨 (Y) — буква есть, но не тут\n"+
			"⬛️ (B) — буквы нет в слове\n\n"+
			"Если слово угадано, напиши `Guess`.",
			firstGuess),
	))
	return nil
}

func handleFeedBack(ctx *th.Context, update telego.Update) error {
	if update.Message.Text == "" || update.Message.Text[0] == '/' {
		return nil
	}

	chatID := update.Message.Chat.ID
	gamesMu.RLock()
	game, exists := userGames[chatID]
	gamesMu.RUnlock()

	if !exists || !game.IsActive {
		ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Игра не активна. Используй /solve для старта.",
		))
		return nil
	}

	feedback := strings.ToUpper(update.Message.Text)

	switch feedback {
	case "NOTFOUND":
		{
		filtered := []string{}
		for _,w := range game.PossibleWords {
			if w != game.LastGuess {
				filtered = append(filtered, w)
			}
		}
		giveNextGuess(filtered, chatID, game, ctx)
		return nil
		}
	case "LOSE":
		{
		ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		"Эх, проигрыш. Начни заново /solve, я покажу на что способен!",))
		gamesMu.Lock()
		game.IsActive = false
		gamesMu.Unlock()
		return nil
		}
	case "GUESS":
		{
		ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		"🎉 Ура! Я молодец. Используй /solve для новой игры.",))
		gamesMu.Lock()
		game.IsActive = false
		gamesMu.Unlock()
		return nil
		}
	}

	if !isValidFeedBack(feedback) {
		ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Неверный формат. Используй GYBBG (например, GYBBG) или Guess.",
		))
		return nil
	}

	filtered := filterWords(game.PossibleWords, game.LastGuess, feedback)
	giveNextGuess(filtered, chatID, game, ctx)
	return nil
}

func giveNextGuess(filtered []string, chatID int64, game *WordleGame, ctx *th.Context) {
	if len(filtered) == 0 {
		ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: нет подходящих слов. Начни заново /solve.",
		))
		gamesMu.Lock()
		game.IsActive = false
		gamesMu.Unlock()
		return
	}


	nextGuess := chooseNext(filtered)

	gamesMu.Lock()
	game.PossibleWords = filtered
	game.LastGuess = nextGuess
	gamesMu.Unlock()

	ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		fmt.Sprintf("Моя следующая догадка: **%s**", nextGuess),
	))
}

func getOptimalFirstWord() string {
	return optimalFirstWords[rand.Intn(len(optimalFirstWords))]
}

func isValidFeedBack(feedback string) bool {
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

func chooseNext(words []string) string {
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
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words = append(words, strings.ToUpper(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return words
}
