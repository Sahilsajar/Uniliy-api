package repositories

import (
	"context"
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

func (ar *AuthRepo) SignUp(ctx context.Context, user dto.CreateUserRequestDTO, bcryptHash string) (int32, error) {
	createdUser, err := ar.q.CreateUser(ctx, dto.CreateUserRequestDTO{
		Email:        user.Email,
		PasswordHash: bcryptHash,
		Username:     user.Username,
		Name:         user.Name,
		Course:       user.Course,
		YOP:          user.YOP,
	})

	if err != nil {
		return 0, err
	}

	return createdUser.ID, nil
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
	if !record.Verified.Valid &&
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
