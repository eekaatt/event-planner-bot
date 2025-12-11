package models

import ("time")

type Event struct {
	ID int64 `json:"id"` // id мероприятия
	Title string `json:"title"`  // название мероприятия
	Description string `json:"description"`  // описание
	Location string `json:"location"`  // место проведения
	EventDate time.Time `json:"event_date"`  // дата проведения
	CreatedBy int64 `json:"created_by"`  // кем создано мероприятие
	CreatedAt time.Time `json:"created_at"`  // когда создано
    UpdatedAt time.Time `json:"updated_at"`  // когда обновлено
}

type EventStatus string

const (
	StatusPlanned EventStatus = "planned"
	StatusOngoing EventStatus = "ongoing"
	StatusEnded EventStatus = "ended"
	StatusCancelled EventStatus = "cancelled"
)
