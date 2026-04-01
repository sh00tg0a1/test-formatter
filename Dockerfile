# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go build -o /bin/param-formatter ./main.go

FROM alpine:3.20
RUN adduser -D -g '' appuser
USER appuser
WORKDIR /home/appuser

COPY --from=builder /bin/param-formatter /usr/local/bin/param-formatter

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/param-formatter"]
