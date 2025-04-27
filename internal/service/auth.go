package service

import (
	"context"
	"database/sql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
)

var jwtSecretKey = os.Getenv("JWT_SECRET_KEY")

type AuthService struct {
	userRepo *storage.UserRepository
}

func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{userRepo: storage.NewUserRepository(db)}
}

func (s *AuthService) Register(ctx context.Context, request dto.PostRegisterJSONRequestBody) (*models.User, error) {
	if !isValidRole(dto.UserRole(request.Role)) {
		return nil, models.ErrIncorrectUserRole
	}

	hashedPassword, err := hashPassword(request.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    request.Email,
		Password: hashedPassword,
		Role:     dto.UserRole(request.Role),
	}

	if err = s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func isValidRole(role dto.UserRole) bool {
	return role == dto.UserRoleEmployee || role == dto.UserRoleModerator
}

func (s *AuthService) DummyLogin(request dto.PostDummyLoginJSONRequestBody) (*dto.Token, error) {
	if !isValidRole(dto.UserRole(request.Role)) {
		return nil, models.ErrIncorrectUserRole
	}

	token, err := generateToken(dto.UserRole(request.Role))
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *AuthService) Login(ctx context.Context, request dto.PostLoginJSONRequestBody) (*dto.Token, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, request.Email)
	if err != nil {
		return nil, models.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return nil, models.ErrWrongPassword
	}

	token, err := generateToken(user.Role)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func generateToken(role dto.UserRole) (*dto.Token, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.TokenClaims{Role: role, RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		ID:        uuid.New().String(),
	}})

	tokenString, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return nil, err
	}
	return &tokenString, nil
}
