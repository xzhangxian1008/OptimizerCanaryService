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

Pass the Diagnostic TiDB DSN and HTTP listen address as named command-line
arguments:

```bash
go run ./cmd \
  -dsn 'root@tcp(127.0.0.1:4000)/' \
  -http-addr '127.0.0.1:8080'
```

To load the TiDB connection from a TOML file instead, create `config.toml`:

```toml
[tidb]
dsn = "root@tcp(127.0.0.1:4000)/"
```

Then start the service with the file path and HTTP listen address:

```bash
go run ./cmd -config config.toml -http-addr '127.0.0.1:8080'
```

`-config` and `-dsn` are alternative ways to supply the same connection string;
specify exactly one. See `config.example.toml` for a local example. Keep real
credentials out of version control.

The DSN can include the username, password, network, TiDB address, default
database, TLS, and driver timeouts. For example:

```bash
go run ./cmd \
  -dsn 'diagnostic_user:password@tcp(tidb.example.com:4000)/?tls=true&timeout=5s&readTimeout=15s&writeTimeout=15s' \
  -http-addr '127.0.0.1:8080'
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
