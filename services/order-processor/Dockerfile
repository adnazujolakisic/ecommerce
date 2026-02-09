FROM golang:1.23-alpine AS builder

RUN apk add --update --no-cache git || \
    (apk update && apk add --no-cache git)
ENV GOTOOLCHAIN=auto

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o order-processor .

FROM alpine:3.19

RUN apk add --update --no-cache ca-certificates curl || \
    (apk update && apk add --no-cache ca-certificates curl)

WORKDIR /app

COPY --from=builder /app/order-processor .

CMD ["./order-processor"]
