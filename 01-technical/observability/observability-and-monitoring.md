# Observability And Monitoring

> Dokumen konsolidasi berdasarkan topik. Materi sumber dipertahankan agar persyaratan dan keputusan tidak hilang.

## Sumber yang digabung

- `observability-architecture.md`
- `logging-and-request-correlation.md`
- `health-checks-and-metrics.md`

---

## Materi sumber: `observability-architecture.md`

# Observability Architecture — PidieKab

**Status:** MVP operational baseline  
**Architecture alignment:** One Go binary, one `net/http` server, PostgreSQL, in-process SSG worker, systemd, no Redis/message broker.

## 1. Objectives

Observability harus membantu operator menjawab:
- Apakah website publik dan Admin API tersedia?
- Apakah request lambat atau gagal?
- Apakah PostgreSQL dapat diakses dan pool koneksinya sehat?
- Apakah SSG worker tertinggal, gagal, atau macet?
- Apakah kapasitas CPU, RAM, disk, dan backup memadai?
- Kapan terjadi perubahan penting atau security event?

MVP menggunakan structured logs, health endpoints, internal metrics, systemd/journald, dan OS monitoring sederhana. Hindari menambah platform observability berat sebelum ada kebutuhan nyata.

## 2. Signals

| Signal | MVP source | Use |
|---|---|---|
| Logs | Go `log/slog` JSON to stdout/stderr | Diagnosis, audit correlation |
| Health | Go HTTP health/readiness endpoints | Service and dependency state |
| Metrics | Lightweight in-process counters/gauges/histograms or standard exposition endpoint | Request, DB pool, SSG worker |
| Traces | Request ID propagation | Correlate one request across logs |
| Host metrics | OS/systemd tools | CPU, memory, disk, process |
| Backup status | Backup job result + timestamp | Recovery readiness |

Distributed tracing and centralized log platform are deferred unless deployment topology or incident volume justifies them.

## 3. Runtime boundaries

- Go application is one service and owns HTTP, API, and in-process SSG worker.
- Public static files are served by the same Go service.
- PostgreSQL is the primary persistent data store and SSG job queue.
- No Redis or message broker in MVP.
- Health/metrics endpoints must not expose credentials, filesystem paths, private content, or detailed database errors.
- Internal operational endpoints should be restricted to loopback/private management access where possible.

## 4. Request correlation

- Generate a server-side request ID for each request.
- Accept an incoming correlation ID only after strict length/character validation; otherwise generate a new ID.
- Return the request ID in a response header.
- Include request ID in structured request/error logs.
- Do not use request ID as a metric label because it creates unbounded cardinality.

## 5. Operational ownership

- HTTP middleware: request logs, latency, status, request ID.
- Database platform: pool stats and connectivity.
- Publication worker: queue depth, age, retries, failures, processing duration.
- systemd/host: process restart, resource use, disk capacity.
- Backup scripts: success/failure and last successful backup.
- Operator: alert response, incident record, recovery drill.

## 6. Privacy and retention

Logs must not contain passwords, cookies, session tokens, CSRF tokens, Authorization headers, full request bodies, private source notes, or sensitive Admin content. Define retention and access permissions for logs and metrics. Audit logs remain the business/security event record; application logs are operational diagnostics and must not replace audit logs.

---

## Materi sumber: `logging-and-request-correlation.md`

# Logging & Request Correlation — PidieKab

## 1. Logging implementation

Use Go `log/slog` with JSON handler. Emit to stdout/stderr so systemd captures output in journald. Avoid a custom logging framework.

Recommended fields:
- `time` (UTC);
- `level`;
- `msg`;
- `service`;
- `environment`;
- `request_id` where applicable;
- `method`, normalized `route`, `status`, `duration_ms`;
- `admin_id` only when necessary and permitted;
- `module`, `operation`, `error_code`.

Do not log raw URL query strings by default; they may contain search terms or accidental secrets. Prefer route templates over concrete paths to avoid high-cardinality values.

## 2. Log levels

