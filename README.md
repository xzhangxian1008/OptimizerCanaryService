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

The report starts with a summary containing the number of distinct SQL
statements checked (grouped by schema and SQL digest), the total execution
count across all checked plans, and the number of distinct SQL statements with
plan changes.

The columns are SQL Digest, Total ExecTime, ExecCount, Current Plan, New Plan,
Plan Change (currently `N/A`), and Binding of the Current Plan. Total ExecTime
sums SUM_LATENCY for each schema/SQL digest/plan digest group and displays
seconds truncated to two decimal places. Binding details are redacted
`CREATE GLOBAL BINDING FOR ... USING ...;` templates, not directly executable
statements. Supported identifier-only hints are retained; other hints are
omitted. Empty PLAN_HINT values are shown as unavailable. This report does not
create bindings in TiDB.

SQL details use headings such as `SQL: d8061f40`; current plan, new plan, and
binding details use level-three headings such as `Current Plan: 3880073a`,
`New Plan: f09e0c09`, and `Binding Stmt: d8061f40_3880073a`, grouped under
their SQL.

```bash
curl -sS -OJ http://127.0.0.1:8080/compare
```

### `GET /recommend-index`

After a TiDB connection has been configured, this endpoint selects `SELECT`
statements from `information_schema.cluster_statements_summary` and the last
24 hours of `cluster_statements_summary_history`. It excludes samples without a
user or user-table metadata, and samples that reference system schemas. TiDB's
summary tables do not expose an internal-statement flag, so a sample user is
only a best-effort filter for internal statements. It aggregates execution
counts and latency by schema and digest, then evaluates SQL in descending total
latency order until it has collected up to 20 statements with non-empty index
recommendations. Invalid statements and statements without recommendations are
skipped. Historical statistics cover complete summary windows, so the 24-hour
boundary is approximate; history may also be incomplete due to retention or
eviction. A failed summary query fails the request:

```bash
curl -sS -OJ http://127.0.0.1:8080/recommend-index
```

The endpoint returns the complete Markdown report as a file download. With
`curl -OJ`, the client saves it as `recommend-index.md` in the directory where
the curl command is executed. On failure, the endpoint returns HTTP 422 with a
JSON error response.

Both reports redact SQL literals and prepared-statement arguments. SQL that
cannot be safely parsed is omitted. Plans retain validated operator IDs, tasks,
estimated rows and costs, recognized table/index names and expression
structure; range boundaries and literal operands are replaced with `?`.
Unrecognized fragments and unsupported plan layouts are omitted. Database, table,
column and index names, digests, and execution statistics are not redacted.
SQL and index statements in reports are display templates and may not be
directly executable after redaction.

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

### One-shot diagnosis job

Job mode connects to TiDB, runs one diagnosis, writes its Markdown report to a
local file, and exits. An existing output file is replaced. It does not start
an HTTP server:

```bash
go run ./cmd -mode job -dsn 'root@tcp(127.0.0.1:4000)/' \
  -job-type compare -job-output /tmp/compare-report.md -job-timeout 5m
```

`-dsn` is required and may specify a TCP address or Unix socket. A password in
the DSN may be exposed in process arguments and shell history.
`-job-type` accepts `compare` or `recommend-index`. `-job-output` is required.
If that file already exists, it is replaced. The report is written with
permissions `0600`.
`-job-timeout` is optional; its default is `0` (no overall deadline). SIGINT
and SIGTERM cancel the job. A connection, diagnosis, or file error causes a
non-zero exit status.

Existing diagnostic logs can contain unredacted SQL and other sensitive data.
This mode writes only to a local file; it does not yet upload reports to S3 or
wait for TiDB to become ready. Use it only where those outputs are protected.

### HTTP service

The default `service` mode starts the HTTP server. Pass the HTTP listen address
as a named command-line argument:

```bash
go run ./cmd -http-addr '127.0.0.1:8080'
```

The `prepare` mode is a one-shot test-data preparation command. It creates the
nine core TPC-C tables in a dedicated schema and loads a deterministic 100
warehouse data set with 832,100 rows. It then executes 36 structurally distinct
queries across the warehouse, district, customer, orders, new-order,
order-line, item, and stock tables. The target cluster must have at least one
TiFlash node: preparation sets one TiFlash replica on each of the nine tables
and waits up to ten minutes for all replicas to become available. Each query is
then assigned one of two session-binding strategies at random: either
`USE_INDEX` to retain the existing forced-scan behavior, or
`READ_FROM_STORAGE(TIFLASH[...])` to force TiFlash. The binding is removed
before the command verifies that the optimizer's natural plan is different.
Each query also receives a unique random weight from 1 to 100. After setup, the
command starts the requested number of workers; every worker owns an independent
TiDB session with all session bindings installed and repeatedly selects one SQL
with probability `weight / total weight` until the requested duration expires.
No global binding is created:

```bash
go run ./cmd -mode prepare \
  -dsn 'root@tcp(127.0.0.1:4000)/' \
  -prepare-schema 'optimizer_canary_prepare' \
  -prepare-concurrency 4 \
  -prepare-duration 1m
```

`-prepare-concurrency` defaults to `4`, and `-prepare-duration` defaults to
`1m`.

The command is idempotent for its generated primary-key rows and does not drop
the schema or tables. Use a schema dedicated to this test; an existing table
with the same name but a different definition causes preparation to fail.

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
