package models

import "time"

type Event struct {
	Base
	Title       string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	Location    string    `gorm:"size:255"`
	StartTime   time.Time
	EndTime     time.Time

	CreatedBy uint
}