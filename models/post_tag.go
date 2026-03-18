package models

type PostTag struct {
	Base
	PostID       int64
	TaggedUserID int64
}