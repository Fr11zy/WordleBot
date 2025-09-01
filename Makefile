APP := wordle-bot
MAIN := ./cmd/wordle-bot
OUT := bin

.PHONY: build run start

build:
	@mkdir -p $(OUT)
	go build -o $(OUT)/$(APP) $(MAIN)

run:
	go run $(MAIN)

start:
	./$(OUT)/$(APP)

