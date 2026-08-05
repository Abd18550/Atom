# ─── Stage 1: Build ───
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy dependency files first (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /atom-backend .

# ─── Stage 2: Runtime ───
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for HTTPS and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /atom-backend .

# Copy questions data (needed for seeding)
COPY --from=builder /app/questions ./questions

# Copy sandbox config (needed at runtime even though Docker sandbox won't work on Render)
COPY --from=builder /app/sandbox ./sandbox

EXPOSE 8080

CMD ["./atom-backend"]
