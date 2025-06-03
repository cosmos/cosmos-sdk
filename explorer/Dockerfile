# Multi-stage build for production optimization
FROM node:18-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache \
    python3 \
    make \
    g++ \
    curl

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production && \
    npm cache clean --force

# Copy application code
COPY . .

# Remove development files
RUN rm -rf \
    .git \
    .gitignore \
    README.md \
    docker-compose.yml \
    Dockerfile \
    *.md

# Production stage
FROM node:18-alpine AS production

# Create non-root user for security
RUN addgroup -g 1001 -S tajeor && \
    adduser -S tajeor -u 1001

# Install runtime dependencies
RUN apk add --no-cache \
    curl \
    ca-certificates \
    tzdata && \
    rm -rf /var/cache/apk/*

# Set working directory
WORKDIR /app

# Copy application from builder stage
COPY --from=builder --chown=tajeor:tajeor /app /app

# Create necessary directories
RUN mkdir -p /app/logs /app/blockchain-data && \
    chown -R tajeor:tajeor /app

# Set environment variables
ENV NODE_ENV=production \
    PORT=3000 \
    TZ=UTC

# Switch to non-root user
USER tajeor

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/api/health || exit 1

# Expose port
EXPOSE 3000

# Start application
CMD ["node", "server.js"] 