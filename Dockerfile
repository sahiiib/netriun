FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o netriun ./cmd/server

FROM alpine:latest

WORKDIR /app
RUN addgroup -S netriun && adduser -S netriun -G netriun

COPY --from=builder --chown=netriun:netriun /app/netriun .
COPY --from=builder --chown=netriun:netriun /app/web ./web

ENV PORT=8088
USER netriun

EXPOSE 8088
#HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD wget -qO- "http://127.0.0.1:${PORT}/healthz" >/dev/null || exit 1

CMD ["./netriun"]
