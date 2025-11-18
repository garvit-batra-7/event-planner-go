package event

import (
	"context"
	"database/sql"
	"time"
)

// Repository defines the interface for event data operations
type Repository interface {
	Create(event *Event) error
	GetByID(id int) (*Event, error)
	GetAll() ([]*Event, error)
	Update(event *Event) error
	Delete(id int) error
}

type repository struct {
	db *sql.DB
}

// NewRepository creates a new event repository
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// Create inserts a new event
func (r *repository) Create(event *Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `INSERT INTO events (owner_id, name, description, date, location) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		event.OwnerID,
		event.Name,
		event.Description,
		event.Date,
		event.Location,
	).Scan(&event.ID)
}

// GetByID retrieves a single event
func (r *repository) GetByID(id int) (*Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, owner_id, name, description, date, location FROM events WHERE id = $1"

	var event Event
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.OwnerID,
		&event.Name,
		&event.Description,
		&event.Date,
		&event.Location,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &event, nil
}

// GetAll retrieves all events
func (r *repository) GetAll() ([]*Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, owner_id, name, description, date, location FROM events"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.OwnerID,
			&event.Name,
			&event.Description,
			&event.Date,
			&event.Location,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// Update modifies an existing event
func (r *repository) Update(event *Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE events 
	          SET name = $1, description = $2, date = $3, location = $4 
	          WHERE id = $5`

	_, err := r.db.ExecContext(ctx, query,
		event.Name,
		event.Description,
		event.Date,
		event.Location,
		event.ID,
	)

	return err
}

// Delete removes an event
func (r *repository) Delete(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "DELETE FROM events WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}