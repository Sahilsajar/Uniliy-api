package models

import "time"

type User struct {
	Base
	Username     string
	Email        string
	Name         string
	DOB          *time.Time
	ProfilePic   string
	CoverImage   string
	PasswordHash string
	CollegeID    *int64
}