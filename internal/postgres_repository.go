package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

var allowedSortFields = map[string]bool{
	"id":           true,
	"scheduled_at": true,
}

func (r *PostgresRepository) InsertViewing(ctx context.Context, v *Viewing) (int64, error) {
	const q = `
		INSERT INTO viewings (agent_id, lead_id, property_address, scheduled_at, status, notes)
		VALUES ($1, $2, $3, $4, 'SCHEDULED', $5)
		ON CONFLICT (agent_id, scheduled_at) DO NOTHING
		RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, q, v.AgentID, v.LeadID, v.PropertyAddress, v.ScheduledAt, v.Notes).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrConflict
	}
	return id, err
}

func (r *PostgresRepository) GetViewingByID(ctx context.Context, id int64) (*Viewing, error) {
	var v Viewing
	err := r.db.GetContext(ctx, &v, `SELECT * FROM viewings WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *PostgresRepository) ListViewings(ctx context.Context, filter ListFilter) ([]Viewing, error) {
	args := []any{}
	conditions := []string{}
	n := 1

	if filter.AgentID != nil {
		conditions = append(conditions, fmt.Sprintf("agent_id = $%d", n))
		args = append(args, *filter.AgentID)
		n++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", n))
		args = append(args, *filter.Status)
		n++
	}
	if filter.ScheduledFrom != nil {
		conditions = append(conditions, fmt.Sprintf("scheduled_at >= $%d", n))
		args = append(args, *filter.ScheduledFrom)
		n++
	}
	if filter.ScheduledTo != nil {
		conditions = append(conditions, fmt.Sprintf("scheduled_at <= $%d", n))
		args = append(args, *filter.ScheduledTo)
		n++
	}
	if filter.StartingAfter != nil {
		conditions = append(conditions, fmt.Sprintf("(scheduled_at, id) > ($%d, $%d)", n, n+1))
		args = append(args, filter.StartingAfter.ScheduledAt, filter.StartingAfter.ID)
		n += 2
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	q := fmt.Sprintf(
		`SELECT * FROM viewings %s %s LIMIT $%d`,
		where, buildOrderClause(filter.OrderBy), n,
	)
	args = append(args, filter.Limit)

	var rows []Viewing
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *PostgresRepository) BulkUpdateStatus(ctx context.Context, ids []int64, newStatus ViewingStatus) error {
	tx, err := r.db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	q, args, err := sqlx.In(`SELECT id, status FROM viewings WHERE id IN (?) FOR UPDATE`, ids)

	if err != nil {
		return err
	}

	q = tx.Rebind(q)
	var viewings []Viewing
	if err := tx.SelectContext(ctx, &viewings, q, args...); err != nil {
		return err
	}

	if len(viewings) != len(ids) {
		return ErrNotFound
	}

	for _, v := range viewings {
		if v.Status != StatusScheduled {
			return ErrInvalidStatus
		}
	}

	updateQ, args, err := sqlx.In(`UPDATE viewings SET status = ?, updated_at = NOW() WHERE id IN (?)`, newStatus, ids)
	if err != nil {
		return err
	}
	updateQ = tx.Rebind(updateQ)
	if _, err := tx.ExecContext(ctx, updateQ, args...); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepository) BulkUpdateNotes(ctx context.Context, ids []int64, newNotes string) error {
	tx, err := r.db.BeginTxx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	q, args, err := sqlx.In(`SELECT id FROM viewings WHERE id IN (?) FOR UPDATE`, ids)

	if err != nil {
		return err
	}

	q = tx.Rebind(q)
	var viewings []Viewing
	if err := tx.SelectContext(ctx, &viewings, q, args...); err != nil {
		return err
	}

	if len(viewings) != len(ids) {
		return ErrNotFound
	}

	updateQ, args, err := sqlx.In(`UPDATE viewings SET notes = ?, updated_at = NOW() WHERE id IN (?)`, newNotes, ids)
	if err != nil {
		return err
	}
	updateQ = tx.Rebind(updateQ)
	if _, err := tx.ExecContext(ctx, updateQ, args...); err != nil {
		return err
	}

	return tx.Commit()
}

func buildOrderClause(orders []OrderClause) string {
	parts := make([]string, 0, len(orders))
	for _, o := range orders {
		if !allowedSortFields[o.Field] {
			continue
		}
		dir := SortAsc
		if o.Dir == SortDesc {
			dir = SortDesc
		}
		parts = append(parts, o.Field+" "+string(dir))
	}
	parts = append(parts, "id ASC")
	if len(parts) == 1 {
		return "ORDER BY scheduled_at ASC, id ASC"
	}
	return "ORDER BY " + strings.Join(parts, ", ")
}

func (r *PostgresRepository) MarkMissedViewings(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE viewings
		SET status = 'MISSED', updated_at = NOW()
		WHERE status = 'SCHEDULED' AND scheduled_at < NOW() - INTERVAL '1 hour'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
