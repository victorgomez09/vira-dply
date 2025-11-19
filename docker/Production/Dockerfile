# Multi-stage build for optimal image size
FROM node:20-alpine AS frontend-builder
# Set working directory for frontend
WORKDIR /app/web
# Copy frontend package files
COPY web/package.json web/pnpm-lock.yaml ./
# Install pnpm and dependencies
RUN npm install -g pnpm && pnpm install
# Copy frontend source
COPY web/ .
# Build frontend
RUN pnpm build

# Backend builder stage - Use Debian-based image instead of Alpine
FROM golang:1.24-bookworm AS backend-builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
  git \
  ca-certificates \
  gcc \
  g++ \
  pkg-config \
  libgpgme-dev \
  libdevmapper-dev \
  libbtrfs-dev \
  && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend from frontend-builder
COPY --from=frontend-builder /app/web/dist ./assets/dist

# Build the server binary with embedded frontend (CGO enabled for DuckDB)
RUN CGO_ENABLED=1 GOOS=linux go build -a -o mikrocloud-server ./cmd/api

# Final stage - Debian slim for minimal but compatible runtime
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
  ca-certificates \
  tzdata \
  wget \
  libgpgme11 \
  libdevmapper1.02.1 \
  && rm -rf /var/lib/apt/lists/* \
  && adduser --disabled-password --shell /bin/sh --gecos "" mikrocloud

WORKDIR /app

# Copy server binary from backend-builder stage
COPY --from=backend-builder /app/mikrocloud-server .
COPY --from=backend-builder /app/mikrocloud.toml .

# Copy migrations
COPY --from=backend-builder /app/migrations ./migrations

# Change ownership
RUN chown -R mikrocloud:mikrocloud /app

# Switch to non-root user
USER mikrocloud

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Default command
CMD ["./mikrocloud-server", "serve"]
