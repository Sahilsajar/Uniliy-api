package models

type CommentLike struct {
	Base
	CommentID uint `gorm:"index"`
	UserID    uint `gorm:"index"`
}