package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unilly-api/api"
	"github.com/unilly-api/services"
)

type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (uc *UserController) GetFollowers(ctx *gin.Context) error {
	userIDTemp, _ := ctx.Get("user_id")
	userID := userIDTemp.(int64)
	response, err := uc.userService.GetFollowers(ctx.Request.Context(), userID)
	if err != nil {
		return err
	}
	api.Success(ctx, http.StatusOK, "Followers retrieved successfully", response)
	return nil
}

func (uc *UserController) GetFollowing(ctx *gin.Context) error {
	userIDTemp, _ := ctx.Get("user_id")
	userID := userIDTemp.(int64)
	response, err := uc.userService.GetFollowing(ctx.Request.Context(), userID)
	if err != nil {
		return err
	}
	api.Success(ctx, http.StatusOK, "Following retrieved successfully", response)
	return nil
}
