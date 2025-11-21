# --- Stage 1: Builder ---
FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod go.sum ./
COPY telegram-bot-api ./telegram-bot-api
RUN go mod download
COPY . .


RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o bot ./cmd/bot

# --- Stage 2: Minimal runtime ---
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /app/bot .
USER nonroot:nonroot
ENTRYPOINT ["/app/bot"]