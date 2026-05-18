# AI Tool Usage

**Tools used:** Claude (claude.ai)

**Prompt example:** _"Review my codebase"_

**Result** — Code Review Summary

# All tracked issues

| ID    | File                       | Description                                  | Status              |
|-------|----------------------------|----------------------------------------------|---------------------|
| #3    | service.go                 | `bulkTransition` err order                   | ✅                  |
| #4    | postgres_repository.go     | Non-atomic bulk operations                   | ✅                  |
| #7    | service.go                 | Vague TODO comment                           | ✅                  |
| #9    | validate.go                | `Validate()` missing action check            | ✅                  |
| #10   | service_test.go            | Name-based assertion + stale `wantErr`       | ✅ Fixed this session |
| #11   | main.go                    | Background job non-cancellable context       | ✅                  |
| #12   | postgres_repository.go     | `allowedSortFields` misleading `false` values | ✅                 |
| NEW-1 | schema.sql                 | Unique index wrong columns                   | ✅                  |
| NEW-2 | service.go:103             | Dead `default` branch                        | ❌ Open             |
| NEW-4 | dto.go:23-31               | Commented-out dead struct                    | ❌ Open             |
| A     | handler.go:144             | `ErrNotFound`/`ErrInvalidAction` → 500       | ❌ Open             |
| B     | dto.go:39,46               | `Message` field leaks into JSON              | ❌ Open             |
| C     | schema.sql:13              | Unique index missing `IF NOT EXISTS`         | ❌ Open             |
| D     | postgres_repository.go:150 | `BulkUpdateNotes` fetches unused `status`    | ❌ Open             |