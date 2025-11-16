# Build arguments with defaults
ARG GO_VERSION=1.25
ARG ALPINE_VERSION=3.20
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE=unknown

# Build stage
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Install build dependencies
RUN apk add --no-cache \
    gcc \
    libc-dev \
    make \
    git \
    ca-certificates \
    tzdata \
    postgresql-dev \
    postgis-dev

WORKDIR /app

# Copy module files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Set Go environment variables for production
ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64 \
    GOMAXPROCS=4

# Build arguments
ARG APP=collector
ARG VERSION
ARG COMMIT_SHA
ARG BUILD_DATE
ENV APP=${APP} \
    VERSION=${VERSION} \
    COMMIT_SHA=${COMMIT_SHA} \
    BUILD_DATE=${BUILD_DATE}

# Copy the rest of the application
COPY . .

# Build the application with version information
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s -extldflags '-static' \
    -X 'main.version=${VERSION}' \
    -X 'main.commit=${COMMIT_SHA:0:7}' \
    -X 'main.buildDate=${BUILD_DATE}' \
    -X 'main.goVersion=$(go version | cut -d ' ' -f 3)'" \
    -trimpath \
    -o /bin/${APP} \
    ./cmd/${APP}

# Final stage
FROM alpine:${ALPINE_VERSION}

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    postgis \
    curl

# Create non-root user
RUN addgroup -g 10001 -S appuser && \
    adduser -u 10001 -S appuser -G appuser && \
    mkdir -p /app/migrations /app/web /app/static && \
    chown -R appuser:appuser /app

WORKDIR /app

# Copy application binary
COPY --from=builder /bin/collector /bin/web /bin/
COPY --from=builder /app/migrations /app/migrations/
COPY --from=builder /app/web /app/web/

# Build arguments
ARG APP=collector
ARG VERSION
ARG COMMIT_SHA
ARG BUILD_DATE
ENV APP=${APP} \
    VERSION=${VERSION} \
    COMMIT_SHA=${COMMIT_SHA} \
    BUILD_DATE=${BUILD_DATE}

# Set permissions
RUN chmod +x /bin/collector /bin/web && \
    chown -R appuser:appuser /bin/collector /bin/web

# Expose ports
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Labels for container metadata
LABEL org.opencontainers.image.title="lot-collector" \
      org.opencontainers.image.description="Land lots monitoring service with Yandex Maps integration" \
      org.opencontainers.image.url="https://github.com/Promcalc/zemla_02" \
      org.opencontainers.image.source="https://github.com/Promcalc/zemla_02" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${COMMIT_SHA}" \
      org.opencontainers.image.licenses="MIT"

# Run as non-root user
USER appuser

# Start the application
ENTRYPOINT ["/bin/${APP}"]