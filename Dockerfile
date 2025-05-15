FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 go build -o /go-mailer

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go-mailer /go-mailer
RUN adduser -D appuser
USER appuser