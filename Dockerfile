FROM golang:1.25-alpine AS generator
WORKDIR /app
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
COPY api/openapi.yaml ./
RUN oapi-codegen -package generated -generate types,chi-server -o internal/generated/api.gen.go ./openapi.yaml

FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY --from=generator /app/internal/generated ./internal/generated
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wallet-service ./cmd/app

FROM golang:1.25-alpine
WORKDIR /app
COPY --from=builder /app/wallet-service /app/config.env ./
COPY --from=builder /app/migrations ./migrations
RUN chmod -R a+r ./migrations
RUN adduser -D -s /bin/sh appuser
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1
CMD ["./wallet-service"]