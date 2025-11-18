package event

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"event-planner-go/pkg/shared"
)

// Handler handles event HTTP requests
type Handler struct {
	service *Service
}

// NewHandler creates a new event handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Create creates a new event
// @Summary      Create event
// @Description  Creates a new event for the authenticated user. The ownerId is taken from the JWT, not the request body.
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        body  body      CreateEventRequest  true  "Event payload"
// @Success      201   {object}  Event
// @Failure      400   {object}  map[string]string  "Invalid request payload"
// @Failure      401   {object}  map[string]string  "User not found in context / unauthorized"
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Router       /api/v1/events [post]
// @Security     BearerAuth
func (h *Handler) Create(c *gin.Context) {
	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	userID, ok := shared.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	event := &Event{
		OwnerID:     userID,
		Name:        req.Name,
		Description: req.Description,
		Date:        req.Date,
		Location:    req.Location,
	}

	token, _ := shared.GetRawTokenFromContext(c) // if needed
	if err := h.service.Create(c.Request.Context(), event, token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// GetAll returns all events
// @Summary      List events
// @Description  Returns all events
// @Tags         events
// @Accept       json
// @Produce      json
// @Success      200  {array}   Event
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/events [get]
// @Security     BearerAuth
func (h *Handler) GetAll(c *gin.Context) {
	events, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetByID returns a single event
// @Summary      Get event by ID
// @Description  Returns a single event by its ID
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Event ID"
// @Success      200  {object}  Event
// @Failure      400  {object}  map[string]string  "Invalid event ID"
// @Failure      404  {object}  map[string]string  "Event not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/events/{id} [get]
// @Security     BearerAuth
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.service.GetByID(id)
	if err != nil {
		if err.Error() == "event not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// Update updates an existing event
// @Summary      Update event
// @Description  Updates an existing event. Only the owner of the event can update it.
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id    path      int    true  "Event ID"
// @Param        body  body      Event  true  "Updated event payload (ownerId is ignored)"
// @Success      200   {object}  Event
// @Failure      400   {object}  map[string]string  "Invalid event ID or payload"
// @Failure      401   {object}  map[string]string  "User not found in context"
// @Failure      403   {object}  map[string]string  "Not authorized to update this event"
// @Failure      404   {object}  map[string]string  "Event not found"
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Router       /api/v1/events/{id} [put]
// @Security     BearerAuth
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.ID = id

	// Get authenticated user from context
	userID, ok := shared.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
		return
	}

	// Update event
	err = h.service.Update(&event, userID)
	if err != nil {
		if err.Error() == "event not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "not authorized to update this event" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// Delete deletes an existing event
// @Summary      Delete event
// @Description  Deletes an existing event. Only the owner of the event can delete it. Also triggers attendee cleanup.
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        id   path  int  true  "Event ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string  "Invalid event ID"
// @Failure      401  {object}  map[string]string  "User not found in context"
// @Failure      403  {object}  map[string]string  "Not authorized to delete this event"
// @Failure      404  {object}  map[string]string  "Event not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/events/{id} [delete]
// @Security     BearerAuth
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID, ok := shared.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	err = h.service.Delete(c, id, userID, c.GetHeader("Authorization"))
	if err != nil {
		if err.Error() == "event not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "not authorized to delete this event" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
