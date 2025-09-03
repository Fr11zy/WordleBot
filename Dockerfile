FROM golang:1.24.1-alpine AS builder
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/wordle-bot ./cmd/wordle-bot

FROM alpine:3.18
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/wordle-bot .
COPY --from=builder /app/assets ./assets

CMD ["./wordle-bot"]