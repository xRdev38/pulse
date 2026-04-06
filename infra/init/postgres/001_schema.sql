-- ─────────────────────────────────────────────────────────────────────────────
-- Pulse Phase 1 — Schéma PostgreSQL + TimescaleDB
-- ─────────────────────────────────────────────────────────────────────────────

CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Tenants
CREATE TABLE tenants (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- API Keys — stockage SHA-256 uniquement, jamais le plaintext (Ch.8 Security)
CREATE TABLE api_keys (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    key_hash   TEXT NOT NULL UNIQUE,
    key_prefix TEXT NOT NULL,
    scopes     TEXT[] NOT NULL DEFAULT '{"ingest","read"}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_used  TIMESTAMPTZ
);

-- Métriques — TimescaleDB hypertable (Ch.7 Performance)
CREATE TABLE metrics (
    time      TIMESTAMPTZ      NOT NULL,
    tenant_id UUID             NOT NULL,
    name      TEXT             NOT NULL,
    value     DOUBLE PRECISION NOT NULL,
    tags      JSONB            DEFAULT '{}'
);

SELECT create_hypertable('metrics', 'time');
CREATE INDEX ON metrics (tenant_id, name, time DESC);
CREATE INDEX ON metrics USING GIN (tags);

-- Alert Rules
CREATE TABLE alert_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    condition   TEXT NOT NULL CHECK (condition IN ('gt','lt','eq','gte','lte')),
    threshold   DOUBLE PRECISION NOT NULL,
    enabled     BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Audit Log — append-only (Ch.8 Security)
CREATE TABLE audit_log (
    id         BIGSERIAL PRIMARY KEY,
    tenant_id  UUID,
    action     TEXT NOT NULL,
    resource   TEXT,
    actor      TEXT,
    metadata   JSONB        DEFAULT '{}',
    created_at TIMESTAMPTZ  DEFAULT NOW()
);
REVOKE UPDATE, DELETE ON audit_log FROM PUBLIC;

-- Row-Level Security — isolation multi-tenant (Ch.8 Security)
ALTER TABLE metrics     ENABLE ROW LEVEL SECURITY;
ALTER TABLE alert_rules ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_metrics ON metrics
    USING (tenant_id = current_setting('app.tenant_id', true)::UUID);

CREATE POLICY tenant_rules ON alert_rules
    USING (tenant_id = current_setting('app.tenant_id', true)::UUID);

-- ── Seed de développement ─────────────────────────────────────────────────────

INSERT INTO tenants (id, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'acme-corp');

-- Clé de test : "test-key-tenant-1"
-- SHA-256 = f5bd0fc78112354fed66582430cfca923b439973fc97721c63abfc8c32678e2e
INSERT INTO api_keys (tenant_id, key_hash, key_prefix, scopes) VALUES
    (
        '00000000-0000-0000-0000-000000000001',
        'f5bd0fc78112354fed66582430cfca923b439973fc97721c63abfc8c32678e2e',
        'test-key',
        '{"ingest","read","admin"}'
    );
