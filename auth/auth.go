package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jmoiron/sqlx"
)

var jwtKey = []byte("your_secret_key")

type Credentials struct {
	Password string `json:"password"`
}

type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.StandardClaims
}

// SigninHandler обрабатывает вход пользователя и возвращает JWT токен
func SigninHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
			return
		}

		expectedPassword := os.Getenv("TODO_PASSWORD")
		if expectedPassword == "" {
			http.Error(w, "Пароль не задан в переменной окружения", http.StatusInternalServerError)
			return
		}

		if creds.Password != expectedPassword {
			http.Error(w, "Неверный пароль", http.StatusUnauthorized)
			return
		}

		expirationTime := time.Now().Add(8 * time.Hour)
		claims := &Claims{
			PasswordHash: creds.Password,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "Ошибка создания токена", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})

		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}

// Auth проверяет JWT токен
func Auth(next http.HandlerFunc, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			tokenStr := cookie.Value
			claims := &Claims{}

			token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}
