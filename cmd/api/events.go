package main

import (
	"event-planner-go/internal/database"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (app *application) createEvent(c *gin.Context) {
	var event database.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := app.models.Events.Insert(&event)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

func (app *application) getAllEvents(c *gin.Context) {
	events, err := app.models.Events.GetAll()

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
	}

	c.JSON(http.StatusOK, events)
}

func (app *application) getEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Event ID"})
	}


	event, err := app.models.Events.Get(id)
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event Not Found"})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error"})
	}

	c.JSON(http.StatusOK, event)
}

func (app *application) updateEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Event ID"})
	}

	existingEvent, err := app.models.Events.Get(id)
	if existingEvent == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event Not Found"})
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Internal Server Error"})
	}

	updatedEvent := &database.Event{}

	if err := c.ShouldBindJSON(updatedEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
	}

	updatedEvent.Id = id

	if err := app.models.Events.Update(updatedEvent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, updatedEvent)

}

func (app *application) deleteEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
	}

	if err := app.models.Events.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
	}

	c.JSON(http.StatusNoContent, nil)
}

func (app *application) addAttendeeToEvent(c *gin.Context) {
	// Extract event ID
	eventId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		fmt.Printf("[DEBUG] Invalid event id param: %v\n", c.Param("id"))
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event id"})
		return
	}
	fmt.Printf("[DEBUG] Parsed eventId: %d\n", eventId)

	// Extract user ID
	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		fmt.Printf("[DEBUG] Invalid userId param: %v\n", c.Param("userId"))
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}
	fmt.Printf("[DEBUG] Parsed userId: %d\n", userId)

	// Get event
	event, err := app.models.Events.Get(eventId)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to get event with id=%d: %v\n", eventId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve event"})
		return
	}
	if event == nil {
		fmt.Printf("[DEBUG] Event with id=%d not found\n", eventId)
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	fmt.Printf("[DEBUG] Retrieved event: %+v\n", event)

	// Get user
	userToAdd, err := app.models.Users.Get(userId)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to get user with id=%d: %v\n", userId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to retrieve user"})
		return
	}
	if userToAdd == nil {
		fmt.Printf("[DEBUG] User with id=%d not found\n", userId)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	fmt.Printf("[DEBUG] Retrieved user: %+v\n", userToAdd)

	// Check if attendee already exists
	existingAttendee, err := app.models.Attendees.GetByEventAndAttendee(event.Id, userToAdd.Id)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to get attendee for eventId=%d userId=%d: %v\n", event.Id, userToAdd.Id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendee"})
		return
	}
	if existingAttendee != nil {
		fmt.Printf("[DEBUG] Attendee already exists for eventId=%d userId=%d\n", event.Id, userToAdd.Id)
		c.JSON(http.StatusConflict, gin.H{"error": "Attendee already exists"})
		return
	}

	// Insert attendee
	attendee := database.Attendees{
		EventId: event.Id,
		UserId:  userToAdd.Id,
	}
	fmt.Printf("[DEBUG] Inserting attendee: %+v\n", attendee)

	_, err = app.models.Attendees.Insert(&attendee)
	if err != nil {
		fmt.Printf("[DEBUG] Failed to insert attendee for eventId=%d userId=%d: %v\n", event.Id, userToAdd.Id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add attendee"})
		return
	}

	fmt.Printf("[DEBUG] Successfully added attendee: %+v\n", attendee)
	c.JSON(http.StatusCreated, attendee)
}


func (app *application) getAttendeesForEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid event id"})
		return
	}

	users, err := app.models.Attendees.GetAttendeesByEvent(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendees for event"})
		return
	}

	c.JSON(http.StatusOK, users)
}


func (app *application) deleteAttendeeFromEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event id"})
		return
	}

	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user id"})
		return
	}

	err = app.models.Attendees.Delete(userId, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete attendee"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (app *application) getEventsByAttendee(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid attendee id"})
		return
	}

	events, err := app.models.Attendees.GetAttendeesByEvent(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get events"})
		return
	}

	c.JSON(http.StatusOK, events)
}