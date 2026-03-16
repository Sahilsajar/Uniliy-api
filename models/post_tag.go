package models

type PostTag struct {
	Base
	PostID uint `gorm:"index"`
	TaggedUserID uint `gorm:"index"`
}