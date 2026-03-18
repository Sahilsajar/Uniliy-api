package models

import "time"

type Event struct {
	Base
	Title       string
	Description string
	Location    string
	StartTime   time.Time
	EndTime     time.Time
	CreatedBy   int64
}