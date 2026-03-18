package models

type CommentLike struct {
	Base
	CommentID int64
	UserID    int64
}