package models

import "time"

type User struct {
	Base
	Username  string     `gorm:"size:100;uniqueIndex;not null"`
	Email     string     `gorm:"size:150;uniqueIndex"`
	Name      string     `gorm:"size:150"`
	DOB       *time.Time
	ProfilePic string    `gorm:"size:255"`
	CoverImage string    `gorm:"size:255"`

	CollegeID uint
	College   College

	Posts     []Post
	Comments  []Comment
}