package models

type PostImage struct {
	Base
	ImageURL string `gorm:"size:255;not null"`

	PostID uint
}