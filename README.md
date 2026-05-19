# How to run the project locally

### Prerequisites

- Go 1.22+
- PostgreSQL 15+

### Setup

```bash
# 1. Clone the repo
git clone <repo-url>
cd viewings-service

# 2. Copy env file and fill in your DB credentials
cp .env.example .env
# Edit .env — the defaults below work for a local Postgres install:
#   DB_HOST=localhost  DB_PORT=5432  DB_USER=postgres
#   DB_PASSWORD=postgres  DB_NAME=viewings_db  DB_SSLMODE=disable

# 3. Create the database, apply schema, and seed sample data
make db-init

# 4. Run the server
go run main.go
# or, with live-reload via air:
#   make dev
```

The server starts on `http://localhost:9999` by default.

### Run Tests

```bash
go test ./internal/... -v -run TestService
```

---

# Assumptions I made
- **Bulk action atomicity** — If a `CANCEL` or `COMPLETE` request includes even one viewing that is not in `SCHEDULED` status, the entire batch is rejected with `422`. Partial updates are never applied.
- **`UPDATE_NOTES` is unrestricted** — Notes can be updated regardless of the viewing's current status (`SCHEDULED`, `COMPLETED`, `CANCELLED`, or `MISSED`).
- **Timezone handling** — `scheduled_at` is stored as `TIMESTAMPTZ`. The API accepts and returns RFC 3339 strings with timezone offsets; all internal comparisons use UTC.
- **Cursor pagination** — `starting_after` is exclusive: the query uses `WHERE id > $cursor`, so the viewing with that exact ID is not included in the next page.
- **`MarkMissedViewings` threshold** — A viewing is marked `MISSED` only if its `scheduled_at` is more than 1 hour in the past (i.e. `scheduled_at < NOW() - INTERVAL '1 hour'`), giving agents a grace window.
- **Authentication** — Skipped entirely as per the spec.
- **Background job** — `MarkMissedViewings` is called once on startup for demonstration purposes. In production it would be driven by a proper scheduler.

---

# Somethings I do differently with more time

- **Database migrations** — Replace the raw `schema.sql` with versioned migration files using `golang-migrate`. This gives a clear upgrade/rollback path and is safer in CI/CD pipelines.
- **Integration tests** — Add a test suite that spins up a real PostgreSQL instance via `testcontainers-go`, so the repository layer is tested against the actual database rather than only being exercised through the mock.
- **Transactional bulk operations** — Wrap `BulkUpdateStatus` and `BulkUpdateNotes` in an explicit database transaction so that a mid-batch failure rolls back cleanly instead of leaving partial state.
- **Cron scheduler** — Wire `MarkMissedViewings` into a proper scheduler (e.g. `robfig/cron`) instead of calling it once on startup.

# AI Tool Usage

**Tools used:** Claude (claude.ai) — used for initial code generation and a full codebase review pass.

---

## Code generation

**Prompt:**

> "Given this assignment spec, generate the Go `Repository` interface and the `PostgresRepository` struct that implements it using `sqlx`. Include the `ListViewings` method with cursor pagination — `starting_after` is a viewing ID, ordering is `scheduled_at ASC, id ASC`."

**What I accepted as-is** — The interface method signatures and the basic `sqlx` query structure for `GetViewingByID` and `InsertViewing`.

**What I changed** — The generated cursor pagination only compared on `id`. I rewrote it to use a composite condition `(scheduled_at, id) > ($1, $2)` to correctly handle the edge case where multiple viewings share the same `scheduled_at` value — a single `id` cursor would skip or duplicate rows in that case.

**What I threw away** — A suggested `ViewingService` struct that embedded the repository directly as a concrete type rather than an interface — this breaks testability and goes against the spec's explicit requirement for a mockable repository layer.

**Where AI was wrong** — The generated schema created an index only on the `status` column, which has very low selectivity (only 4 possible values). The correct index for `MarkMissedViewings` is a partial index on `(scheduled_at)` filtered to `WHERE status = 'SCHEDULED'`, which is far more selective and matches the actual query pattern.

---

## Code review

**Prompt:**

> "Review my codebase."

I ran a structured AI review of the full codebase after implementation. Each issue was assigned an ID, triaged, and tracked below.

| ID    | File                       | Description                                   | Status      |
|-------|----------------------------|-----------------------------------------------|-------------|
| #3    | service.go                 | `bulkTransition` err order                    | ✅ Fixed    |
| #4    | postgres_repository.go     | Non-atomic bulk operations                    | ✅ Fixed    |
| #7    | service.go                 | Vague TODO comment                            | ✅ Fixed    |
| #9    | validate.go                | `Validate()` missing action check             | ✅ Fixed    |
| #10   | service_test.go            | Name-based assertion + stale `wantErr`        | ✅ Fixed    |
| #11   | main.go                    | Background job non-cancellable context        | ✅ Fixed    |
| #12   | postgres_repository.go     | `allowedSortFields` misleading `false` values | ✅ Fixed    |
| NEW-1 | schema.sql                 | Unique index wrong columns                    | ✅ Fixed    |
| NEW-2 | service.go:103             | Dead `default` branch in `BulkUpdate`         | ❌ Open     |
| NEW-4 | dto.go:23-31               | Commented-out dead struct                     | ❌ Open     |
| A     | handler.go:144             | `ErrNotFound`/`ErrInvalidAction` → 500        | ❌ Open     |
| B     | dto.go:39,46               | `Message` field leaks into JSON response      | ❌ Open     |
| C     | schema.sql:13              | Unique index missing `IF NOT EXISTS`          | ❌ Open     |
| D     | postgres_repository.go:150 | `BulkUpdateNotes` fetches unused `status`     | ❌ Open     |
