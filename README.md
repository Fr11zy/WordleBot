# ***WordleBot***
***
<small>Сервис-ассистент для решения ежедневных задач игры Wordlе от NYT, а также для помощи пользователям в их решении.</small>

# *Установка и запуск*
***
## Получение Telegram Bot Token
- Напишите @BotFather в Telegram
- Используйте команду /newbot для создания нового бота
- Скопируйте полученный токен и добавьте его в файл .env как показано далее
## Установка
```
git clone https://github.com/Fr11zy/WordleBot.git
cd WordleBot
echo "TG_TOKEN=your_telegram_bot_token" > .env
```
## Запуск
- Docker: 
  - Через docker-compose
  ```
  docker-compose build
  docker-compose up
  ```
  - Через make
  ```
  make up
  ```
- Make:
  - Создание и запуск бинарника
  ```
  make build
  make start
  ```
  - Запуск main.go
  ```
  make run
  ```
- Go:
  ```
  go run cmd/wordle-bot/main.go
  ```

# Использование
 - `/start` - показывает краткое описание и Как взаимодействовать с ботом
 - `/solve` - решает с нуля задачу игры Wordle
 - `/help`  - помогает пользователю в решении задачи игры Wordle с любого шага