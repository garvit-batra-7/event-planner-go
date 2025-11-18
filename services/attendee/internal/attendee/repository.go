package attendee

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Repository defines only DB persistence operations
type Repository interface {
	Create(ctx context.Context, attendee *Attendee) error
	GetByEventAndUser(ctx context.Context, eventID, userID int) (*Attendee, error)
	Delete(ctx context.Context, userID, eventID int) error

	// These now return IDs only
	GetUserIDsByEvent(ctx context.Context, eventID int) ([]int, error)
	GetEventIDsByUser(ctx context.Context, userID int) ([]int, error)
	DeleteByEvent(ctx context.Context, eventID int) error
}

type repository struct {
	db           *sql.DB
	queryTimeout time.Duration
}

// NewRepository now only takes a DB
func NewRepository(db *sql.DB) Repository {
	return &repository{
		db:           db,
		queryTimeout: 3 * time.Second,
	}
}

func (r *repository) withTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, r.queryTimeout)
}

// ---------------- DB-backed methods ----------------

func (r *repository) Create(ctx context.Context, a *Attendee) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	const query = `INSERT INTO attendees (event_id, user_id) VALUES ($1, $2) RETURNING id`

	if err := r.db.QueryRowContext(ctx, query, a.EventID, a.UserID).Scan(&a.ID); err != nil {
		return fmt.Errorf("insert attendee: %w", err)
	}
	return nil
}

func (r *repository) GetByEventAndUser(ctx context.Context, eventID, userID int) (*Attendee, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	const query = `SELECT id, event_id, user_id FROM attendees WHERE event_id = $1 AND user_id = $2`

	var a Attendee
	err := r.db.QueryRowContext(ctx, query, eventID, userID).
		Scan(&a.ID, &a.EventID, &a.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get attendee by event/user: %w", err)
	}
	return &a, nil
}

func (r *repository) Delete(ctx context.Context, userID, eventID int) error {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	const query = `DELETE FROM attendees WHERE user_id = $1 AND event_id = $2`

	if _, err := r.db.ExecContext(ctx, query, userID, eventID); err != nil {
		return fmt.Errorf("delete attendee: %w", err)
	}
	return nil
}

// ---------------- Simplified read model helpers ----------------

// 1) Just return user IDs — the service will hydrate users
func (r *repository) GetUserIDsByEvent(ctx context.Context, eventID int) ([]int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	const query = `SELECT user_id FROM attendees WHERE event_id = $1`

	rows, err := r.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("get user_ids for event: %w", err)
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var uid int
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan user_id: %w", err)
		}
		userIDs = append(userIDs, uid)
	}
	return userIDs, rows.Err()
}

// 2) Just return event IDs — the service will hydrate events
func (r *repository) GetEventIDsByUser(ctx context.Context, userID int) ([]int, error) {
	ctx, cancel := r.withTimeout(ctx)
	defer cancel()

	const query = `SELECT event_id FROM attendees WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get event_ids for user: %w", err)
	}
	defer rows.Close()

	var eventIDs []int
	for rows.Next() {
		var eid int
		if err := rows.Scan(&eid); err != nil {
			return nil, fmt.Errorf("scan event_id: %w", err)
		}
		eventIDs = append(eventIDs, eid)
	}
	return eventIDs, rows.Err()
}

func (r *repository) DeleteByEvent(ctx context.Context, eventID int) error {
    ctx, cancel := r.withTimeout(ctx)
    defer cancel()

    const query = `DELETE FROM attendees WHERE event_id = $1`

    if _, err := r.db.ExecContext(ctx, query, eventID); err != nil {
        return fmt.Errorf("delete attendees by event: %w", err)
    }
    return nil
}
