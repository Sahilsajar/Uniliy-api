package models

type UserFollow struct {
	Base
	FollowerUserID  uint `gorm:"index"`
	FollowingUserID uint `gorm:"index"`
}