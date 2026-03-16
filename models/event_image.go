package models

type EventImage struct {
	Base
	EventID  uint
	ImageURL string `gorm:"size:255;not null"`
}