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