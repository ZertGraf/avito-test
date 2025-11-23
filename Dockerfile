# build stage
FROM golang:1.24 AS builder

# install build dependencies (debian uses apt-get, not apk)
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    make \
    && rm -rf /var/lib/apt/lists/*

# set working directory
WORKDIR /build

# copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# copy source code
COPY . .

# build binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o pr-reviewer-service \
    ./cmd/server

# runtime stage
FROM alpine:3.19

# install ca-certificates for https connections
RUN apk --no-cache add ca-certificates tzdata wget

# create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

WORKDIR /app

# copy binary from builder
COPY --from=builder /build/pr-reviewer-service .

# copy migrations
COPY --from=builder /build/migrations ./migrations

# change ownership
RUN chown -R app:app /app

# switch to non-root user
USER app

# expose port
EXPOSE 8080

# health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# run binary (ВАЖНО: правильное имя бинарника)
ENTRYPOINT ["./pr-reviewer-service"]