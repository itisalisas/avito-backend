package auth

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"github.com/itisalisas/avito-backend/internal/models"
	"github.com/itisalisas/avito-backend/internal/storage"
)

var jwtSecretKey = os.Getenv("JWT_SECRET_KEY")

type Service struct {
	userRepo storage.UserRepositoryInterface
}

func NewAuthService(userRepo storage.UserRepositoryInterface) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) Register(ctx context.Context, request dto.PostRegisterJSONRequestBody) (*models.User, error) {
	_, err := s.userRepo.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := s.userRepo.Rollback()
		if err != nil {
			log.Fatalf("Error while rolling back transaction: %s", err)
		}
	}()

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

	if err = s.userRepo.Commit(); err != nil {
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

func (s *Service) DummyLogin(request dto.PostDummyLoginJSONRequestBody) (*dto.Token, error) {
	if !isValidRole(dto.UserRole(request.Role)) {
		return nil, models.ErrIncorrectUserRole
	}

	token, err := generateToken(dto.UserRole(request.Role))
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) Login(ctx context.Context, request dto.PostLoginJSONRequestBody) (*dto.Token, error) {
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
