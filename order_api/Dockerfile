# --- Builder ---
FROM golang:1.26.2-trixie AS builder
ENV CGO_ENABLED=1
WORKDIR /app
RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    build-essential \
    && rm -rf /var/lib/apt/lists/*
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o order_api ./cmd/main

# --- Runner ---
FROM debian:trixie-slim
ARG APP_PATH=/opt/order_api
WORKDIR $APP_PATH
RUN apt-get update && apt-get install -y \
    ca-certificates \
    netcat-openbsd \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/order_api ./

ENV PORT=8080
EXPOSE ${PORT}

ENTRYPOINT ["./order_api"]
