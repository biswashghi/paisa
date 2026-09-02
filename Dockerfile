FROM golang:1.23-bookworm AS api-builder

WORKDIR /src/accts-api

COPY accts-api/go.mod accts-api/go.sum ./
RUN go mod download

COPY accts-api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/paisa-api ./server.go

FROM alpine:3.20 AS api

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=api-builder /out/paisa-api /usr/local/bin/paisa-api
COPY db/schema.sql /app/db/schema.sql

ENV PAISA_HTTP_ADDR=0.0.0.0:8080
ENV PAISA_SCHEMA_PATH=/app/db/schema.sql

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1

CMD ["paisa-api"]

FROM node:20-bookworm-slim AS web-builder

WORKDIR /src/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

FROM caddy:2-alpine AS web

COPY frontend/Caddyfile /etc/caddy/Caddyfile
COPY frontend/docker-entrypoint.sh /usr/local/bin/paisa-web-entrypoint
COPY --from=web-builder /src/frontend/dist /srv
RUN chmod 0755 /usr/local/bin/paisa-web-entrypoint

ENV PAISA_PUBLIC_API_URL=http://127.0.0.1:8080

EXPOSE 80

ENTRYPOINT ["/usr/local/bin/paisa-web-entrypoint"]
CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile", "--adapter", "caddyfile"]
