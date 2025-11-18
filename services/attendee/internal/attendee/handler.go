package attendee

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"event-planner-go/pkg/shared"
)

// Handler handles attendee HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new attendee handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// AddToEvent adds an attendee to an event
// @Summary Adds an attendee to an event
// @Description Adds an attendee to an event
// @Tags attendees
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Param userId path int true "User ID"
// @Success 201 {object} Attendee
// @Router /api/v1/events/{id}/attendees/{userId} [post]
// @Security BearerAuth
func (h *Handler) AddToEvent(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get authenticated user ID from context (set by AuthMiddleware)
	requestingUserID, ok := shared.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Grab request context and pass it to service
	ctx := c.Request.Context()

	// Add attendee
	attendee, err := h.service.AddToEvent(ctx, eventID, userID, requestingUserID)
	if err != nil {
		switch err.Error() {
		case "event not found", "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		case "not authorized to add attendees to this event":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		case "user is already attending this event":
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, attendee)
}

// RemoveFromEvent removes an attendee from an event
// @Summary Deletes an attendee from an event
// @Description Deletes an attendee from an event
// @Tags attendees
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Param userId path int true "User ID"
// @Success 204
// @Router /api/v1/events/{id}/attendees/{userId} [delete]
// @Security BearerAuth
func (h *Handler) RemoveFromEvent(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get authenticated user ID from context
	requestingUserID, ok := shared.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// Grab request context
	ctx := c.Request.Context()

	// Remove attendee
	err = h.service.RemoveFromEvent(ctx, eventID, userID, requestingUserID)
	if err != nil {
		switch err.Error() {
		case "event not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		case "not authorized to remove attendees from this event":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		case "user is not attending this event":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetUsersByEvent returns all attendees for a given event
// @Summary Returns all attendees for a given event
// @Description Returns all attendees for a given event
// @Tags attendees
// @Accept json
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} []User
// @Router /api/v1/events/{id}/attendees [get]
func (h *Handler) GetUsersByEvent(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	ctx := c.Request.Context()

	users, err := h.service.GetUsersByEvent(ctx, eventID)
	if err != nil {
		if err.Error() == "event not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetEventsByUser returns all events a user is attending
// @Summary Returns all events for a given attendee
// @Description Returns all events for a given attendee
// @Tags attendees
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} []Event
// @Router /api/v1/attendees/{id}/events [get]
func (h *Handler) GetEventsByUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx := c.Request.Context()

	events, err := h.service.GetEventsByUser(ctx, userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// DeleteAttendeesForEvent removes all attendees from an event
func (h *Handler) DeleteAttendeesForEvent(c *gin.Context) {
    eventID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
        return
    }

    ctx := c.Request.Context()

    if err := h.service.DeleteAttendeesForEvent(ctx, eventID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusNoContent, nil)
}
