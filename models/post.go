package models

type Post struct {
	Base
	Title  string `gorm:"size:255"`
	Body   string `gorm:"type:text"`
	Status string `gorm:"size:50;index"`

	UserID uint
	User   User

	Images   []PostImage
	Comments []Comment
	Likes    []PostLike
	Tags     []PostTag
}