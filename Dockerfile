FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o avito-shop ./cmd/avito-shop

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/avito-shop .

EXPOSE 8080
CMD ["./avito-shop"]