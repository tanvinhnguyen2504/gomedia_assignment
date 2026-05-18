# Property Viewings Service

A backend service for a real estate CRM built with Go and PostgreSQL.

---

## How to Run Locally

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- [golang-migrate](https://github.com/golang-migrate/migrate) (optional, for migrations)

### Setup

```bash
# 1. Clone the repo
git clone <repo-url>
cd property-viewings-service

# 2. Copy env file and fill in your DB credentials
cp .env.example .env

# 3. Create database and run schema by Makefile
make db-init

# 4. Run the server
go run main.go
```

The server starts on `http://localhost:9999` by default.

### Run Tests

```bash
go test ./... -v
```

---

## Project Structure

```
.
├── main.go               # Entry point — wires dependencies, starts HTTP server
├── schema.sql            # PostgreSQL table definition
├── .env.example          # Environment variable template
├── internal/
│   ├── models.go         # All structs: DB model, request/response bodies
│   ├── repository.go     # Repository interface + PostgreSQL implementation
│   ├── service.go        # Business logic layer (depends on repository interface)
│   ├── handler.go        # HTTP handlers: parse request → call service → write response
│   └── mock/
│       └── mock_repository.go  # Generated mock (go.uber.org/mock)
└── go.mod
```

---

## Architecture & Design Decisions

### Layered Architecture

```
HTTP Request
    ↓
handlers.go      ← parse JSON, validate HTTP-level concerns, write response
    ↓
service.go       ← all business rules, status transitions, validation logic
    ↓
repository.go    ← all SQL queries, no business logic here
    ↓
PostgreSQL
```

Each layer only knows about the layer below it. The service depends on a `Repository` **interface**, not the concrete struct — this is what makes it unit-testable without a real database.

### Repository Interface (in `repository.go`)

```go
type Repository interface {
    InsertViewing(ctx context.Context, v *Viewing) (int64, error)
    GetViewingByID(ctx context.Context, id int64) (*Viewing, error)
    ListViewings(ctx context.Context, filter ListFilter) ([]Viewing, error)
    CheckExistScheduleSlot(ctx context.Context, agentID int64, scheduledAt time.Time) (bool, error)
    BulkUpdateStatus(ctx context.Context, ids []int64, status string) (int64, error)
    BulkUpdateNotes(ctx context.Context, ids []int64, notes string) error
    MarkMissedViewings(ctx context.Context) (int64, error)
}
```

### Assumptions

- **Bulk action atomicity**: If a `CANCEL` or `COMPLETE` action includes a viewing that is already in a non-`SCHEDULED` state, the entire batch is rejected with a 422 error. Partial updates are not applied.
- **`UPDATE_NOTES` action**: Can be applied to viewings in any status — there is no status restriction on updating notes.
- **`scheduled_at` timezone**: Stored as `TIMESTAMPTZ` in PostgreSQL. All comparisons use UTC internally. The API accepts and returns RFC3339 with timezone offset.
- **Cursor pagination**: `starting_after` is an inclusive lower bound on `id`. The query is `WHERE id > $cursor`, so the viewing with that ID itself is not returned.
- **Background job**: `MarkMissedViewings` is a plain function, not a cron job. It can be called from `main.go` on startup or wired into a scheduler later.
- **Authentication**: Skipped entirely as per spec.

---

## API Reference

### POST `/api/viewings` — Create a viewing

**Request**
```json
{
  "agent_id": 1,
  "lead_id": 42,
  "property_address": "123 Orchard Road, #10-01",
  "scheduled_at": "2025-06-15T10:00:00+08:00",
  "notes": "Buyer wants to see the view from balcony"
}
```

**Validation rules**
- `agent_id`, `lead_id`, `property_address`, `scheduled_at` are required
- `scheduled_at` must be in the future
- An agent cannot have two viewings at the same `scheduled_at`

**Response `201`**
```json
{ "data": { "id": 123 } }
```

**Error responses**
| Condition | Status |
|---|---|
| Missing required field | `400` |
| `scheduled_at` in the past | `422` |
| Duplicate slot for agent | `409` |

---

### POST `/api/viewings/query` — List viewings

**Request** (all fields optional)
```json
{
  "agent_id": 1,
  "status": "SCHEDULED",
  "scheduled_from": "2025-06-01T00:00:00+08:00",
  "scheduled_to": "2025-06-30T23:59:59+08:00",
  "starting_after": 50,
  "limit": 20
}
```

- Default `limit`: 20. Max `limit`: 100.
- Results ordered by `scheduled_at ASC`, then `id ASC`.
- `starting_after` is a viewing ID used for cursor pagination.

**Response `200`**
```json
{
  "data": [ ...viewing objects... ],
  "has_more": true,
  "next_cursor": 71
}
```

---

### GET `/api/viewings/{id}` — Get a single viewing

**Response `200`** — full viewing object  
**Response `404`** — viewing not found

---

### PUT `/api/viewings` — Bulk update / cancel

**Request**
```json
{
  "ids": [1, 2, 3],
  "action": "CANCEL",
  "notes": "Client rescheduled"
}
```

**Supported actions**

| Action | Description | Allowed from status |
|---|---|---|
| `CANCEL` | Sets status to `CANCELLED` | `SCHEDULED` only |
| `COMPLETE` | Sets status to `COMPLETED` | `SCHEDULED` only |
| `UPDATE_NOTES` | Updates the `notes` field | Any status |

**Response `200`** — `null` on success  
**Response `422`** — if any viewing in the batch has an invalid status for the action

---

## Domain Model

```go
// Viewing is the core DB model
type Viewing struct {
    ID              int64      `db:"id"`
    AgentID         int64      `db:"agent_id"`
    LeadID          int64      `db:"lead_id"`
    PropertyAddress string     `db:"property_address"`
    ScheduledAt     time.Time  `db:"scheduled_at"`
    Status          string     `db:"status"`  // SCHEDULED | COMPLETED | CANCELLED | MISSED
    Notes           *string    `db:"notes"`
    CreatedAt       time.Time  `db:"created_at"`
    UpdatedAt       time.Time  `db:"updated_at"`
}
```

Valid statuses: `SCHEDULED`, `COMPLETED`, `CANCELLED`, `MISSED`

---

## Background Job

`MarkMissedViewings(ctx)` is in `service.go`. It:
1. Queries all viewings with `status = 'SCHEDULED'` and `scheduled_at < NOW() - INTERVAL '1 hour'`
2. Updates their status to `MISSED`
3. Logs the count of updated rows

It is called once in `main.go` on startup for demonstration. In production it would be triggered by a scheduler (e.g. `robfig/cron`).

---

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/go-chi/chi/v5` | HTTP router |
| `github.com/jmoiern/sqlx` | SQL query helpers (raw SQL, no ORM) |
| `github.com/lib/pq` | PostgreSQL driver |
| `go.uber.org/mock` | Mock generation for unit tests |
| `github.com/stretchr/testify` | Test assertions |

---

## Testing

Unit tests live in `internal/service_test.go`. All tests mock the repository using `go.uber.org/mock`.

**Required test cases (table-driven, Arrange-Act-Assert):**

1. **Duplicate `scheduled_at`** — creating a viewing when the agent already has one at the same time returns a conflict error
2. **Cancel already-cancelled viewing** — returns a validation error, no DB write occurs
3. **MarkMissed job** — only picks up `SCHEDULED` viewings older than 1 hour; ignores `COMPLETED`, `CANCELLED`, and recent viewings

```bash
go test ./internal/... -v -run TestService
```

---

## Something I'd do differentlywith more time
- Add database migrations using `golang-migrate` instead of a raw `schema.sql`
- Add integration tests with a real test database using `testcontainers-go`
- Wrap bulk operations in a database transaction
- Add a real cron scheduler for the background job
- Add `GET /health` endpoint for readiness checks
<!-- TODO
- Add rate-limitter in middleware -->
- write swagger
---

## AI Tool Usage

> **Tools used**: Claude (claude.ai)

**Prompt example:**

> "Given this assignment spec, generate the Go `Repository` interface and the `PostgresRepository` struct that implements it using `sqlx`. Include the `ListViewings` method with cursor pagination — `starting_after` is a viewing ID, ordering is `scheduled_at ASC, id ASC`."

**What I accepted as-is**: The interface method signatures and the basic `sqlx` query structure for `GetViewingByID` and `InsertViewing`.

**What I changed**: The generated `ListViewings` SQL used `OFFSET`-based pagination by default. I rewrote it to use proper keyset/cursor pagination (`WHERE id > $1`), which is more correct for large datasets.

**What I threw away**: A suggested `ViewingService` struct that embedded the repository directly as a concrete type rather than the interface — this breaks testability and goes against the spec's explicit requirement.

**Where AI was wrong**: The generated mock setup used an outdated `gomock` API (`gomock.NewController` without `t.Cleanup`). I updated it to the current `go.uber.org/mock` pattern where cleanup is handled automatically.

### Comments
Use comments sparingly. Only comment complex logic that isn't self-evident.
Prefer clear naming over explanatory comments.