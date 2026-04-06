# Pulse — Phase 1

Plateforme de monitoring distribué — implémentation des tactiques du livre
**Software Architecture in Practice** (Bass, Clements, Kazman).

## Démarrage Windows

```powershell
# 1. Lance Docker Desktop et attends qu'il soit prêt
# 2. Dans le dossier pulse/ :
.\pulse.ps1 build        # compile les images (3-5 min la 1ere fois)
.\pulse.ps1 up           # démarre les 7 services
# Attendre 30 secondes...
.\pulse.ps1 health       # doit afficher "healthy"
.\pulse.ps1 ingest-test  # envoie une métrique de test
```

## Structure

```
pulse/
├── pulse.ps1                        ← script principal (Windows)
├── services/
│   ├── collector/                   ← Go + Gin (port 8081)
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── cmd/server/              ← main.go + helpers.go
│   │   └── internal/
│   │       ├── pipeline/            ← Pipe-and-Filter (Ch.13)
│   │       ├── circuit/             ← Circuit Breaker (Ch.4)
│   │       ├── health/              ← Health Check (Ch.4)
│   │       └── ingest/              ← Handler + connexions DB/AMQP
│   └── gateway/                     ← NestJS (port 3000)
│       ├── Dockerfile
│       ├── package.json
│       └── src/
│           ├── auth/                ← API Key Guard SHA-256 (Ch.8)
│           ├── metrics/             ← POST /api/metrics/ingest
│           ├── health/              ← GET /health
│           └── common/
└── infra/
    ├── docker-compose.yml
    ├── init/postgres/001_schema.sql
    └── prometheus/prometheus.yml
```

## Tactiques architecturales implémentées (Phase 1)

| Tactique | Chapitre | Implémentation |
|---|---|---|
| Pipe-and-Filter | Ch.13 | `collector/internal/pipeline/` |
| Circuit Breaker | Ch.4 | `collector/internal/circuit/breaker.go` |
| Retry + backoff | Ch.4 | `gateway/src/metrics/metrics.service.ts` |
| Health Check | Ch.4 | `GET /health` sur chaque service |
| Rate Limiting | Ch.4 | ThrottlerModule NestJS (10 req/s) |
| Authenticate Actors | Ch.8 | SHA-256 API Key lookup |
| Row-Level Security | Ch.8 | PostgreSQL RLS par tenant_id |
| Audit Log append-only | Ch.8 | Table audit_log + REVOKE |

## API Key de test

```
Header : X-API-Key: test-key-tenant-1
Tenant : 00000000-0000-0000-0000-000000000001
```

## Endpoints

| Méthode | URL | Description |
|---|---|---|
| POST | http://localhost:3000/api/metrics/ingest | Ingérer une métrique |
| GET | http://localhost:3000/health | Health du Gateway |
| GET | http://localhost:8081/health | Health du Collector |
| GET | http://localhost:8081/metrics | Métriques Prometheus |

## Interfaces web

| Service | URL | Credentials |
|---|---|---|
| Grafana | http://localhost:3001 | admin / admin |
| RabbitMQ | http://localhost:15672 | pulse / pulse_secret |
| Prometheus | http://localhost:9090 | — |

## Exemple d'ingestion

```powershell
# PowerShell
$headers = @{ 'Content-Type' = 'application/json'; 'X-API-Key' = 'test-key-tenant-1' }
$body = '{"name":"cpu.usage","value":87.5,"tags":{"host":"web-01"}}'
Invoke-RestMethod -Uri 'http://localhost:3000/api/metrics/ingest' -Method POST -Headers $headers -Body $body
```

```bash
# curl (si installé)
curl -X POST http://localhost:3000/api/metrics/ingest \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-key-tenant-1" \
  -d '{"name":"cpu.usage","value":87.5,"tags":{"host":"web-01"}}'
```
