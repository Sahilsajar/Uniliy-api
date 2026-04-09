package repositories

import (
	"context"
	"fmt"

	db "github.com/unilly-api/db/sqlc"
	"github.com/unilly-api/dto"
)

type UserRepository struct {
	q *db.Queries
}

func NewUserRepository(q *db.Queries) *UserRepository {
	return &UserRepository{
		q: q,
	}
}

func (ur *UserRepository) GetFollowers(ctx context.Context, userID int64) (dto.GetFollowersResponseDTO, error) {
	followers, err := ur.q.GetUserFollowers(ctx, userID)
	if err != nil {
		return dto.GetFollowersResponseDTO{}, fmt.Errorf("failed to get followers: %w", err)
	}
	return dto.GetFollowersResponseDTO{Followers: followers}, nil
}

func (ur *UserRepository) GetFollowing(ctx context.Context, userID int64) (dto.GetFollowingResponseDTO, error) {
	following, err := ur.q.GetUserFollowing(ctx, userID)
	if err != nil {
		return dto.GetFollowingResponseDTO{}, fmt.Errorf("failed to get following: %w", err)
	}
	return dto.GetFollowingResponseDTO{Following: following}, nil
}
