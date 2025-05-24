FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY ./go-signsafe/go.mod ./go-signsafe/go.sum ./

RUN go mod download

COPY ./go-signsafe .

RUN go build -o app .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8000

CMD ["./app"]

