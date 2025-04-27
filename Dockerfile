FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
COPY .env .
ENV CGO_ENABLED=0
RUN go mod download
RUN go build -o pvz-service ./cmd

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/pvz-service .
COPY --from=builder /app/migrations ./migrations

RUN chmod +x pvz-service
CMD ["./pvz-service"]
