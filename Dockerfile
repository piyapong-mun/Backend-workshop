FROM golang:1.25.6-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git
RUN apk add --no-cache curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-386.tar.gz | tar xvz
RUN go mod tidy
RUN go build -o main main.go
CMD ["./main"]

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY ./db/migration ./migration
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY ./eks .
RUN chmod +x migrate
RUN chmod +x start.sh
RUN chmod +x wait-for.sh
EXPOSE 8000
CMD ["./main"]
ENTRYPOINT [ "./start.sh" ]
