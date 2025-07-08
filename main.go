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
	"github.com/joho/godotenv"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type WordleGame struct {
	PossibleWords []string
	LastGuess     string
	IsActive      bool
	Mode 		  string
	Attempts 	  int
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
	_ = godotenv.Load()
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
	log.Println("handleStart called")
	chatID := update.Message.Chat.ID
	ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		fmt.Sprint("Привет, я твой помощник в решении ежедевных Wordle от New York Times, и не только.\n"+
		"Я буду давать тебе новые слова, анализируя твой фидбэк от прошлого слова.\n"+
		"Чтобы начать решение Wordle, используй команду /solve.\n"+
		"Также ты можешь попросить у меня подсказку, если при самостоятельном решении Wordle где-то застрял - используй команду /help."),
	))
	return nil
}

func handleHelp(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	gamesMu.Lock()
	userGames[chatID] = &WordleGame{
		PossibleWords: wordlist,
		IsActive:      true,
		Mode:		   "HELP",
		Attempts: 	   0,	
	}
	gamesMu.Unlock()
	ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		fmt.Sprint("Тебе нужна подсказка? - Отлично. Отправь мне все известные слова и их статусы через пробел: `TRAIN` `bygbb` (каждая пара в отдельной строчке):\n"+
		"🟩 (G) — буква на месте\n"+
		"🟨 (Y) — буква есть, но не тут\n"+
		"⬛️ (B) — буквы нет в слове\n\n"),
	))
	return nil
}

func handleSolve(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	gamesMu.Lock()
	userGames[chatID] = &WordleGame{
		PossibleWords: wordlist,
		IsActive:      true,
		Mode:		   "SOLVE",
		Attempts: 	   1,	
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
			"Если слово угадано, напиши `Guess`.\n"+
			"Если слово не подходит, напиши `Notfound`.\n"+
			"При проигрыше напиши `Lose`.",
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
			"Игра не активна. Используй /solve или /help для старта.",
		))
		return nil
	}

	switch game.Mode {
	case "SOLVE":
		return handleSolveFeedBack(ctx, update, game)
	case "HELP":
		return handleHelpFeedBack(ctx, update, game)
	case "CHILL": 
		{
		ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			fmt.Sprint("На данный момент я в состоянии отдыха, потому что не выполняю никаких задач.\n"+
			"Попробуй использовать команды start, solve или help."),
		))
		}
	}

	return nil
}

func handleHelpFeedBack(ctx *th.Context, update telego.Update, game *WordleGame) error {
	chatID := update.Message.Chat.ID
	input := strings.TrimSpace(strings.ToUpper(update.Message.Text))

	handled := handleSpecialFeedback(ctx, game, chatID, input)
	if handled {
		return nil
	}

	lines := strings.Split(input, "\n")
	
	gamesMu.Lock()
	game.Attempts += len(lines)
	gamesMu.Unlock()

	if len(lines) == 0 {
		ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Пожалуйста, отправь слова и их статусы в формате `TRAIN-bygbb`, по одному на строку.",
		))
		return nil
	}
	var validInputs [][]string
	for _,line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) != 2 || len(parts[0]) != 5 || len(parts[1]) != 5 {
			ctx.Bot().SendMessage(ctx, tu.Message(
                tu.ID(chatID),
                fmt.Sprintf("Неверный формат строки: `%s`. Используй формат `TRAIN bygbb`.", line),
            ))
            return nil
		}
		
		word := parts[0]
		feedback := parts[1]

		if !isValidWord(word) || !isValidFeedBack(feedback) {
			ctx.Bot().SendMessage(ctx, tu.Message(
				tu.ID(chatID),
				fmt.Sprintf("Неверное слово или фидбэк: `%s-%s`. Слово должно быть 5 букв, фидбэк — gybbg.", word, feedback),
			))
			return nil
		}
		validInputs = append(validInputs, []string{word, feedback})
	}
	
	filtered := game.PossibleWords
	for _, input := range validInputs {
		word, feedback := input[0], input[1]
		filtered = filterWords(filtered, word, feedback)
	}

	if len(filtered) == 0 {
		ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: нет подходящих слов на основе твоего ввода. Проверь данные или начни заново с /help.",
		))
		gamesMu.Lock()
		game.IsActive = false
		game.Mode = "CHILL"
		game.Attempts = 0
		delete(userGames, chatID)
		gamesMu.Unlock()
		return nil
	}

	Guess := chooseNext(filtered)

	gamesMu.Lock()
	game.PossibleWords = filtered
	game.LastGuess = Guess
	gamesMu.Unlock()

	gamesMu.RLock()
	attempt := game.Attempts
	gamesMu.RUnlock()

	ctx.Bot().SendMessage(ctx, tu.Message(
        tu.ID(chatID),
        fmt.Sprintf("Моя подсказка: **%s**(количество оставшихся попыток для победы: %d)\n\n"+
		"Отправь новый фидбэк в формате `TRAIN bygbb` или guess, если я угадал.", Guess, 5-attempt),
    ))

	return nil
}

