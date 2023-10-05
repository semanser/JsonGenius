# Stage 1: Build the application
FROM golang:alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o jsongenius

# Stage 2: Create a minimal image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/jsongenius /app/jsongenius

EXPOSE 8080

CMD ["./jsongenius"]
