FROM golang:1.25 AS builder

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# runtime image
FROM golang:1.25
COPY --from=builder /go/bin/goose /usr/local/bin/goose
WORKDIR /app
CMD ["goose", "up"]