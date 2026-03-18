package models

type PostImage struct {
	Base
	ImageURL string
	PostID   int64
}