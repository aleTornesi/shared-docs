FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.work go.work.sum ./
COPY db/ db/
COPY kafka/ kafka/
RUN cd kafka && go build -o /kafka-processor .

FROM alpine:3.21
COPY --from=builder /kafka-processor /kafka-processor
ENTRYPOINT ["/kafka-processor"]
