# === Build stage ===
FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener ./cmd

FROM alpine:latest  

WORKDIR /root/

RUN apk --no-cache add tzdata

COPY --from=builder /app/url-shortener .

ENV SERVER_HOST=0.0.0.0
ENV SERVER_PORT=8080

EXPOSE 8080

CMD ["./url-shortener"]