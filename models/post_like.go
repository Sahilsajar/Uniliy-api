package models

type PostLike struct {
	Base
	PostID uint `gorm:"index"`
	UserID uint `gorm:"index"`
}