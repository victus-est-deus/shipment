FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o shipment-server ./cmd/server

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/shipment-server .
COPY --from=builder /app/internal/infrastructure/persistence/postgres/migrations ./migrations

EXPOSE 50051

CMD ["./shipment-server"]
