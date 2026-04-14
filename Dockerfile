# ── Stage 1: Builder ──────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o bin/warungku .

# ── Stage 2: Runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary
COPY --from=builder /app/bin/warungku .

# Copy runtime assets (templates + static must travel with the binary)
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static    ./static
COPY --from=builder /app/db        ./db

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 3000

# App reads DB_URL, JWT_SECRET, PORT from environment
CMD ["./warungku"]
