package services

import (
	"context"

	"github.com/unilly-api/dto"
	"github.com/unilly-api/repositories"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (us *UserService) GetFollowers(ctx context.Context, userID int64) (dto.GetFollowersResponseDTO, error) {
	return us.userRepo.GetFollowers(ctx, userID)
}

func (us *UserService) GetFollowing(ctx context.Context, userID int64) (dto.GetFollowingResponseDTO, error) {
	return us.userRepo.GetFollowing(ctx, userID)
}
