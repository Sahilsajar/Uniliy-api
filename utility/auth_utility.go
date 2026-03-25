package utility

import (
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func CreateAccessToken(userID, email string) (string, error) {
	secretKey := []byte(os.Getenv("JWT_ACCESS_SECRET"))

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"type":    "access",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     uuid.NewString(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func CreateRefreshToken(userID, email string) (string, error) {
	secretKey := []byte(os.Getenv("JWT_REFRESH_SECRET"))

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"type":    "refresh",
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
		"jti":     uuid.NewString(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
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
