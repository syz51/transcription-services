# Build stage
FROM golang:1.24.4-alpine AS builder

# Set working directory
WORKDIR /app

ENV GOCACHE=/root/.cache/go-build

# Copy source code
COPY . .

# Build the application
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    CGO_ENABLED=0 GOOS=linux go build -a -o server cmd/server/main.go

# --- FFmpeg static binary stage ---
FROM alpine:3.19 AS ffmpeg
# Install minimal tools and download the static build
RUN --mount=type=cache,target=/var/cache/apk \
    apk add --no-cache curl xz \
    && curl -L -o /tmp/ffmpeg.tar.xz https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz \
    && mkdir -p /ffmpeg \
    && tar -xJf /tmp/ffmpeg.tar.xz -C /ffmpeg --strip-components=1 \
    && rm /tmp/ffmpeg.tar.xz

# Final stage
FROM gcr.io/distroless/static-debian12:nonroot

# Create app directory
WORKDIR /app

# Copy the AWS Lambda Web Adapter and ffmpeg
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.9.1 /lambda-adapter /opt/extensions/lambda-adapter

# Copy the ffprobe binary from ffmpeg stage
COPY --from=ffmpeg /ffmpeg/ffprobe /usr/local/bin/ffprobe

# Copy the binary from builder stage
COPY --from=builder /app/server .

# Copy config directory
COPY --from=builder /app/configs ./configs

# Expose port
EXPOSE 8080

# Command to run
CMD ["/app/server"]