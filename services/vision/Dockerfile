# syntax=docker/dockerfile:1

# Stage 1: Build virtualenv using official uv binary
FROM ghcr.io/astral-sh/uv:latest AS uv
FROM python:3.12-slim-bookworm AS builder

ENV UV_COMPILE_BYTECODE=1
ENV UV_LINK_MODE=copy

WORKDIR /workspace

# Install uv from the official uv image
COPY --from=uv /uv /uvx /bin/

# Copy project definition and source
COPY gen/python/ ./gen/python/
COPY services/vision/ ./services/vision/

# Install dependencies into virtual environment
RUN cd services/vision && uv sync --no-dev

# Stage 2: Minimal runtime image
FROM python:3.12-slim-bookworm AS runner

WORKDIR /app

# Run as non-root user
RUN groupadd -r appgroup && useradd -r -g appgroup appuser

# Copy virtualenv and code from builder
COPY --from=builder /workspace/services/vision/.venv /app/.venv
COPY --from=builder /workspace/gen/python /app/gen/python
COPY --from=builder /workspace/services/vision /app/services/vision

ENV PATH="/app/.venv/bin:$PATH"
ENV PYTHONPATH="/app/gen/python:/app/services/vision/src:$PYTHONPATH"

USER appuser

ENTRYPOINT ["python", "-m", "vision"]
