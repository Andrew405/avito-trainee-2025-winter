FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY .env /app/.env

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o avito-shop ./cmd/avito-shop/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/avito-shop .
COPY .env /app/.env
COPY . .

EXPOSE 8080

CMD ["./avito-shop"]