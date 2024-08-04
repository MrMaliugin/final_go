package middleware

import (
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
)

var jwtKey = []byte("your_secret_key")

// Claims структура для JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Auth middleware для проверки JWT токена
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
			tknStr := cookie.Value

			claims := &Claims{}
			tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})

			if err != nil || !tkn.Valid {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
