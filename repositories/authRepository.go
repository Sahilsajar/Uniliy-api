package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/unilly-api/db/sqlc"
	"github.com/unilly-api/dto"
)

type AuthRepo struct {
	q *db.Queries
}

func NewAuthRepo(q *db.Queries) *AuthRepo {
	return &AuthRepo{
		q: q,
	}
}

func (ar *AuthRepo) SignUp(ctx context.Context, user dto.CreateUserRequestDTO, bcryptHash string) error {
	err := ar.q.CreateUser(ctx, db.CreateUserParams{
		Email:        user.Email,
		PasswordHash: bcryptHash,
		Username:     user.Username,
		Name:         pgtype.Text{String: user.Name, Valid: true},
		Course:       pgtype.Text{String: user.Course, Valid: true},
		Yop:          pgtype.Int4{Int32: user.YOP, Valid: true},
	})

	if err != nil {
		return err
	}

	return nil
}

func (ar *AuthRepo) CanRequestOTP(
	ctx context.Context,
	email string,
) (bool, error) {

	record, err := ar.q.GetLatestOTP(ctx, email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return true, nil
		}
		return false, err
	}

	// Cooldown check
	if time.Since(record.CreatedAt.Time) < 30*time.Second {
		return false, nil
	}

	// Block brute force on active OTP
	if record.Verified.Valid &&
		!record.Verified.Bool &&
		record.Attempts.Int32 >= record.MaxAttempts.Int32 &&
		time.Now().Before(record.ExpiresAt.Time) {
		return false, nil
	}

	return true, nil
}

func (ar *AuthRepo) SaveOTP(ctx context.Context, email, otpHash string, expiresAt pgtype.Timestamp) error {
	resp, err := ar.q.GetOTPRequestByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check existing otp: %w", err)
	}

	if resp.ID != 0 {
		// Update existing record
		return ar.q.UpdateOTPRequest(ctx, db.UpdateOTPRequestParams{
			Attempts:  pgtype.Int4{Int32: 0, Valid: true},
			Verified:  pgtype.Bool{Bool: false, Valid: true},
			ExpiresAt: expiresAt,
			OtpHash:   otpHash,
			Email:     email,
		})
	}

	// Create new record
	return ar.q.CreateOTPRequest(ctx, db.CreateOTPRequestParams{
		Email:       email,
		OtpHash:     otpHash,
		ExpiresAt:   expiresAt,
		MaxAttempts: pgtype.Int4{Int32: 5, Valid: true},
	})
}

func (ar *AuthRepo) GetLatestOTP(ctx context.Context, email string) (db.OtpRequest, error) {
	return ar.q.GetLatestOTP(ctx, email)
}

func (ar *AuthRepo) MarkOTPAsVerified(ctx context.Context, email string) error {
	return ar.q.MarkOTPAsVerified(ctx, email)
}

func (ar *AuthRepo) IncrementOTPAttempts(ctx context.Context, email string) error {
	return ar.q.IncrementOTPAttempts(ctx, email)
}

func (ar *AuthRepo) DeleteOTP(ctx context.Context, email string) error {
	return ar.q.DeleteOTPRequest(ctx, email)
}

func (ar *AuthRepo) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return ar.q.GetUserByEmail(ctx, email)
}

func (r *AuthRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	return r.q.DeleteRefreshToken(ctx, token)
}

func (r *AuthRepo) GetRefreshToken(ctx context.Context, token string) (*db.RefreshToken, error) {
	rt, err := r.q.GetRefreshToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	return &rt, nil
}

func (r *AuthRepo) CreateRefreshToken(ctx context.Context, userID int64, token string, exp time.Time) error {
	_, err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    pgtype.Int8{Int64: int64(userID), Valid: true},
		TokenHash: token,
		ExpiresAt: pgtype.Timestamp{Time: exp, Valid: true},
	})
	return err
}

func (ar *AuthRepo) GetUserByID(ctx context.Context, id int64) (db.User, error) {
	return ar.q.GetUserByID(ctx, id)
}

func (ar *AuthRepo) GetUserByUserName(ctx context.Context, username string) (db.User, error) {
	return ar.q.GetUserByUsername(ctx, username)
}
