package models

type Comment struct {
	Base
	Message string `gorm:"type:text;not null"`

	PostID uint
	UserID uint

	ParentCommentID *uint
	Replies         []Comment `gorm:"foreignKey:ParentCommentID"`

	Likes []CommentLike
}