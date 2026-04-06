FROM golang:1.22-alpine AS builder

WORKDIR /app
RUN apk --no-cache add git ca-certificates

# Copier TOUT le code source d'abord
COPY . .

# go mod tidy voit maintenant tous les imports et génère un go.sum complet
RUN go mod tidy

# Compiler
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/collector ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add wget ca-certificates tzdata
RUN addgroup -S pulse && adduser -S pulse -G pulse
USER pulse
COPY --from=builder /bin/collector /bin/collector
EXPOSE 8081
HEALTHCHECK --interval=10s --timeout=5s --retries=5 \
    CMD wget -qO- http://localhost:8081/health || exit 1
ENTRYPOINT ["/bin/collector"]
