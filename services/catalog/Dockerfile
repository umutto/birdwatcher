# syntax=docker/dockerfile:1

# Stage 1: Build binary
FROM golang:1.27-alpine AS builder

WORKDIR /workspace

# Install system dependencies (certificates and timezone data)
RUN apk add --no-cache ca-certificates tzdata

# Copy only catalog and the shared/generated modules it depends on
COPY gen/go/ ./gen/go/
COPY shared/ ./shared/
COPY services/catalog/ ./services/catalog/

# Initialize a scoped workspace containing only catalog and its dependencies
RUN go work init \
    ./services/catalog \
    ./gen/go \
    ./shared/logger \
    ./shared/nats \
    ./shared/telemetry

# Build the statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-w -s" \
    -o /bin/catalog \
    ./services/catalog/cmd/catalog

# Stage 2: Minimal runtime image
FROM alpine:3.20 AS runner

WORKDIR /app

# Run as a non-privileged user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /bin/catalog /usr/local/bin/catalog

USER appuser

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/catalog"]
