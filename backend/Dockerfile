FROM golang:1.22 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api

FROM alpine:3.18

WORKDIR /root/
COPY --from=builder /app/app .

CMD ["./app"]
