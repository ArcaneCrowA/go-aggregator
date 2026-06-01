FROM golang:1.26 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o web ./cmd/web

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/web .
COPY --from=builder /app/migrations ./migrations
COPY .env .

EXPOSE 8080

ENTRYPOINT [ "./web" ]
