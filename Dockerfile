
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p /app/logs /app/metrics && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

ENV PORT=:8080
ENV LOG_LEVEL=info
ENV HTTP_LOGGING=true
ENV TRACING_ENABLED=false
ENV SERVICE_NAME=employee-management-system

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

CMD ["./server"]

