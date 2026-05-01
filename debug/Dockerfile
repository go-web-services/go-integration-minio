FROM golang:1.24-alpine

# Define build arguments for private module access
ARG GITHUB_TOKEN

ENV GO111MODULE=on \
    GOPRIVATE=github.com/go-web-services/*

# Install dependencies and configure git in one layer
RUN apk add --no-cache git ca-certificates tzdata && \
    git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"

WORKDIR /app
COPY . .

# Run go mod download with the git config in place
RUN go mod download
