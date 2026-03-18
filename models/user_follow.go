package models

type UserFollow struct {
	Base
	FollowerUserID  int64
	FollowingUserID int64
}