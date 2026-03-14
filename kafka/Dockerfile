FROM golang:1.25 AS builder

WORKDIR /app

COPY kafka/go.mod .
COPY kafka/go.sum .

COPY db /db

RUN ["go", "mod", "download"]

COPY kafka/ .

RUN CGO_ENABLED=0 go build -o kafka

FROM alpine:3.17

RUN ["apk", "add", "--no-cache", "ca-certificates"]

WORKDIR /app

COPY --from=builder /app/kafka .

CMD ["/app/kafka"]
