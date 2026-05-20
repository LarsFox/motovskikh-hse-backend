package entities

import "time"

type Test struct {
	ID             int64
	UserID         int64
	Title          string
	QuestionsCount int
	Rating         float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
