# Стадия сборки
FROM golang:1.24 as builder

WORKDIR /app

# Копируем модульные файлы для загрузки зависимостей
COPY go.mod go.sum ./

# Копируем исходники и статические файлы
COPY main.go .
COPY data ./data
COPY static ./static
COPY templates ./templates

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o dayzhelper .

# Финальный легкий образ
FROM alpine:latest

WORKDIR /root/

# Копируем бинарник и статические файлы из стадии сборки
COPY --from=builder /app/dayzhelper .
COPY --from=builder /app/data ./data
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Устанавливаем права на папки
RUN chmod -R 755 ./data && \
    chmod -R 755 ./static && \
    chmod -R 755 ./templates

# Открываем порт
EXPOSE 3000

# Запускаем приложение
CMD ["./dayzhelper"]
