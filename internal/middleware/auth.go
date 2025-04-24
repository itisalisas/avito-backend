package middleware

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
	"time"
)

type userRole string

func CheckAuth() func(next http.Handler) http.Handler {
	secretKey := os.Getenv("SECRET_KEY")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/login" || r.URL.Path == "/register" || r.URL.Path == "/dummyLogin" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				if time.Now().Unix() > int64(claims["expire"].(float64)) {
					http.Error(w, "Token expired", http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), userRole("userRole"), claims["role"])
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			}
		})
	}
}
