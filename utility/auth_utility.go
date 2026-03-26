package utility

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type TokenData struct {
	UserID int32
	Type   string
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func IsEmail(s string) bool {
	return strings.Contains(s, "@")
}

func ValidateToken(tokenStr string) (string, string, error) {
	secretKey := []byte(os.Getenv("JWT_ACCESS_SECRET"))
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})

	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(string)
		email, emailOk := claims["email"].(string)
		if !ok || !emailOk {
			return "", "", jwt.ErrInvalidKey
		}
		if !ok {
			return "", "", jwt.ErrInvalidKeyType
		}
		return userID, email, nil
	}

	return "", "", jwt.ErrInvalidKey
}

// 🔐 get secrets
func getAccessSecret() ([]byte, error) {
	secret := os.Getenv("JWT_ACCESS_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("missing access secret")
	}
	return []byte(secret), nil
}

func getRefreshSecret() ([]byte, error) {
	secret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("missing refresh secret")
	}
	return []byte(secret), nil
}

// ✅ GENERATE ACCESS TOKEN
func GenerateAccessToken(userID int32) (string, error) {
	secret, err := getAccessSecret()
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"type":    "access",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	})

	return token.SignedString(secret)
}

// ✅ GENERATE REFRESH TOKEN
func GenerateRefreshToken(userID int32) (string, error) {
	secret, err := getRefreshSecret()
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	})

	return token.SignedString(secret)
}

// ✅ VALIDATE REFRESH TOKEN
func ValidateRefreshToken(tokenStr string) (*TokenData, error) {
	secret, err := getRefreshSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	id, ok := claims["user_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid user_id")
	}

	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	return &TokenData{
		UserID: int32(id),
		Type:   tokenType,
	}, nil
}
