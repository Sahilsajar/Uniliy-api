package models

type Comment struct {
	Base
	Message          string
	PostID           int64
	UserID           int64
	ParentCommentID  *int64 // nullable (for replies)
}