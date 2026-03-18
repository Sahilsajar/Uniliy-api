package repositories

import (
	"context"
	db "github.com/unilly-api/db/sqlc"
	"github.com/unilly-api/models"
)

type AuthRepo struct {
	q *db.Queries
}

func NewAuthRepo(q *db.Queries) *AuthRepo {
	return &AuthRepo{
		q: q,
	}
}


func (ar *AuthRepo) SignUp(ctx context.Context, user *models.User) (int32, error) {
	createdUser, err := ar.q.CreateUser(ctx, db.CreateUserParams{
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Username:     user.Username,
	})

	if err != nil {
		return 0, err
	}

	return createdUser.ID, nil
}
