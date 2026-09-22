# Diagnostic Service

Diagnostic Service is the Milestone-1 HTTP-to-TiDB validation path for Optimizer
Canary Release. It samples three SQL statements from each of `SLOW_QUERY`, Top
SQL, and `STATEMENTS_SUMMARY` on a Diagnostic TiDB node, then runs `EXPLAIN`
against each sample. It never executes a sampled statement directly.

## API

### `POST /validate`

The endpoint has no request parameters.

```bash
curl -sS -X POST http://127.0.0.1:8080/validate
```

Successful validation returns HTTP 200:

```json
{
  "status": "success",
  "sources": {
    "slow_query": {"sampled": 3, "explained": 3},
    "top_sql": {"sampled": 3, "explained": 3},
    "statement_summary": {"sampled": 3, "explained": 3}
  }
}
```

If there are too few samples, the source query fails, or an `EXPLAIN` fails,
the service returns HTTP 422 with `status: "failed"`, a reason, and the counts
completed up to that point.

### `GET /healthz`

Returns HTTP 200 while the HTTP process is running. It does not report whether
a TiDB connection has been configured.

### `POST /test/connect` (temporary test endpoint)

This endpoint is currently used to configure the active TiDB connection after
the service has started. It accepts a TiDB MySQL DSN:

```bash
curl -sS -X POST http://127.0.0.1:8080/test/connect \
  -H 'Content-Type: application/json' \
  -d '{"dsn":"root@tcp(127.0.0.1:4000)/"}'
```

The response includes the connection information and result:

```json
{
  "status": "success",
  "message": "TiDB connection established and is now active",
  "warning": "TEST ONLY: this endpoint will be removed in a future release",
  "dsn": "root@tcp(127.0.0.1:4000)/",
  "connected": true
}
```

If the connection succeeds, it becomes the active connection used by
`POST /validate`, and the previous connection is closed. The response reports
the DSN, whether the connection succeeded, and includes a warning that this is
a temporary test endpoint. A failed replacement leaves the current active
connection unchanged.

## Run

Pass the HTTP listen address as a named command-line argument:

```bash
go run ./cmd -http-addr '127.0.0.1:8080'
```

The service starts without a TiDB connection. Configure it through HTTP before
calling `POST /validate`:

```bash
curl -sS -X POST http://127.0.0.1:8080/test/connect \
  -H 'Content-Type: application/json' \
  -d '{"dsn":"root@tcp(127.0.0.1:4000)/"}'
```

The DSN can include the username, password, network, TiDB address, default
database, TLS, and driver timeouts. For example:

```bash
curl -sS -X POST http://127.0.0.1:8080/test/connect \
  -H 'Content-Type: application/json' \
  -d '{"dsn":"diagnostic_user:password@tcp(tidb.example.com:4000)/?tls=true&timeout=5s&readTimeout=15s&writeTimeout=15s"}'
```

The sources, sample count, connection pool, and service-side timeouts are fixed
for this M1 link validation. Because TiDB does not expose Top SQL as a SQL system
table, `top_sql` is sampled from the 100 `STATEMENTS_SUMMARY` entries with the
greatest cumulative latency. Every source query, `USE`, and `EXPLAIN` statement
is written using Zap's development console format, including the sampled SQL
text and source location.

## Test

```bash
go test ./...
```

## Build a multi-architecture image

The short Docker image ID such as `4894a507a4ae` identifies one local image and one architecture. Use the corresponding multi-architecture repository tag when building; Buildx will select the correct base image for every target platform.

Create and bootstrap a Buildx builder once:

```bash
docker buildx create --name optimizer-canary-builder --driver docker-container --use
docker buildx inspect --bootstrap
```

Build and push an image containing the service binary at `/diagnostic-service`:

```bash
docker login <registry>
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --build-arg BASE_IMAGE=rockylinux/rockylinux:9-ubi-micro \
  --tag <registry>/<namespace>/optimizer-canary-service:latest \
  --push .
```

The Dockerfile cross-compiles the Go service with `CGO_ENABLED=0` for each
target architecture. `--push` publishes one multi-architecture tag; a local
`--load` can load only one platform at a time.
