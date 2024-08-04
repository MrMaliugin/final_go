FROM golang:1.18-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o main .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/web ./web

EXPOSE 7540

CMD ["./main"]
