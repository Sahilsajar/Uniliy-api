package models

type Post struct {
	Base
	Title  string
	Body   string
	Status string
	UserID int64
}