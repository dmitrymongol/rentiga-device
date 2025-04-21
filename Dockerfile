FROM golang:1.24-bookworm as builder

RUN apt-get update && apt-get install -y \
    libgtk-3-dev \
    libcairo2-dev \
    libglib2.0-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o ip-display .

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    libgtk-3-0 \
    libcairo2 \
    libglib2.0-0 \
    xauth \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/ip-display /ip-display

CMD ["/ip-display"]