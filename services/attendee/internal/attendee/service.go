package attendee

import (
	"context"
	"errors"
	"fmt"

	"event-planner-go/pkg/shared"
	"event-planner-go/services/attendee/internal/eventclient"
	"event-planner-go/services/attendee/internal/userclient"
)

// User is what this service returns to API callers.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Event is what this service returns to API callers.
type Event struct {
	ID          int    `json:"id"`
	OwnerID     int    `json:"ownerId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Location    string `json:"location"`
}

// Service handles attendee business logic.
type Service struct {
	repo   Repository
	events eventclient.Client
	users  userclient.Client
	logger *shared.Logger
}

// NewService creates a new attendee service.
func NewService(
	repo Repository,
	events eventclient.Client,
	users userclient.Client,
	logger *shared.Logger,
) *Service {
	return &Service{
		repo:   repo,
		events: events,
		users:  users,
		logger: logger,
	}
}

// AddToEvent adds a user as attendee to an event.
func (s *Service) AddToEvent(
	ctx context.Context,
	eventID, userID, requestingUserID int,
) (*Attendee, error) {
	// 1) Check event exists and requester is owner.
	evt, err := s.events.GetEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, eventclient.ErrNotFound) {
			return nil, errors.New("event not found")
		}
		s.logger.Error("Failed to fetch event", "error", err.Error(), "event_id", eventID)
		return nil, fmt.Errorf("failed to fetch event")
	}
	if evt.OwnerID != requestingUserID {
		return nil, errors.New("not authorized to add attendees to this event")
	}

	// 2) Check user exists.
	if _, err := s.users.GetUser(ctx, userID); err != nil {
		if errors.Is(err, userclient.ErrNotFound) {
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to fetch user", "error", err.Error(), "user_id", userID)
		return nil, fmt.Errorf("failed to fetch user")
	}

	// 3) Check if attendee already exists.
	existing, err := s.repo.GetByEventAndUser(ctx, eventID, userID)
	if err != nil {
		s.logger.Error("Failed to check attendee", "error", err.Error())
		return nil, fmt.Errorf("failed to check attendee")
	}
	if existing != nil {
		return nil, errors.New("user is already attending this event")
	}

	attendee := &Attendee{
		EventID: eventID,
		UserID:  userID,
	}

	if err := s.repo.Create(ctx, attendee); err != nil {
		s.logger.Error("Failed to add attendee", "error", err.Error())
		return nil, fmt.Errorf("failed to add attendee")
	}

	s.logger.LogEvent(
		shared.LevelInfo,
		fmt.Sprintf("%d", eventID),
		"Attendee added",
		"user_id", userID,
		"requesting_user_id", requestingUserID,
		"attendee_id", attendee.ID,
	)

	return attendee, nil
}

// RemoveFromEvent removes a user from an event.
func (s *Service) RemoveFromEvent(
	ctx context.Context,
	eventID, userID, requestingUserID int,
) error {
	// 1) Check event exists & ownership.
	evt, err := s.events.GetEvent(ctx, eventID)
	if err != nil {
		if errors.Is(err, eventclient.ErrNotFound) {
			return errors.New("event not found")
		}
		s.logger.Error("Failed to fetch event", "error", err.Error(), "event_id", eventID)
		return fmt.Errorf("failed to fetch event")
	}
	if evt.OwnerID != requestingUserID {
		return errors.New("not authorized to remove attendees from this event")
	}

	// 2) Check attendee exists.
	existing, err := s.repo.GetByEventAndUser(ctx, eventID, userID)
	if err != nil {
		s.logger.Error("Failed to check attendee", "error", err.Error())
		return fmt.Errorf("failed to check attendee")
	}
	if existing == nil {
		return errors.New("user is not attending this event")
	}

	if err := s.repo.Delete(ctx, userID, eventID); err != nil {
		s.logger.Error("Failed to remove attendee", "error", err.Error())
		return fmt.Errorf("failed to remove attendee")
	}

	s.logger.LogEvent(
		shared.LevelInfo,
		fmt.Sprintf("%d", eventID),
		"Attendee removed",
		"user_id", userID,
		"requesting_user_id", requestingUserID,
	)

	return nil
}

// GetUsersByEvent retrieves all users attending an event.
func (s *Service) GetUsersByEvent(
	ctx context.Context,
	eventID int,
) ([]*User, error) {
	// Optional but useful: verify event exists.
	if _, err := s.events.GetEvent(ctx, eventID); err != nil {
		if errors.Is(err, eventclient.ErrNotFound) {
			return nil, errors.New("event not found")
		}
		s.logger.Error("Failed to fetch event", "error", err.Error(), "event_id", eventID)
		return nil, fmt.Errorf("failed to fetch event")
	}

	// 1) Get user IDs from DB.
	userIDs, err := s.repo.GetUserIDsByEvent(ctx, eventID)
	if err != nil {
		s.logger.Error("Failed to get user IDs", "error", err.Error(), "event_id", eventID)
		return nil, fmt.Errorf("failed to get attendees")
	}

	// 2) Hydrate via user service.
	users := make([]*User, 0, len(userIDs))
	for _, uid := range userIDs {
		u, err := s.users.GetUser(ctx, uid)
		if err != nil {
			if errors.Is(err, userclient.ErrNotFound) {
				// user was deleted but still referenced; skip
				continue
			}
			s.logger.Error("Failed to fetch user", "error", err.Error(), "user_id", uid)
			return nil, fmt.Errorf("failed to fetch user %d", uid)
		}
		if u == nil {
			continue
		}
		users = append(users, &User{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		})
	}

	return users, nil
}

// GetEventsByUser retrieves all events a user is attending.
func (s *Service) GetEventsByUser(
	ctx context.Context,
	userID int,
) ([]*Event, error) {
	// Optional: verify user exists.
	if _, err := s.users.GetUser(ctx, userID); err != nil {
		if errors.Is(err, userclient.ErrNotFound) {
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to fetch user", "error", err.Error(), "user_id", userID)
		return nil, fmt.Errorf("failed to fetch user")
	}

	// 1) Get event IDs from DB.
	eventIDs, err := s.repo.GetEventIDsByUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get event IDs", "error", err.Error(), "user_id", userID)
		return nil, fmt.Errorf("failed to get events")
	}

	// 2) Hydrate via event service.
	events := make([]*Event, 0, len(eventIDs))
	for _, eid := range eventIDs {
		e, err := s.events.GetEvent(ctx, eid)
		if err != nil {
			if errors.Is(err, eventclient.ErrNotFound) {
				// event was deleted but still referenced; skip
				continue
			}
			s.logger.Error("Failed to fetch event", "error", err.Error(), "event_id", eid)
			return nil, fmt.Errorf("failed to fetch event %d", eid)
		}
		if e == nil {
			continue
		}
		events = append(events, &Event{
			ID:          e.ID,
			OwnerID:     e.OwnerID,
			Name:        e.Name,
			Description: e.Description,
			Date:        e.Date,
			Location:    e.Location,
		})
	}

	return events, nil
}

// DeleteAttendeesForEvent deletes all attendees for a given event.
// Called by the Event service after the event is deleted.
func (s *Service) DeleteAttendeesForEvent(ctx context.Context, eventID int) error {
    if err := s.repo.DeleteByEvent(ctx, eventID); err != nil {
        s.logger.Error("Failed to delete attendees for event", "error", err.Error(), "event_id", eventID)
        return fmt.Errorf("failed to delete attendees for event")
    }
    return nil
}
