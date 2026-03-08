FROM golang:1.25 AS builder

WORKDIR /app

COPY ws/go.mod .
COPY ws/go.sum .

COPY db /db

RUN ["go", "mod", "download"]

COPY ws/ .

RUN ["go", "build", "-o", "ws"]

FROM alpine:3.17

RUN ["apk", "add", "--no-cache", "ca-certificates"]

WORKDIR /app

COPY --from=builder /app/ws .

CMD ["/app/ws"]
