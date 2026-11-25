FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /url-shortener ./cmd

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /url-shortener .
COPY config ./config

EXPOSE 8082

CMD ["./url-shortener"]