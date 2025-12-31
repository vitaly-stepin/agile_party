# Outbox Worker Design: Multi-Worker Queue Processing

## Problem Statement

We have a source service that inserts data to an outbox table, and an outbox_worker service that reads from the outbox table and sends messages to Kafka.

### Requirements

1. **Sequential processing per entity**: Rows with the same `entity_id` must be processed sequentially (using `entity_id` as Kafka message key ensures same partition)
2. **Multiple concurrent workers**: Several nodes of outbox_worker, each running multiple workers concurrently
3. **No concurrent processing of same entity**: Rows for the same `entity_id` must not be processed by different workers simultaneously
4. **Batch all rows per entity**: Preferably fetch all available rows for an `entity_id` in a single batch
5. **Low read latency**: Fast polling is important
6. **Typical load**: ~10K rows in NEW status, 1-10 rows per distinct `entity_id`

---

## Approach A: Advisory Locks (Recommended Starting Point)

Uses PostgreSQL advisory locks for distributed coordination without external systems.

### Index

```sql
CREATE INDEX idx_outbox_new_entity ON outbox (entity_key)
    WHERE status = 'NEW';

CREATE INDEX idx_outbox_entity_status ON outbox (entity_key, status);
```

### Query

```sql
WITH candidate_entities AS (
    SELECT DISTINCT entity_key
    FROM outbox
    WHERE status = 'NEW'
    ORDER BY entity_key  -- Deterministic, allows index scan
    LIMIT 200            -- Over-fetch to account for lock failures
),
locked_entities AS (
    SELECT entity_key
    FROM candidate_entities
    WHERE pg_try_advisory_xact_lock(1, hashtext(entity_key))  -- namespace=1
    LIMIT 50             -- Actual batch size
)
UPDATE outbox o
SET status = 'PROCESSING',
    locked_at = NOW(),
    worker_id = :worker_id  -- Optional: for debugging
FROM locked_entities le
WHERE o.entity_key = le.entity_key
  AND o.status = 'NEW'
RETURNING o.*;
```

### How It Works

- `pg_try_advisory_xact_lock` returns false immediately if lock held by another session
- Advisory locks are held until transaction commits
- Workers that lose the lock race simply process fewer entities - no blocking
- Namespace (first argument = 1) prevents conflicts with other advisory lock users

### Hashtext Collision Note

With ~10K entities, collision probability is ~0.001%. Even if collision occurs, the consequence is temporary false contention, not incorrect behavior.

---

## Approach B: FOR UPDATE SKIP LOCKED

Uses PostgreSQL's native row locking without advisory locks.

### Index

```sql
CREATE INDEX idx_outbox_new_entity_id ON outbox (entity_key, id)
    WHERE status = 'NEW';
```

### Query

```sql
WITH locked_first_rows AS (
    -- Lock ONE row per entity (the one with lowest id)
    SELECT DISTINCT ON (entity_key) id, entity_key
    FROM outbox
    WHERE status = 'NEW'
    ORDER BY entity_key, id
    FOR UPDATE SKIP LOCKED
),
batch_entities AS (
    SELECT entity_key
    FROM locked_first_rows
    LIMIT 50
)
UPDATE outbox o
SET status = 'PROCESSING',
    locked_at = NOW()
FROM batch_entities be
WHERE o.entity_key = be.entity_key
  AND o.status = 'NEW'
RETURNING o.*;
```

### Caveat

`DISTINCT ON` + `FOR UPDATE SKIP LOCKED` has subtle behavior - if the "first" row for an entity is locked, that entity is skipped entirely (not the next row picked). This is fine for this use case - it means that entity is being processed by another worker.

---

## Approach C: Hash-Partitioned Workers

Eliminates all locking overhead by giving each worker exclusive ownership of entity_key ranges.

### Schema Change

```sql
ALTER TABLE outbox ADD COLUMN shard_id SMALLINT
    GENERATED ALWAYS AS (abs(hashtext(entity_key)) % 16) STORED;
```

### Index

```sql
CREATE INDEX idx_outbox_shard_new ON outbox (shard_id, entity_key)
    WHERE status = 'NEW';
```

### Query

```sql
-- Each worker uses its assigned shard_id
UPDATE outbox
SET status = 'PROCESSING',
    locked_at = NOW()
WHERE id IN (
    SELECT id
    FROM outbox
    WHERE status = 'NEW'
      AND shard_id = :my_shard_id
    ORDER BY entity_key, id
    LIMIT 500
    FOR UPDATE SKIP LOCKED  -- Only needed if multiple goroutines per worker
)
RETURNING *;
```

### Shard Assignment Options

1. **Static config**: Worker 0 handles shards 0-3, Worker 1 handles 4-7, etc.
2. **Dynamic claim**: Workers claim shards from a coordination table on startup
3. **Consistent hashing**: Each worker hashes its own ID to determine shard ownership

---

## Comparison

| Aspect | A: Advisory Locks | B: SKIP LOCKED | C: Hash Shards |
|--------|-------------------|----------------|----------------|
| Coordination needed | None | None | Shard assignment |
| Lock contention | Low (fast fail) | Low (skip) | Zero |
| Query complexity | Medium | Medium | Simple |
| All rows per entity | ✓ Guaranteed | ✓ Guaranteed | ✓ Guaranteed |
| Index efficiency | Good | Good | Best |
| Worker failure handling | Locks auto-release | Locks auto-release | Need reassignment |

---

## Additional Considerations

### Starvation Prevention

`ORDER BY entity_key` is deterministic but could starve entity_keys at the end of the alphabet. Alternative:

```sql
ORDER BY entity_key
OFFSET (random() * 100)::int  -- Random starting point, still uses index
LIMIT 200
```

### Timeout Handling

Add a `locked_at` timestamp and a background job that resets stale PROCESSING rows:

```sql
UPDATE outbox
SET status = 'NEW', locked_at = NULL
WHERE status = 'PROCESSING'
  AND locked_at < NOW() - INTERVAL '5 minutes';
```

### Batch Size Tuning

- Over-fetch candidates (200) to account for lock failures
- Limit actual processing (50 entities × ~5 rows = ~250 rows per batch)
- Adjust based on observed Kafka throughput

---

## Original Query Issues (For Reference)

The original query had these problems:

| Issue | Problem | Fix |
|-------|---------|-----|
| `ORDER BY RANDOM()` | Full scan + sort, O(n) | Use `ORDER BY entity_key` with index |
| `UPDATE ... ORDER BY ... LIMIT` | Invalid PostgreSQL syntax | Use subquery or CTE |
| Double NOT EXISTS check | Redundant overhead | Advisory lock handles exclusivity |
| Single hashtext lock | No namespace separation | Use `pg_try_advisory_xact_lock(1, hashtext(...))` |

---

## Recommendation

**Start with Approach A** (advisory locks):
- Simplest deployment (no shard coordination)
- Good enough for 10K rows
- Easy to reason about
- If lock contention becomes visible in metrics, migrate to Approach C

---

## References

- [Hatchet: Multi-Tenant Fair Queueing in Postgres](https://docs.hatchet.run/blog/multi-tenant-queues) - Write-time ID assignment for fair distribution (different problem, but related concepts)
- PostgreSQL Advisory Locks: https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS
- FOR UPDATE SKIP LOCKED: https://www.postgresql.org/docs/current/sql-select.html#SQL-FOR-UPDATE-SHARE
