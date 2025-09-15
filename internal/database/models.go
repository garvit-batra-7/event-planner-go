package database

import "database/sql"

type Models struct {
	Users UserModel
	Events EventsModel
	Attendees AttendeesModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB : db},
		Events: EventsModel{DB : db},
		Attendees: AttendeesModel{DB : db},
	}
}