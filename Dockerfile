# Этап сборки
FROM golang:1.24 AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/main

# Этап запуска
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из этапа сборки
COPY --from=builder /app/server .

# Открываем порт
EXPOSE 8090

# Запускаем
CMD ["./server"]