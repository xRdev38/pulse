# ADR-001 : Choix du stockage time-series — TimescaleDB vs InfluxDB

**Date :** 2024-01  
**Statut :** Accepté  
**Décideurs :** Équipe Pulse  
**Contexte du livre :** Bass/Clements/Kazman — Ch.17 (Documenting Software Architectures) + Ch.19 (ATAM)

---

## Contexte et Problème

Pulse ingère des métriques horodatées à haute fréquence (potentiellement >10k/s par tenant).
Le stockage doit supporter :
- Des insertions en masse rapides (Performance)
- Des agrégations sur des fenêtres temporelles arbitraires (requêtes TSDB)
- Une isolation multi-tenant (Security/Modifiability)
- Une évolutivité sans migration massive (Modifiability)

---

## Options Considérées

| Critère | **TimescaleDB** | **InfluxDB OSS** | **PostgreSQL vanilla** |
|---|---|---|---|
| Modèle | Extension PostgreSQL | Base propriétaire | Relationnel pur |
| Query language | SQL standard | Flux / InfluxQL | SQL standard |
| RLS multi-tenant | ✅ natif PostgreSQL | ❌ non supporté | ✅ natif |
| Joins avec autres tables | ✅ SQL JOIN | ❌ impossible | ✅ |
| Compression | ✅ ~95% | ✅ ~95% | ❌ |
| Partitionnement temps | ✅ hypertable auto | ✅ shards | ❌ manuel |
| Licence | Apache 2.0 / Timescale | BSL (InfluxDB 3 fermé) | PostgreSQL |
| Opérateur Docker | timescale/timescaledb | influxdb | postgres |
| ORM compatible | Prisma, pgx, pg | SDK propriétaire | Tous |
| Alert rules (JOIN) | ✅ | ❌ | ✅ |

---

## Analyse ATAM (Ch.19)

### Scénarios de qualité évalués

**Scénario 1 — Performance**
> "Le système ingère 5000 métriques/s pendant 1h et une query agrège p95 sur 24h en <500ms."

- TimescaleDB : hypertable partitionne par chunk de temps → query ne lit que les chunks pertinents. Continuous Aggregates pré-calculent les agrégats. **Répond au scénario.**
- InfluxDB : performant sur ce cas. Mais query language Flux est non-standard.
- PostgreSQL vanilla : sans partitionnement, seq scan sur 18M lignes → **ne répond pas.**

**Scénario 2 — Modifiability**
> "Dans 3 mois, on ajoute des alert_rules qui JOINent sur les métriques récentes."

- TimescaleDB : `SELECT * FROM metrics m JOIN alert_rules a ON m.name = a.metric_name WHERE m.time > NOW() - INTERVAL '5m'` → SQL natif. **Coût de changement : faible.**
- InfluxDB : impossible de JOINer avec une table externe. Nécessite un service dédié d'évaluation. **Coût de changement : élevé.**

**Scénario 3 — Security (Multi-tenant)**
> "Un tenant ne peut jamais voir les métriques d'un autre tenant."

- TimescaleDB : Row-Level Security PostgreSQL, activé via `SET LOCAL app.tenant_id`. **Garanti au niveau DB.**
- InfluxDB : isolation par organisation/bucket → nécessite une logique applicative. **Risque d'erreur de configuration.**

### Tradeoffs identifiés (style ATAM)

| Tradeoff | TimescaleDB | InfluxDB |
|---|---|---|
| Perf pure time-series | Légèrement inférieur | Supérieur |
| Flexibilité query | SQL universel | Flux puissant mais propriétaire |
| Vendor lock-in | Faible (PostgreSQL) | Élevé (InfluxDB 3 = cloud only) |
| Opérationnel | 1 seul service DB | 2 services (Influx + Postgres pour relations) |
| Expertise équipe | SQL connu | Flux à apprendre |

---

## Décision

**TimescaleDB est retenu.**

**Raison principale :** La contrainte de multi-tenant RLS et les alert_rules relationnelles imposent PostgreSQL comme base. TimescaleDB est une extension PostgreSQL — on obtient les performances TSDB **sans sacrifier** SQL, les JOINs, le RLS, ni l'écosystème d'outils (Prisma, pgx, pg).

Le léger déficit de performance pure par rapport à InfluxDB est compensé par les **Continuous Aggregates** (vues matérialisées rafraîchies automatiquement).

---

## Conséquences

- ✅ Un seul service de base de données à opérer
- ✅ RLS multi-tenant garanti au niveau DB
- ✅ SQL standard → zéro apprentissage supplémentaire
- ✅ Prisma/pgx compatibles sans adaptation
- ⚠️ Nécessite l'extension TimescaleDB (image Docker dédiée)
- ⚠️ Les Continuous Aggregates doivent être configurés manuellement (Phase 2)

---

## Références

- Bass, L., Clements, P., Kazman, R. — *Software Architecture in Practice*, 4th ed., Ch.19 ATAM
- TimescaleDB docs : https://docs.timescale.com
- Benchmark TimescaleDB vs InfluxDB : https://www.timescale.com/blog/timescaledb-vs-influxdb