- DEBUG: temporary diagnostic detail; disabled in normal production.
- INFO: startup/shutdown, successful operational milestones, completed request summary where useful.
- WARN: recoverable failures, retry, throttling, degraded dependency.
- ERROR: failed operation requiring attention.
- FATAL is not a routine level; controlled shutdown should be used instead of abrupt process termination.

## 3. Redaction policy

Never log:
- password or password hash;
- session cookie/token or CSRF token;
- Authorization header;
- TLS private key, database URL/password, API tokens;
- full upload bytes or full request bodies;
- unpublished article body or private editorial notes;
- personal information not required for diagnosis.

Use allowlisted structured fields. Redact user-controlled values before logging and cap field length.

## 4. HTTP request logs

Record method, route template, status, duration, request ID, and response size when available. Exclude health probe noise or sample it carefully, but preserve failures. Do not log full IP addresses indefinitely without a documented security need and retention policy.

## 5. Error handling

- Public responses use the established RFC 9457 problem response format.
- Client receives a stable error code and request ID.
- Internal logs may include stack/context only when safe and supported; never expose it to public clients.
- Avoid logging the same error at every layer. Log once at the boundary that can add useful context.

## 6. Journald operations

Use systemd journal limits/retention policy appropriate to disk capacity. Ensure service logs are readable only by authorized operators. Monitor journal disk use and avoid unbounded verbose logging.

---

## Materi sumber: `health-checks-and-metrics.md`

# Health Checks & Metrics — PidieKab

## 1. Health endpoint semantics

Expose distinct endpoints with minimal responses:

- `/health/live`: process is alive and event loop/HTTP server can respond. Must not depend on PostgreSQL.
- `/health/ready`: service can accept normal traffic. Checks critical dependencies such as PostgreSQL with a short timeout.
- Optional internal diagnostic endpoint: detailed dependency state, restricted to loopback/private access.

Do not expose raw dependency errors or configuration through public health output. A readiness failure should remove the instance from traffic if an external traffic manager is ever introduced; current single-instance deployment should return a non-ready status and trigger alerting.

## 2. Probe behavior

- Health handlers must be fast and bounded by deadlines.
- Avoid expensive database queries.
- Use a short PostgreSQL ping or lightweight check for readiness.
- Do not let concurrent health probes exhaust the DB pool.
- Liveness must not restart the process merely because PostgreSQL is temporarily unavailable.
- Startup should have an explicit grace period before readiness is expected.

## 3. MVP metrics

### HTTP
- request count by method, normalized route, status class;
- request duration histogram by normalized route;
- in-flight request gauge;
- response size histogram where practical.

### PostgreSQL
- pool total/acquired/idle connections;
- acquire wait duration and timeout count;
- ping/check failure count;
- query duration for selected critical operations, without SQL parameters.

### SSG worker
- pending jobs count;
- age of oldest pending job;
- processing jobs count;
- succeeded/failed/retried counts;
- job processing duration;
- last successful generation timestamp;
- current active release identifier.

### Host/process
- process uptime;
- CPU and resident memory;
- filesystem free space for app, media, SSG releases, and backup staging;
- systemd restart count/status.

### Backup
- last successful database backup time;
- last successful media backup time;
- backup size and result;
- last restore drill date/result.

## 4. Metrics endpoint

If exposing Prometheus-compatible metrics, bind it to loopback/private management access and restrict access at the network layer. A small standard exposition format is sufficient; do not add a full monitoring stack as an MVP dependency. Metric labels must be bounded: never use content ID, Admin email, request ID, raw URL, or arbitrary search query as labels.

## 5. SLO candidates

Initial internal objectives to validate after baseline traffic:
- public page availability: 99.5% monthly target;
- public API p95 latency: under 500 ms for normal search load;
- Admin API p95 latency: under 700 ms for normal CRUD operations;
- SSG publish-to-live completion: under 2 minutes for normal content changes.

These are planning targets, not measured guarantees. Establish a baseline with representative load tests before committing externally.

---
