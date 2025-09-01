APP := wordle-bot
MAIN := ./cmd/wordle-bot
OUT := bin
BIN := $(OUT)/$(APP)

.PHONY: build run start

build:
	@mkdir -p $(OUT)
	go build -o $(OUT)/$(APP) $(MAIN)

run:
	go run $(MAIN)

start:
	@if [ -f "$(BIN)" ]; then \
		$(BIN); \
	else \
		echo "❌ Бинарь не найден! Сначала собери его: make build"; \
		exit 1; \
	fi

