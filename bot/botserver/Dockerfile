FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Copy modules first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the whole repo
COPY . .

# Build botserver
WORKDIR /app/bot/botserver
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/botserver

FROM alpine:3.19
RUN apk add --no-cache ca-certificates curl && update-ca-certificates && addgroup -S app && adduser -S app -G app
USER app
WORKDIR /home/app
COPY --from=builder /app/bin/botserver /usr/local/bin/botserver
EXPOSE 8080
ENV LISTEN_ADDR=:8080
ENTRYPOINT ["/usr/local/bin/botserver"]
