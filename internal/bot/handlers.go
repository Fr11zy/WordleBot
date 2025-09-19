package bot

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/Fr11zy/WordleBot/internal/game"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func handleStart(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	_, err := ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		fmt.Sprint("Привет, я твой помощник в решении ежедевных Wordle от New York Times, и не только.\n"+
			"Я буду давать тебе новые слова, анализируя твой фидбэк от прошлого слова.\n"+
			"Чтобы начать решение Wordle, используй команду /solve.\n"+
			"Также ты можешь попросить у меня подсказку, если при самостоятельном решении Wordle где-то застрял - используй команду /help."),
	))
	return err
}

func handleHelp(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	if err := game.StartGame(chatID, "HELP"); err != nil {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: не удалось загрузить список слов. Попробуйте позже.",
		))
		return err
	}
	_, err := ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		fmt.Sprint("Тебе нужна подсказка? - Отлично. Отправь мне все известные слова и их статусы через пробел: `TRAIN` `bygbb` (каждая пара в отдельной строчке):\n"+
			"🟩 (G) — буква на месте\n"+
			"🟨 (Y) — буква есть, но не тут\n"+
			"⬛️ (B) — буквы нет в слове\n\n"),
	))
	return err
}

func handleSolve(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	if err := game.StartGame(chatID, "SOLVE"); err != nil {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: не удалось загрузить список слов. Попробуйте позже.",
		))
		return err
	}

	firstGuess := game.GetOptimalFirstWord()
	if wg,exists := game.GetWGame(chatID); exists {
		wg.UpdateLastGuess(firstGuess)
	}

	_, err := ctx.Bot().SendMessage(ctx, tu.Message(
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
	return err
}

func handlePLay(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	if err := game.StartGame(chatID, "PLAY"); err != nil {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: не удалось загрузить список слов. Попробуйте позже.",
		))
		return err
	}

	_, err := ctx.Bot().SendMessage(ctx,tu.Message(
		tu.ID(chatID),
		fmt.Sprint("Начинаем игру Wordle,я загадал одно слово из пяти букв, постарайся угадать его за 6 попыток.\n"+
		"Я буду отправлять тебе сообщения о статусе букв в твоем слове."),
	))
	return err
}

func handleFeedBack(ctx *th.Context, update telego.Update) error {
	if update.Message.Text == "" || update.Message.Text[0] == '/' {
		return nil
	}

	chatID := update.Message.Chat.ID
	input := strings.TrimSpace(strings.ToUpper(update.Message.Text))

	wg, exists := game.GetWGame(chatID)
	if !exists || !wg.GetState() {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Игра не активна. Используй /solve или /help для старта.",
		))
		return err
	}

	switch wg.GetMode() {
	case "SOLVE":
		return handleSolveFeedBack(ctx, chatID, input)
	case "HELP":
		return handleHelpFeedBack(ctx, chatID, input)
	case "PLAY":
		return handlePlayFeedBack(ctx, chatID, input)
	}
	return nil
}

func handlePlayFeedBack(ctx *th.Context, chatID int64, input string) error {
	guessWord := strings.ToUpper(input)
	pg, exists := game.GetPGame(chatID)
	if !exists {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Игра не активна. Используй /solve,/help для старта.",
		))
		return err
	}
	if !game.IsValidWord(guessWord) && !slices.Contains(pg.WordGame.GetPossibleWords(), guessWord){
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			fmt.Sprint("Неверный формат ввода или такого слова не существует.\n"+
			"Помни, что слово должно состоять из 5 букв от a до z."),
		))
		return err
	}
	
	return nil
}

func handleHelpFeedBack(ctx *th.Context, chatID int64, input string) error {
	handled, err := handleSpecialFeedback(ctx, chatID, input)
	if handled {
		return err
	}

	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Пожалуйста, отправь слова и их статусы в формате `TRAIN-bygbb`, по одному на строку.",
		))
		return err
	}
	wg, exists := game.GetWGame(chatID)
	if !exists {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: игра не найдена. Используй /solve,/help для старта.",
		))
		return err
	}
	wg.IncrementAttempts(len(lines))

	var validInputs [][]string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) != 2 || len(parts[0]) != 5 || len(parts[1]) != 5 {
			_, err := ctx.Bot().SendMessage(ctx, tu.Message(
				tu.ID(chatID),
				fmt.Sprintf("Неверный формат строки: `%s`. Используй формат `TRAIN bygbb`.", line),
			))
			return err
		}

		word := parts[0]
		feedback := parts[1]

		if !game.IsValidWord(word) || !game.IsValidFeedBack(feedback) {
			_, err := ctx.Bot().SendMessage(ctx, tu.Message(
				tu.ID(chatID),
				fmt.Sprintf("Неверное слово или фидбэк: `%s-%s`. Слово должно быть 5 букв от a до z, фидбэк — gybbg.", word, feedback),
			))
			return err
		}
		validInputs = append(validInputs, []string{word, feedback})
	}

	filtered := wg.FilterWords(validInputs)

	if len(filtered) == 0 {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: нет подходящих слов на основе твоего ввода. Проверь данные и начни заново с /help.",
		))
		if err != nil {
			return err
		}
		if err := game.EndGame(chatID); err != nil {
			return err
		}
		return nil
	}

	Guess := game.ChooseNext(filtered)

	wg.UpdateGameState(filtered, Guess)

	attempt := wg.GetAttempts()

	_, err = ctx.Bot().SendMessage(ctx, tu.Message(
		tu.ID(chatID),
		fmt.Sprintf("Моя подсказка: **%s**(количество оставшихся попыток для победы: %d)\n\n"+
			"Отправь новый фидбэк в формате `TRAIN bygbb` или guess, если я угадал.", Guess, 6-attempt),
	))
	
	return err
}

