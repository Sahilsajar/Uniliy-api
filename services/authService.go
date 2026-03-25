package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/unilly-api/api"
	db "github.com/unilly-api/db/sqlc"
	"github.com/unilly-api/dto"
	"github.com/unilly-api/repositories"
	"github.com/unilly-api/utility"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo *repositories.AuthRepo
	client   *http.Client
}

func NewAuthService(authRepo *repositories.AuthRepo) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (as *AuthService) SignUp(ctx context.Context, user dto.CreateUserRequestDTO) error {
	if err := as.requireVerifiedOTP(ctx, user.Email); err != nil {
		return err
	}

	bycryptHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := as.authRepo.SignUp(ctx, user, string(bycryptHash)); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return api.Conflict("USER_ALREADY_EXISTS", "User with this email or username already exists").WithCause(err)
		}
		return api.Internal("SIGNUP_FAILED", "Failed to sign up").WithCause(err)
	}
	if err := as.authRepo.DeleteOTP(ctx, user.Email); err != nil {
		return api.Internal("OTP_CLEANUP_FAILED", "Failed to finalize signup").WithCause(err)
	}
	return nil
}

func GenerateOTP(length int) (string, error) {
	if length <= 0 {
		length = 6
	}

	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Use string instead of Int64 (avoids overflow issues)
	otp := n.String()

	// Zero pad manually
	if len(otp) < length {
		otp = strings.Repeat("0", length-len(otp)) + otp
	}

	return otp, nil
}
func hashOTP(otp string) string {
	hash := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(hash[:])
}

func (as *AuthService) GenerateAndSendOTP(
	ctx context.Context,
	email string,
) error {

	// Rate limit check (critical to do this BEFORE generating OTP to prevent abuse)
	allowed, err := as.authRepo.CanRequestOTP(ctx, email)
	if err != nil {
		return api.Internal("OTP_RATE_LIMIT_CHECK_FAILED", "Failed to process OTP request").WithCause(err)
	}
	if !allowed {
		return api.TooManyRequests("OTP_RATE_LIMITED", "Too many OTP requests. Please wait before retrying")
	}

	// Generate OTP
	otp, err := GenerateOTP(6)
	if err != nil {
		return api.Internal("OTP_GENERATION_FAILED", "Failed to generate OTP").WithCause(err)
	}
	otpHash := hashOTP(otp)
	expiresAt := pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute), Valid: true}

	//  Save OTP FIRST
	err = as.authRepo.SaveOTP(ctx, email, otpHash, expiresAt)
	if err != nil {
		return api.Internal("OTP_SAVE_FAILED", "Failed to generate OTP").WithCause(err)
	}

	// Send email
	if err := as.sendEmail(ctx, email, otp); err != nil {
		return api.Internal("OTP_DELIVERY_FAILED", "Failed to send OTP").WithCause(err)
	}
	return nil
}

