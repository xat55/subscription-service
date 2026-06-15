FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -mod=vendor -o /subscription-service ./cmd/app

FROM alpine:latest

COPY --from=builder /subscription-service /subscription-service

EXPOSE 8080
CMD ["/subscription-service"]
