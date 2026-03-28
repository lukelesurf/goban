FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o goban ./cmd/goban

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

WORKDIR /data
COPY --from=builder /app/goban /usr/local/bin/goban

EXPOSE 8080 9090

ENTRYPOINT ["goban"]
CMD ["--db", "/data/goban.db", "serve"]
