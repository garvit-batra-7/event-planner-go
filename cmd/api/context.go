package main

import (
	"event-planner-go/internal/database"

	"github.com/gin-gonic/gin"
)

func (app *application) getUserFromContext(c *gin.Context) *database.User {
	contextUser, exist := c.Get("user")
	if !exist {
		return &database.User{}
	}
	user, ok := contextUser.(*database.User)
	if !ok {
		return &database.User{}
	}

	return user
}