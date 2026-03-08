FROM golang:1.25 AS builder

WORKDIR /app

RUN ["go", "install", "github.com/a-h/templ/cmd/templ@latest"]

COPY frontend/go.mod frontend/go.sum ./

COPY db /db

RUN ["go", "mod", "download"]

COPY frontend/ .

RUN ["templ", "generate"]
RUN ["go", "build", "-o", "frontend"]

FROM alpine:3.17

RUN ["apk", "add", "--no-cache", "ca-certificates"]

WORKDIR /app

COPY --from=builder /app/frontend .

CMD ["/app/frontend"]
