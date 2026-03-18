package services

import (
	"context"

	"github.com/unilly-api/models"
	"github.com/unilly-api/repositories"
)

type AuthService struct {
	authRepo *repositories.AuthRepo
}

func NewAuthService(authRepo *repositories.AuthRepo) *AuthService {
	return &AuthService{
		authRepo: authRepo,
	}
}

func (as *AuthService) SignUp(ctx context.Context, user *models.User) (int, error) {
	return as.authRepo.SignUp(ctx, user)
}
