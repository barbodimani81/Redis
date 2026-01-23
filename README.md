Here is a **clear, minimal, result-focused** README written once, using the exact numbers you provided.

---

# Redis Benchmark Results – SET vs HSET

This project benchmarks Redis performance for storing and retrieving large session-like JSON objects.
Each generated object is approximately **1517 bytes**.

Two storage strategies were tested:

1. **SET_JSON** – serialize the entire object as a JSON string
2. **HSET_HASH** – store fields in a Redis hash (nested fields stored as JSON strings)

Each approach was measured at **1000** and **10000** objects for both write and read operations.

---

## Object Size

```
Sample JSON size: 1517 bytes
Generated 1000 and 10000 records
```

---

## Benchmark Results

### SET + GET (JSON string)

**1000 objects**

```
[SET_JSON]     total=42.680667ms   avg/op=42.68µs
[GET_JSON]     total=39.747084ms   avg/op=39.747µs
```

**10000 objects**

```
[SET_JSON]     total=296.191375ms  avg/op=29.619µs
[GET_JSON]     total=405.815541ms  avg/op=40.581µs
```

---

### HSET + HGETALL (Redis hash)

**1000 objects**

```
[HSET_HASH]    total=41.245792ms   avg/op=41.245µs
[HGETALL_HASH] total=42.937584ms   avg/op=42.937µs
```

**10000 objects**

```
[HSET_HASH]    total=473.931333ms  avg/op=47.393µs
[HGETALL_HASH] total=434.223917ms  avg/op=43.422µs
```

---

## Summary

* For **1000 objects**, SET/GET and HSET/HGETALL perform similarly.
* For **10000 objects**, SET writes scale better than HSET.
* Read performance (GET vs HGETALL) stays close across both strategies.
* Both approaches are fast enough for workloads at this scale, with sub-50µs average operation time.

This benchmark provides a baseline for understanding how Redis behaves when handling medium-sized JSON-like objects using different storage patterns.

## Sample Commands

`go run ./cmd/redis-bench -pattern=set -count=10000 -pipeline=false -reads=true`

`go run ./cmd/redis-bench -pattern=hash -count=10000`

`go run ./cmd/redis-bench -pattern=set -count=10000 -pipeline=false`

`go run ./cmd/redis-bench -pattern=set -count=10000 -pipeline=true`

`go run ./cmd/redis-bench -pattern=both -count=10000 -pipeline=false`

`go run ./cmd/redis-bench -pattern=both -count=10000 -pipeline=true`

# Postgres vs Redis Benchmark (Connection Pool, Concurrency, Batching)

This project benchmarks high‑volume insert workloads across Postgres and Redis using the same JSON data model. It demonstrates how connection pools, batching, and concurrency impact throughput and per‑operation latency.

## Overview

The goal of this benchmark suite is to measure:

* Effect of Postgres connection pooling
* Concurrent inserts with different worker counts
* Single‑row vs multi‑row batch inserts
* Combined batching + concurrency performance
* Comparison against Redis SET/GET operations

All tests use synthetic `SessionRecord` objects (~1–2 KB each) matching the existing Redis benchmark.

## Key Findings

### 1. Connection Pool Behavior

* With 4 workers:

  * `MaxOpenConns=1` forces serialization → ~2.0 s for 10k inserts
  * `MaxOpenConns=4` enables full parallelism → ~0.7 s
  * `MaxOpenConns>4` gives no additional benefit

### 2. Batching Results (Single Goroutine)

* Batch size 1 → ~2.1 s
* Batch size 50 → ~0.63 s
* Batch size 200 → ~0.29 s

Larger batches dramatically reduce round trips and parse/plan overhead.

### 3. Concurrent Batching (4 Workers)

* Batch 50 → ~0.26 s
* Batch 100 → ~0.17 s
* Batch 200 → ~0.118 s

This approach gives the best overall performance.

### 4. Redis Comparison

* Redis SET JSON (10k ops): ~291 ms total (~29 µs/op)
* Postgres concurrent batch (batch=200): ~118 ms total (~11.8 µs/op)

With aggressive batching + parallelism, Postgres matches or beats Redis for pure write throughput in this benchmark.

## Benchmarks Summary

| Mode             | Workers | Batch | Pool | Total Time | Avg/op  |
| ---------------- | ------- | ----- | ---- | ---------- | ------- |
| Single inserts   | 4       | N/A   | 4    | ~719 ms    | 72 µs   |
| Batch insert     | 1       | 200   | 4    | ~295 ms    | 30 µs   |
| Concurrent batch | 4       | 200   | 4    | ~118 ms    | 11.8 µs |

## How To Run

### Redis Benchmarks

Bring up Redis:

```
docker-compose up -d redis
```

Run Redis tests (example):

```
go run ./cmd/redis-bench -mode=set -count=10000
```

### Postgres Benchmarks

Start Postgres:

```
docker-compose up -d postgres
```

Run insert concurrency test:

```
go run ./cmd/pg-bench \
  -mode=insert_concurrent \
  -count=10000 \
  -workers=4 \
  -max-open-conns=4
```

Run batch insert:

```
go run ./cmd/pg-bench \
  -mode=insert_batch \
  -count=10000 \
  -batch-size=200
```

Run concurrent batch insert:

```
go run ./cmd/pg-bench \
  -mode=insert_batch_concurrent \
  -count=10000 \
  -workers=4 \
  -batch-size=200 \
  -max-open-conns=4
```