func handleSolveFeedBack(ctx *th.Context, update telego.Update, game *WordleGame) error {
	chatID := update.Message.Chat.ID

	feedback := strings.TrimSpace(strings.ToUpper(update.Message.Text))

	handled := handleSpecialFeedback(ctx, game, chatID, feedback)
	if handled {
		return nil
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
		game.Mode = "CHILL"
		game.Attempts = 0
		gamesMu.Unlock()
		return
	}


	nextGuess := chooseNext(filtered)

	gamesMu.Lock()
	game.PossibleWords = filtered
	game.LastGuess = nextGuess
	game.Attempts += 1
	gamesMu.Unlock()
	gamesMu.RLock()
	mode := game.Mode
	attempt := game.Attempts
	gamesMu.RUnlock()

	switch mode {
	case "SOLVE":
		{
		ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		fmt.Sprintf("Моя %d-ая догадка: **%s**", attempt, nextGuess),
	))
	}
	case "HELP":
	{
		ctx.Bot().SendMessage(ctx, tu.Message(
        tu.ID(chatID),
        fmt.Sprintf("Моя подсказка: **%s**(количество оставшихся попыток для победы: %d)\n\nОтправь новый фидбэк в формате `TRAIN bygbb` или guess, если я угадал.", nextGuess, 5-attempt),
    ))
	}
	}
	
}

func handleSpecialFeedback(ctx *th.Context, game* WordleGame, chatID int64, feedback string) bool {
	switch feedback {
	case "NOTFOUND":
		{
		gamesMu.Lock()
		game.PossibleWords = filteredOut(game.PossibleWords, game.LastGuess)
		game.Attempts -= 1
		gamesMu.Unlock()
		giveNextGuess(game.PossibleWords, chatID, game, ctx)
		return true
		}
	case "LOSE":
		{
		ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		"Эх, проигрыш. Попробуй еще раз, я покажу на что способен!",))
		gamesMu.Lock()
		game.IsActive = false
		game.Mode = "CHILL"
		game.Attempts = 0
		delete(userGames, chatID)
		gamesMu.Unlock()
		return true
		}
	case "GUESS":
		{
		ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Я рад, что смог тебе помочь решить wordle!",
		))
		gamesMu.Lock()
		game.IsActive = false
		game.Mode = "CHILL"
		game.Attempts = 0
		delete(userGames, chatID)	
		gamesMu.Unlock()
		return true
		}
	}
	return false
}

func getOptimalFirstWord() string {
	return optimalFirstWords[rand.Intn(len(optimalFirstWords))]
}

func isValidWord(word string) bool {
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

func filteredOut(words []string, exclude string) []string {
	filtered := []string{}
	for _,w := range words {
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
