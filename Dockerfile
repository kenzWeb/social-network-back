# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/social-network ./cmd

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/social-network /app/social-network
COPY openapi.json /app/openapi.json
COPY uploads /app/uploads

EXPOSE 8080

ENTRYPOINT ["/app/social-network"]
