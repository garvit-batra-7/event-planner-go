package event

type Event struct {
	ID          int    `json:"id"`
	OwnerID     int    `json:"ownerId"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required,min=10"`
	Date        string `json:"date" binding:"required,datetime=2006-01-02|datetime=2006-01-02T15:04:05Z07:00"`
	Location    string `json:"location" binding:"required,min=3"`
}

type CreateEventRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Date        string `json:"date"`
    Location    string `json:"location"`
}
