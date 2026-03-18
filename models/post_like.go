package models

type PostLike struct {
	Base
	PostID int64
	UserID int64
}