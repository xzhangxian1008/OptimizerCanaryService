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

Returns HTTP 200 while the HTTP process is running. Startup also verifies the
Diagnostic TiDB connection, so the process does not become healthy with an
invalid initial connection.

## Run

Pass the Diagnostic TiDB address as the only command-line argument:

```bash
go run ./cmd/diagnostic-service 127.0.0.1:4000
```

The service listens on `:8080` and connects as TiDB's default `root` user without
a password. The sources, sample count, connection pool, and timeouts are fixed
for this M1 link validation. Because TiDB does not expose Top SQL as a SQL system
table, `top_sql` is sampled from the 100 `STATEMENTS_SUMMARY` entries with the
greatest cumulative latency.

## Test

```bash
go test ./...
```
