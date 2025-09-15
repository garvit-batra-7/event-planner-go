package database

import "database/sql"

type AttendeesModel struct {
	DB *sql.DB
}

type Attendees struct {
	Id int `json:"id"`
	UserId int `json:"userId"`
	EventId int `json:"eventId"`
}