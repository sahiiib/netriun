FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o netriun ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/netriun .
COPY --from=builder /app/web ./web

EXPOSE 8088

CMD ["./netriun"]