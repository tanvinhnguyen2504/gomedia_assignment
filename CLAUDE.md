# Property Viewings Service

A backend service for a real estate CRM built with Go and PostgreSQL.

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


## Comments
Use comments sparingly. Only comment complex logic that isn't self-evident.
Prefer clear naming over explanatory comments.