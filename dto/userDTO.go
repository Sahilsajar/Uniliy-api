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
