FROM golang:1.25 AS builder

WORKDIR /app

COPY docs/go.mod .
COPY docs/go.sum .

COPY db /db

RUN ["go", "mod", "download"]

COPY docs/ .

RUN ["go", "build", "-o", "docs"]

FROM alpine:3.17

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/docs .

CMD ["/app/docs"]
