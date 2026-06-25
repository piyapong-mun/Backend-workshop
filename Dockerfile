FROM golang:1.25.6-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git
RUN go mod tidy
RUN go build -o main main.go
CMD ["./main"]

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
EXPOSE 8000
CMD ["./main"]
