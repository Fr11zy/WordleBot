APP := wordle-bot
MAIN := ./cmd/wordle-bot
OUT := bin
BIN := $(OUT)/$(APP)

.PHONY: build run start up down

up:
	docker-compose up --build -d && docker-compose logs -f

down:
	docker-compose down --remove-orphans

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

