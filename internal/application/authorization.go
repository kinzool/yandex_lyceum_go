package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

const (
	SecretKey = "super_secret_signature"
)

func GenerateJWT(user_id int, login string) (string, error) {
	const hmacSampleSecret = "super_secret_signature"
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user_id,
		"login":   login,
		"nbf":     now.Add(5 * time.Second).Unix(),
		"exp":     now.Add(10 * time.Minute).Unix(),
		"iat":     now.Unix(),
	})
	tokenString, err := token.SignedString([]byte(hmacSampleSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func GeneratePassword(password string) (string, error) {
	if len(password) > 72 {
		password = password[:72]
	}
	saltedBytes := []byte(password)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", nil
	}
	hash := string(hashedBytes[:])
	return hash, nil
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, `{"error": "Missing token"}`, http.StatusUnauthorized)
			return
		}
		token, err := jwt.Parse(cookie.Value, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(SecretKey), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusForbidden)
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
