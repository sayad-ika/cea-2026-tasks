# ── Build stage ──────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Cache module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy full source (cmd/* + internal/*)
COPY . .

# Build the self binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /bootstrap ./cmd/self

# ── Runtime stage ────────────────────────────────────────────
FROM public.ecr.aws/lambda/provided:al2023

COPY --from=builder /bootstrap ${LAMBDA_TASK_ROOT}/bootstrap

ENTRYPOINT ["/bootstrap"]
