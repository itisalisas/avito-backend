package handlers

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/itisalisas/avito-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"os"
	"time"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var newUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if newUser.Role != string(models.Employee) && newUser.Role != string(models.Moderator) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO - save to DB with hashed password
	_, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user := models.User{
		ID:    uuid.New(),
		Email: newUser.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(models.User{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	secretKey := os.Getenv("SECRET_KEY")
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// TODO - get user from DB
	user := models.User{}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role":   string(user.Role),
		"expire": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	secretKey := os.Getenv("SECRET_KEY")
	var request struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if request.Role != string(models.Employee) && request.Role != string(models.Moderator) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role":   request.Role,
		"expire": time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
