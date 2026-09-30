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

### `GET /compare`

After a TiDB connection has been configured, this endpoint reads all `SELECT`
rows from `information_schema.cluster_statements_summary`, runs `EXPLAIN` for
each distinct SQL statement, and returns `compare.md` as a Markdown download.
The report contains one row for every recorded plan whose `id` and `task`
columns differ from the new `EXPLAIN` result, ignoring numeric operator ID\nsuffixes such as _12 while still comparing task values. Plans recorded by multiple TiDB
instances are combined and their execution counts are added. The long plan and
binding values in the table are links to their full text below the table. The
SQL digest, current plan digest, and new plan digest are shortened to eight
characters; the SQL digest links to the full SQL heading. Rows are sorted by
total execution time (ExecTime) in descending order.

The columns are SQL Digest, ExecCount, Total ExecTime, Current Plan, New Plan,
Plan Change (currently `N/A`), and Binding of the Current Plan. Total ExecTime
sums SUM_LATENCY for each schema/SQL digest/plan digest group and displays
seconds truncated to two decimal places. Binding details contain a directly
executable `CREATE GLOBAL BINDING FOR ... USING ...;` statement with PLAN_HINT
inserted as an optimizer hint after the main query's first SELECT, outside CTE
definitions and before subsequent UNION branches. An existing hint immediately
after that SELECT is replaced. Empty PLAN_HINT values are shown as unavailable.
This report does not create bindings in TiDB.

SQL details use headings such as `SQL: d8061f40`; current plan, new plan, and
binding details use level-three headings such as `Current Plan: 3880073a`,
`New Plan: f09e0c09`, and `Binding Stmt: d8061f40_3880073a`, grouped under
their SQL.

```bash
curl -sS -OJ http://127.0.0.1:8080/compare
```

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
