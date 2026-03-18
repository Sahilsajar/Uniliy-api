package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/unilly-api/models"
)

type AuthRepo struct {
	db *pgxpool.Pool
}

func NewAuthRepo(db *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{
		db: db,
	}
}

func (ar *AuthRepo) SignUp(ctx context.Context, user *models.User) (int, error) {
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id)`

	var id int
	err := ar.db.QueryRow(ctx, query, user.Username, user.Email, user.Password).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
