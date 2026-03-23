package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/unilly-api/repositories"
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
		return err
	}
	if !allowed {
		return fmt.Errorf("too many requests, please wait before retrying")
	}

	// Generate OTP
	otp, err := GenerateOTP(6)
	if err != nil {
		return fmt.Errorf("failed to generate otp: %w", err)
	}
	otpHash := hashOTP(otp)
	expiresAt := pgtype.Timestamp{Time: time.Now().Add(10 * time.Minute), Valid: true}

	//  Save OTP FIRST
	err = as.authRepo.SaveOTP(ctx, email, otpHash, expiresAt)
	fmt.Print(err)
	if err != nil {
		fmt.Print("lun")
		return fmt.Errorf("failed to save otp: %w", err)
	}

	// Send email
	if err := as.sendEmail(ctx, email, otp); err != nil {
		// mark OTP as failed or delete
		return fmt.Errorf("failed to send email: %w", err)
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

func (as *AuthService) VerifyOTP(
	ctx context.Context,
	email string,
	otp string,
) (bool, error) {

	record, err := as.authRepo.GetLatestOTP(ctx, email)
	if err != nil {
		return false, err
	}

	// 1. Expiry check
	if time.Now().After(record.ExpiresAt.Time) {
		return false, fmt.Errorf("otp expired")
	}

	// 2. Attempt limit
	if record.Attempts.Int32 >= record.MaxAttempts.Int32 {
		return false, fmt.Errorf("too many attempts")
	}

	// 3. Compare hash
	if hashOTP(otp) != record.OtpHash {
		_ = as.authRepo.IncrementOTPAttempts(ctx, record.Email)
		return false, fmt.Errorf("invalid otp")
	}

	// 4. Success → invalidate
	err = as.authRepo.MarkOTPAsVerified(ctx, record.Email)
	if err != nil {
		return false, err
	}

	return true, nil
}