func handleSolveFeedBack(ctx *th.Context, chatID int64, feedback string) error {
	handled, err := handleSpecialFeedback(ctx, chatID, feedback)
	if handled {
		return err
	}

	if !game.IsValidFeedBack(feedback) {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Неверный формат. Используй GYBBG (например, GYBBG) или Guess.",
		))
		return err
	}

	wg, exists := game.GetWGame(chatID)
	if !exists {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: игра не найдена. Используй /solve,/help для старта.",
		))
		return err
	}

	filtered := wg.FilterSingleWord(feedback)
	giveNextGuess(ctx, chatID, filtered)
	return nil
}


func handleSpecialFeedback(ctx *th.Context, chatID int64, feedback string) (bool,error) {
	wg, exists := game.GetWGame(chatID)
	if !exists {
			_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: игра не найдена. Используй /solve-/help для старта.",
		))
		if err != nil {
				log.Printf("Failed to send error message for chat %d: %v", chatID, err)
			}
		return true, game.ErrGameNotFound
	}
	switch feedback {
	case "NOTFOUND":
		{
			wg.FilteredOutLastGuess()
			giveNextGuess(ctx, chatID, wg.GetPossibleWords())
			return true,nil
		}
	case "LOSE":
		{
			_, err := ctx.Bot().SendMessage(ctx, tu.Message(
				tu.ID(chatID),
				"Эх, проигрыш. Попробуй еще раз, я покажу на что способен!"))
			if err != nil {
				log.Printf("Failed to send lose message for chat %d: %v", chatID, err)
			}
			if err := game.EndGame(chatID); err != nil {
				log.Printf("Failed to end game for chat %d: %v", chatID, err)
			}
			return true, nil
		}
	case "GUESS":
		{
			_, err := ctx.Bot().SendMessage(ctx, tu.Message(
				tu.ID(chatID),
				"Я рад, что смог тебе помочь решить wordle!",
			))
			if err != nil {
				log.Printf("Failed to send guess success message for chat %d: %v", chatID, err)
			}
			if err := game.EndGame(chatID); err != nil {
				log.Printf("Failed to end game for chat %d: %v", chatID, err)
			}
			return true, nil
		}
	}
	return false, nil
}

func giveNextGuess(ctx *th.Context, chatID int64, filtered []string) error {
	if len(filtered) == 0 {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: нет подходящих слов. Начни заново /solve.",
		))
		if err != nil {
			log.Printf("Failed to send no matching words message for chat %d: %v", chatID, err)
		}
		if err := game.EndGame(chatID); err != nil {
			log.Printf("Failed to end game for chat %d: %v", chatID, err)
		}
		return nil//нужно сделать ошибку что filtered words пустые
	}

	nextGuess := game.ChooseNext(filtered)
	wg, exists := game.GetWGame(chatID)
	if !exists {
		_, err := ctx.Bot().SendMessage(ctx, tu.Message(
			tu.ID(chatID),
			"Ошибка: игра не найдена. Используй /solve,/help для старта.",
		))
		if err != nil {
			log.Printf("Failed to send error message for chat %d: %v", chatID, err)
		}
		return game.ErrGameNotFound
	}

	wg.UpdateGameState(filtered, nextGuess)

	mode := wg.GetMode()

	attempt := wg.GetAttempts()

	var message string
	switch mode {
	case "SOLVE":
		message = fmt.Sprintf("Моя %d-ая догадка: **%s**", attempt, nextGuess)
	case "HELP":
		message = fmt.Sprintf("Моя подсказка: **%s**(количество оставшихся попыток для победы: %d)\n\nОтправь новый фидбэк в формате `TRAIN bygbb` или guess, если я угадал.", nextGuess, 5-attempt)
	}

	_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), message))
	if err != nil {
		log.Printf("Failed to send next guess message for chat %d: %v", chatID, err)
	}
	return nil
}