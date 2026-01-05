# Stage 1: Build Stage
FROM golang:1.24-alpine AS builder

# Define build arguments for private module access
ARG GITHUB_TOKEN

ENV GO111MODULE=on \
    GOPRIVATE=github.com/Lomank123/*

# Install dependencies and configure git in one layer
RUN apk add --no-cache git ca-certificates tzdata && \
    git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

WORKDIR /app

COPY go.mod go.sum ./
COPY pkg/client/go.mod ./pkg/client/go.mod

RUN go mod download

COPY . .

RUN go build -o go-integration-minio ./cmd/app/main.go

FROM alpine:3.20.3 AS final

WORKDIR /app

COPY --from=builder /app/go-integration-minio .

CMD ["./go-integration-minio"]
