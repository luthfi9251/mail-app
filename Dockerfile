# syntax=docker/dockerfile:1

# ---- 1. Build the React SPA ----
FROM node:24-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

# ---- 2. Build the Go API (static binary) ----
FROM golang:1.27-alpine AS server
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/email-app .

# ---- 3. Runtime: no shell, no package manager, non-root ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=server /out/email-app /app/email-app
COPY --from=web /src/web/dist /app/web
ENV GIN_MODE=release \
    PORT=8080 \
    STATIC_DIR=/app/web
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/email-app"]