func (as *AuthService) sendEmail(ctx context.Context, email, otp string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	emailFrom := os.Getenv("BREVO_EMAIL")
	if emailFrom == "" {
		return fmt.Errorf("brevo sender email not configured")
	}
	if apiKey == "" {
		return fmt.Errorf("brevo api key not configured")
	}

	payload, err := json.Marshal(map[string]any{
		"sender": map[string]string{
			"name":  "Unilly",
			"email": emailFrom,
		},
		"to": []map[string]string{
			{"email": email},
		},
		"subject": "Your OTP Code",
		"htmlContent": fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <style>
    body {
      margin: 0;
      padding: 0;
      font-family: Arial, sans-serif;
      background-color: #f4f4f4;
    }

    .container {
      max-width: 500px;
      margin: auto;
      background: #ffffff;
      padding: 20px;
      border-radius: 12px;
      text-align: center;
    }

    .logo {
      width: 120px;
      margin-bottom: 20px;
    }

    .otp {
      font-size: 28px;
      font-weight: bold;
      letter-spacing: 4px;
      color: #2563eb;
      margin: 20px 0;
    }

    .footer {
      font-size: 12px;
      color: #777;
      margin-top: 20px;
    }

    /* Dark mode */
    @media (prefers-color-scheme: dark) {
      body {
        background-color: #000000;
      }

      .container {
        background: #111111;
        color: #ffffff;
      }

      .light-logo {
        display: none;
      }

      .dark-logo {
        display: block;
      }
    }

    /* Default */
    .dark-logo {
      display: none;
    }
  </style>
</head>

<body>
  <div class="container">

    <!-- Logos -->
    <img src="https://res.cloudinary.com/ddcf3mjcn/image/upload/v1774289181/Light_Version_aqycol.png"
         class="logo light-logo" />

    <img src="https://res.cloudinary.com/ddcf3mjcn/image/upload/v1774289057/UnillyLogo_nqv9ik.png"
         class="logo dark-logo" />

    <h2>Your OTP Code</h2>

    <p>Use the code below to continue:</p>

    <div class="otp">%s</div>

    <p>This code expires in 10 minutes.</p>

    <p>If you didn’t request this, you can safely ignore this email.</p>

    <div class="footer">
      © Unilly • Secure Authentication System
    </div>

  </div>
</body>
</html>
`, otp),
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.brevo.com/v3/smtp/email",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := as.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("email provider error: %s", resp.Status)
	}

	return nil
}

func (as *AuthService) VerifyOTP(ctx context.Context, email string, otp string) error {
	record, err := as.authRepo.GetLatestOTP(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.Unauthorized("OTP_NOT_FOUND", "No OTP request found for this email")
		}
		return api.Internal("OTP_LOOKUP_FAILED", "Failed to verify OTP").WithCause(err)
	}

	if time.Now().After(record.ExpiresAt.Time) {
		return api.Unauthorized("OTP_EXPIRED", "OTP expired")
	}

	if record.Attempts.Int32 >= record.MaxAttempts.Int32 {
		return api.TooManyRequests("OTP_ATTEMPTS_EXCEEDED", "Too many invalid OTP attempts")
	}

	if hashOTP(otp) != record.OtpHash {
		_ = as.authRepo.IncrementOTPAttempts(ctx, record.Email)
		return api.Unauthorized("OTP_INVALID", "Invalid OTP")
	}

	err = as.authRepo.MarkOTPAsVerified(ctx, record.Email)
	if err != nil {
		return api.Internal("OTP_VERIFY_FAILED", "Failed to verify OTP").WithCause(err)
	}

	err = as.authRepo.DeleteOTP(ctx, record.Email)
	if err != nil {
		return api.Internal("OTP_CLEANUP_FAILED", "Failed to finalize OTP verification").WithCause(err)
	}

	return nil
}

func (as *AuthService) requireVerifiedOTP(ctx context.Context, email string) error {
	record, err := as.authRepo.GetLatestOTP(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.Forbidden("OTP_VERIFICATION_REQUIRED", "Verify OTP before continuing")
		}
		return api.Internal("OTP_LOOKUP_FAILED", "Failed to validate OTP status").WithCause(err)
	}

	if time.Now().After(record.ExpiresAt.Time) {
		return api.Forbidden("OTP_EXPIRED", "Verified OTP expired. Request a new OTP")
	}

	if !record.Verified.Valid || !record.Verified.Bool {
		return api.Forbidden("OTP_VERIFICATION_REQUIRED", "Verify OTP before continuing")
	}

	return nil
}

// Login
func (as *AuthService) Login(
	ctx context.Context,
	identifier string,
	password string,
) (string, string, error) {
	var user *db.User
	var err error

	if utility.IsEmail(identifier) {
		u, err := as.authRepo.GetUserByEmail(ctx, identifier)
		if err != nil {
			user = nil
		} else {
			user = &u
		}
	} else {
		u, err := as.authRepo.GetUserByUserName(ctx, identifier)
		if err != nil {
			user = nil
		} else {
			user = &u
		}
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", api.Unauthorized("INVALID_CREDENTIALS", "Invalid email or password")
		}
		return "", "", api.Internal("LOGIN_FAILED", "Failed to login").WithCause(err)
	}
	err = utility.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return "", "", api.Unauthorized("INVALID_CREDENTIALS", "Invalid email or password")
	}
	accessToken, err := utility.CreateAccessToken(fmt.Sprint(user.ID), user.Email)
	if err != nil {
		return "", "", api.Internal("ACCESS_TOKEN_CREATION_FAILED", "Failed to create access token").WithCause(err)
	}
	refreshToken, err := utility.CreateRefreshToken(fmt.Sprint(user.ID), user.Email)
	if err != nil {
		return "", "", api.Internal("REFRESH_TOKEN_CREATION_FAILED", "Failed to create refresh token").WithCause(err)
	}

	return accessToken, refreshToken, nil
}

func (as *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {

	// 1️⃣ Validate refresh token
	tokenData, err := utility.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", api.Unauthorized("INVALID_REFRESH_TOKEN", "Invalid refresh token")
	}

	userID := tokenData.UserID

	// 2️⃣ Check DB
	rt, err := as.authRepo.GetRefreshToken(ctx, hashOTP(refreshToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", "", api.Unauthorized("REFRESH_TOKEN_NOT_FOUND", "Refresh token not found")
		}
		return "", "", api.Internal("REFRESH_TOKEN_LOOKUP_FAILED", "Failed to validate refresh token").WithCause(err)
	}

	// 3️⃣ Expiry check
	if time.Now().After(rt.ExpiresAt.Time) {
		return "", "", api.Unauthorized("REFRESH_TOKEN_EXPIRED", "Refresh token expired")
	}

	// 4️⃣ Generate tokens via utility
	newAccess, err := utility.GenerateAccessToken(userID)
	if err != nil {
		return "", "", api.Internal("ACCESS_TOKEN_FAILED", "Failed to generate access token").WithCause(err)
	}

	newRefresh, err := utility.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", api.Internal("REFRESH_TOKEN_FAILED", "Failed to generate refresh token").WithCause(err)
	}

	// 5️⃣ Rotation
	_ = as.authRepo.DeleteRefreshToken(ctx, hashOTP(refreshToken))

	err = as.authRepo.CreateRefreshToken(ctx, userID, hashOTP(newRefresh), time.Now().Add(7*24*time.Hour))
	if err != nil {
		return "", "", api.Internal("REFRESH_TOKEN_SAVE_FAILED", "Failed to save new refresh token").WithCause(err)
	}

	return newAccess, newRefresh, nil
}

func (as *AuthService) GetProfile(ctx context.Context, userID int64) (*dto.UserProfileDTO, error) {
	user, err := as.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, api.NotFound("USER_NOT_FOUND", "User not found")
		}
		return nil, api.Internal("PROFILE_RETRIEVAL_FAILED", "Failed to retrieve profile").WithCause(err)
	}
	profile := &dto.UserProfileDTO{
		ID:       fmt.Sprint(user.ID),
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name.String,
		Course:   user.Course.String,
		YOP:      user.Yop.Int32,
	}
	return profile, nil
}
