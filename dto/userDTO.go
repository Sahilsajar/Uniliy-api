package dto

type CreateUserRequestDTO struct {
	Email           string `json:"email" binding:"required,email"`
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password" binding:"required,min=6"`
	Name            string `json:"name" binding:"required"`
	Course          string `json:"course" binding:"required"`
	YOP             int32  `json:"yop" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

type CreateUserDTO struct {
	Email        string `json:"email" binding:"required,email"`
	Username     string `json:"username" binding:"required"`
	PasswordHash string `json:"password" binding:"required,min=6"`
	Name         string `json:"name" binding:"required"`
	Course       string `json:"course" binding:"required"`
	YOP          int32  `json:"yop" binding:"required"`
}

type UserProfileDTO struct {
	ID                int64  `json:"id"`
	Email             string `json:"email"`
	Username          string `json:"username"`
	Name              string `json:"name,omitempty"`
	Course            string `json:"course,omitempty"`
	YOP               int32  `json:"yop,omitempty"`
	Dob               string `json:"dob,omitempty"`
	ProfilePic        string `json:"profile_pic,omitempty"`
	CoverImg          string `json:"cover_image,omitempty"`
	CollegeId         int64  `json:"college_id,omitempty"`
	CollegeIdCard     string `json:"college_id_card,omitempty"`
	VerficationStatus string `json:"verification_status,omitempty"`
}
