FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o task-service main.go

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/task-service .
EXPOSE 8081
CMD ["./task-service"]