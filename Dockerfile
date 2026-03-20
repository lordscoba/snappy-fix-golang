# ── Stage 1: Builder ──────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache \
    gcc \
    musl-dev \
    libwebp-dev \
    pkgconfig \
    ca-certificates \
    git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# ❌ Remove all the TARGETARCH/TARGETOS complexity — just do this:
RUN CGO_ENABLED=1 go build -ldflags="-w -s" -o /app/out .

# ─── Stage 2: Runtime ─────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache \
    libwebp \
    ca-certificates \
    tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/out .
COPY --chown=appuser:appgroup log.json .

#create logs dir and make it writable by appuser
RUN mkdir -p /app/logs && chown -R appuser:appgroup /app/logs

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=15s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./out"]