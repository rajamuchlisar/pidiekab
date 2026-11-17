# Operations Alerting And Capacity

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `alerting-and-incident-runbook.md`
- `performance-capacity-and-retention.md`

---

## Materi sumber: `alerting-and-incident-runbook.md`

# Alerting & Incident Runbook — PidieKab

## 1. Alert principles

Alert on actionable conditions, not every log entry. Every alert must have:
- a clear condition and threshold;
- severity;
- operator action;
- escalation path;
- suppression/deduplication behavior.

MVP can use a scheduled host check/monitoring agent or external uptime monitor. Avoid implementing a complex alerting platform inside the application.

## 2. Initial alert conditions

| Alert | Suggested trigger | Severity | First response |
|---|---|---|---|
| Website unavailable | External probe fails repeatedly over 2–5 minutes | Critical | Check service, TLS, host/network |
| Go process stopped | systemd service inactive | Critical | Inspect journal and restart policy |
| Readiness failure | PostgreSQL check fails repeatedly | Critical | Check DB/network/credentials |
| Disk low | Free space below 20%, critical below 10% | Warning/Critical | Identify growth; protect DB/media |
| Memory pressure | Sustained high memory or OOM event | Warning/Critical | Inspect process and host pressure |
| SSG queue delayed | Oldest pending job exceeds 5 minutes | Warning | Inspect worker/job errors |
| SSG repeated failure | Same job reaches retry limit | Critical for publish flow | Inspect generation logs; preserve active release |
| Backup stale | No successful backup within 26 hours | Critical | Inspect backup job/destination |
| TLS expiry | Certificate expires within 21 days | Warning | Check ACME renewal/reload |
| DB pool saturation | Acquire waits/timeouts sustained | Warning | Inspect slow queries and pool sizing |

Thresholds are initial proposals and must be tuned to actual host capacity and usage.

## 3. Incident response sequence

1. Confirm alert and determine user impact.
2. Check external availability and service status.
3. Inspect recent deploy/configuration changes.
4. Review structured logs using request ID/time window.
5. Check PostgreSQL readiness, pool, disk, and resource metrics.
6. For SSG issues, verify active release remains valid before retrying.
7. Contain the issue and restore service using documented runbook.
8. Record timeline, impact, root cause, and follow-up action.

## 4. Common incident playbooks

### Go service unavailable
- `systemctl status` service;
- inspect `journalctl` for startup/runtime error;
- verify port binding and TLS certificate/key permissions;
- validate config and database readiness;
- restart only after understanding repeated crash cause.

### PostgreSQL unavailable
- check PostgreSQL service, disk, memory, connection limit, network/firewall;
- verify application credentials and TLS settings;
- do not expose database port publicly as a quick fix;
- preserve database logs and assess recovery before restore.

### SSG queue stalled
- inspect pending/processing jobs and attempt count;
- verify worker loop and DB lock recovery;
- inspect generation output/temp directory permissions and disk;
- retry safely after root cause; do not delete active release.

### Disk pressure
- identify growth in journal, old releases, media, temp builds, and backups;
- never delete current active SSG release or only available backup;
- apply retention policy and verify free space.

### Backup failure
- inspect exit status, storage availability, encryption, and credentials;
- preserve previous known-good backup;
- rerun and verify integrity;
- escalate if backup age breaches recovery objective.

## 5. Post-incident review

For significant incidents, document:
- timeline and detection method;
- affected services/content and duration;
- customer/visitor impact;
- root cause and contributing factors;
- recovery steps;
- prevention, owner, and due date.

Focus on system improvement, not blame.

---

## Materi sumber: `performance-capacity-and-retention.md`

# Performance, Capacity & Retention — PidieKab

## 1. Performance baseline

Before production, benchmark:
- public static page delivery;
- global search under representative data volume;
- Admin content list/editor operations;
- image delivery;
- SSG generation for normal and large content batches;
- PostgreSQL connection pool and slow query behavior.

Record hardware, dataset size, concurrency, p50/p95/p99 latency, throughput, CPU, RAM, disk I/O, and error rate. Do not claim performance targets without measurements.

## 2. Resource monitoring

Monitor:
- CPU and memory for Go process and PostgreSQL;
- open file descriptors and goroutines;
- PostgreSQL connections, locks, slow queries, and table/index growth;
- disk space and I/O;
- media and SSG release directory growth;
- journald and backup staging growth.

Set resource limits and systemd protections consistent with deployment capacity; avoid limits so tight that normal SSG generation is killed.

## 3. SSG capacity

- Generate to a temporary release directory.
- Ensure temporary output has sufficient disk headroom.
- Activate new release atomically only after validation.
- Retain a bounded number of previous releases according to rollback policy.
- Clean abandoned temporary builds only after confirming they are not active or in use.
- Large batch rebuilds should be observable and bounded.

## 4. Database capacity

- Size `pgxpool` based on PostgreSQL connection budget and actual concurrency.
- Reserve connections for maintenance/DBA operations.
- Monitor pool wait time and PostgreSQL `max_connections`.
- Use `EXPLAIN ANALYZE` for slow queries with representative data.
- Add indexes based on observed query plans and query patterns.
- Avoid caching/Redis as a first response to unmeasured query problems.

## 5. Retention baseline

Proposed initial retention, to confirm against storage and legal/privacy needs:
- application operational logs: 14–30 days;
- security/audit logs: longer, controlled retention (initial proposal 12 months);
- SSG releases: current plus a small number of prior known-good releases (initial proposal 3);
- database backups: follow backup specification (7 daily, 4 weekly, 3 monthly proposal);
- temporary build files: remove after success/failure cleanup and a safety grace period.

Retention must not remove records required for active investigation, legal obligations, or recovery. Define access control and deletion process.

## 6. Capacity review triggers

Review capacity when:
- p95 latency breaches agreed SLO repeatedly;
- DB pool wait/timeouts rise;
- disk free space falls below alert threshold;
- SSG queue age grows persistently;
- backup duration approaches backup window;
- process restarts/OOM increase.

Scale only after identifying the bottleneck. Preserve the simple single-binary architecture until measured constraints justify change.

---
