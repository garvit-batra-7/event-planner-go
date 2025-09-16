package database

import (
	"context"
	"database/sql"
	"time"
)

type AttendeesModel struct {
	DB *sql.DB
}

type Attendees struct {
	Id int `json:"id"`
	UserId int `json:"userId"`
	EventId int `json:"eventId"`
}

func (m *AttendeesModel) Insert(attendee *Attendees) (*Attendees, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "INSERT INTO attendees (event_id, user_id) VALUES ($1,$2) RETURNING id"
	err := m.DB.QueryRowContext(ctx, query, attendee.EventId, attendee.UserId).Scan(&attendee.Id)

	if err != nil {
		return nil, err
	}

	return attendee, nil
}

func (m *AttendeesModel) GetByEventAndAttendee(eventId, userId int) (*Attendees, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT * FROM attendees WHERE event_id = $1 AND user_id = $2"

	var attendee Attendees
	err := m.DB.QueryRowContext(ctx, query, eventId, userId).Scan(&attendee.Id, &attendee.EventId, &attendee.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &attendee, nil
}

func (m *AttendeesModel) GetAttendeesByEvent(eventId int) ([]*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT u.id, u.name, u.email
	FROM users u
	JOIN attendees a ON u.id = a.user_id
	WHERE a.event_id = $1
	`

	rows, err := m.DB.QueryContext(ctx, query, eventId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*User

	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Name, &user.Email)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}

func (m *AttendeesModel) Delete(userId, eventId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "DELETE FROM attendees WHERE user_id = $1 AND event_id = $2"
	_, err := m.DB.ExecContext(ctx, query, userId, eventId)
	if err != nil {
		return err
	}
	return nil
}

func (m *AttendeesModel) GetEventsByAttendee(attendeeId int) ([]*Event,error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	SELECT e.id, e.owner_id, e.name, e.description, e.date, e.location
	FROM events e
	JOIN attendees a ON e.id = a.event_id
	WHERE a.user_id = $1
	`

	rows, err := m.DB.QueryContext(ctx, query, attendeeId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var events []*Event
	for rows.Next() {
		var event Event
		err := rows.Scan(&event.Id, &event.OwnerId, &event.Name, &event.Description, &event.Date, &event.Location)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}
	return events, nil
}