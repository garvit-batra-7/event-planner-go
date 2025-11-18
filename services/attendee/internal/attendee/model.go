package attendee

type Attendee struct {
	ID int `json:"id"`
	UserID int `json:"userId"`
	EventID int `json:"eventId"`
}