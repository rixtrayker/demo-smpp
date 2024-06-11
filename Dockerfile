FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=arm64 go build -buildvcs=false -ldflags="-s -w" -o app cmd/worker/main.go

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/app .
COPY --from=builder /app/config.yaml .

RUN apk add --no-cache git

EXPOSE 2112

CMD ["./app"]