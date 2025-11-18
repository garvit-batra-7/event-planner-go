package event

import (
	"context"
	"errors"
	"fmt"
	"event-planner-go/pkg/shared"
	"event-planner-go/services/event/internal/attendeeclient"
)

// Service handles event business logic
type Service struct {
	eventRepo    Repository
	attendeeClient attendeeclient.Client
	logger       *shared.Logger
}

// NewService creates a new event service
func NewService(eventRepo Repository, attendeeClient attendeeclient.Client, logger *shared.Logger) *Service {
	return &Service{
		eventRepo:      eventRepo,
		attendeeClient: attendeeClient,
		logger:         logger,
	}
}

// Create creates a new event
func (s *Service) Create(ctx context.Context, event *Event, token string) error {
	// Validate event data
	if err := s.validateEvent(event); err != nil {
		s.logger.Warn("Event validation failed", "error", err.Error())
		return err
	}

	// Create event
	err := s.eventRepo.Create(event)
	if err != nil {
		s.logger.Error("Failed to create event", "error", err.Error())
		return fmt.Errorf("failed to create event")
	}

	if err := s.attendeeClient.AddAttendee(ctx, event.ID, event.OwnerID, token); err != nil {
		// log error, decide if you want to treat as fatal or not
		return err
	}

	s.logger.LogEvent(shared.LevelInfo, fmt.Sprintf("%d", event.ID), "Event created",
		"name", event.Name,
		"owner_id", event.OwnerID,
	)

	return nil
}

// GetByID retrieves an event by ID
func (s *Service) GetByID(id int) (*Event, error) {
	event, err := s.eventRepo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to get event", "error", err.Error(), "event_id", id)
		return nil, fmt.Errorf("failed to get event")
	}

	if event == nil {
		return nil, errors.New("event not found")
	}

	return event, nil
}

// GetAll retrieves all events
func (s *Service) GetAll() ([]*Event, error) {
	events, err := s.eventRepo.GetAll()
	if err != nil {
		s.logger.Error("Failed to get events", "error", err.Error())
		return nil, fmt.Errorf("failed to get events")
	}

	return events, nil
}

// Update updates an existing event
func (s *Service) Update(event *Event, userID int) error {
	// Check if event exists
	existing, err := s.eventRepo.GetByID(event.ID)
	if err != nil {
		s.logger.Error("Failed to get event", "error", err.Error(), "event_id", event.ID)
		return fmt.Errorf("failed to get event")
	}

	if existing == nil {
		return errors.New("event not found")
	}

	// Check ownership
	if existing.OwnerID != userID {
		s.logger.LogUser(shared.LevelWarn, fmt.Sprintf("%d", userID), "unauthorized",
			"User attempted to update event they don't own",
			"event_id", event.ID,
		)
		return errors.New("not authorized to update this event")
	}

	// Validate updated data
	if err := s.validateEvent(event); err != nil {
		return err
	}

	// Preserve owner ID
	event.OwnerID = existing.OwnerID

	// Update event
	err = s.eventRepo.Update(event)
	if err != nil {
		s.logger.Error("Failed to update event", "error", err.Error(), "event_id", event.ID)
		return fmt.Errorf("failed to update event")
	}

	s.logger.LogEvent(shared.LevelInfo, fmt.Sprintf("%d", event.ID), "Event updated",
		"name", event.Name,
	)

	return nil
}

// Delete deletes an event
func (s *Service) Delete(ctx context.Context, eventID, userID int, token string) error {
	// Check if event exists
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		s.logger.Error("Failed to get event", "error", err.Error(), "event_id", eventID)
		return fmt.Errorf("failed to get event")
	}

	if event == nil {
		return errors.New("event not found")
	}

	// Check ownership
	if event.OwnerID != userID {
		s.logger.LogUser(shared.LevelWarn, fmt.Sprintf("%d", userID), "unauthorized",
			"User attempted to delete event they don't own",
			"event_id", eventID,
		)
		return errors.New("not authorized to delete this event")
	}

	if err := s.attendeeClient.DeleteAttendeesForEvent(ctx, eventID, token); err != nil {
		// log error, decide if you want to treat as fatal or not
		return err
	}

	// Delete event
	err = s.eventRepo.Delete(eventID)
	if err != nil {
		s.logger.Error("Failed to delete event", "error", err.Error(), "event_id", eventID)
		return fmt.Errorf("failed to delete event")
	}

	s.logger.LogEvent(shared.LevelInfo, fmt.Sprintf("%d", eventID), "Event deleted")

	return nil
}

// validateEvent validates event data
func (s *Service) validateEvent(event *Event) error {
	if event.Name == "" {
		return errors.New("event name is required")
	}

	if len(event.Description) < 10 {
		return errors.New("event description must be at least 10 characters")
	}

	if event.Location == "" || len(event.Location) < 3 {
		return errors.New("event location must be at least 3 characters")
	}

	if event.Date == "" {
		return errors.New("event date is required")
	}

	return nil
}