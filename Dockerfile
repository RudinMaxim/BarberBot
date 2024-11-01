FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /barber-bot ./cmd/main.go

FROM alpine:latest  

WORKDIR /root/

COPY --from=builder /barber-bot .

COPY texts.yaml ./texts.yaml
COPY config.yaml ./config.yaml
COPY credentials/credentials.json ./credentials/credentials.json
COPY credentials/token.json ./credentials/token.json

EXPOSE 8080

CMD ["./barber-bot"]