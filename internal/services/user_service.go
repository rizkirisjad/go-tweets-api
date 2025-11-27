package services

import (
	"context"
	"go-tweets/internal/config"
	"go-tweets/internal/dto"
	"go-tweets/internal/models"
	"go-tweets/internal/repositories"
	appErr "go-tweets/pkg/errors"
	"go-tweets/pkg/jwt"
	"go-tweets/pkg/refreshtoken"

	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (int64, error)
	Login(ctx context.Context, req *dto.LoginRequest) (string, string, error)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, userId int64) (string, string, error)
}

type userService struct {
	cfg        *config.Config
	repository repositories.UserRepository
}

func NewUserService(cfg *config.Config, repository repositories.UserRepository) UserService {
	return &userService{
		cfg:        cfg,
		repository: repository,
	}
}

func (s *userService) Register(ctx context.Context, req *dto.RegisterRequest) (int64, error) {
	// check if user already exists
	userExist, err := s.repository.FindByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return 0, appErr.New("failed checking user", http.StatusInternalServerError, err)
	}

	if userExist != nil {
		return 0, appErr.New("user already exists", http.StatusBadRequest, nil)
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, appErr.New("failed hashing password", http.StatusInternalServerError, err)
	}

	// create user
	now := time.Now()
	user := models.User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  string(hashedPassword),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.repository.Create(ctx, &user)
	if err != nil {
		return 0, appErr.New("failed creating user", http.StatusInternalServerError, err)
	}

	return user.ID, nil
}

func (s *userService) Login(ctx context.Context, req *dto.LoginRequest) (string, string, error) {
	// check if user exists
	userExists, err := s.repository.FindByEmailOrUsername(ctx, req.Email, "")
	if err != nil {
		return "", "", appErr.New("error checking user", http.StatusInternalServerError, err)
	}

	if userExists == nil {
		return "", "", appErr.New("user not found", http.StatusNotFound, nil)
	}

	err = bcrypt.CompareHashAndPassword([]byte(userExists.Password), []byte(req.Password))
	if err != nil {
		return "", "", appErr.New("invalid email or password", http.StatusUnauthorized, err)
	}

	// generate access token
	token, err := jwt.CreateToken(userExists.ID, userExists.Username, s.cfg.JWTSecret)
	if err != nil {
		return "", "", appErr.New("failed creating token", http.StatusInternalServerError, err)
	}

	// get refresh token if exists
	now := time.Now()
	refreshTokenExists, err := s.repository.GetRefreshToken(ctx, userExists.ID, now)
	if err != nil {
		return "", "", appErr.New("failed retrieving refresh token", http.StatusInternalServerError, err)
	}

	if refreshTokenExists != nil {
		return token, refreshTokenExists.RefreshToken, nil
	}

	// generate and store refresh token
	refreshToken, err := refreshtoken.GenerateRefreshToken()
	if err != nil {
		return "", "", appErr.New("failed generating refresh token", http.StatusInternalServerError, err)
	}

	err = s.repository.StoreRefreshToken(ctx, &models.RefreshToken{
		UserID:       userExists.ID,
		RefreshToken: refreshToken,
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour),
	})

	if err != nil {
		return "", "", appErr.New("failed saving refresh token", http.StatusInternalServerError, err)
	}

	// return
	return token, refreshToken, nil
}

func (s *userService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest, userID int64) (string, string, error) {
	// check user exists
	userExists, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return "", "", appErr.New("error checking user", http.StatusInternalServerError, err)
	}

	if userExists == nil {
		return "", "", appErr.New("user not found", http.StatusNotFound, err)
	}
	// get refresh token by user id
	refreshTokenExists, err := s.repository.GetRefreshToken(ctx, userExists.ID, time.Now())
	if err != nil {
		appErr.New("failed generating refresh token", http.StatusInternalServerError, err)
	}

	if refreshTokenExists == nil {
		return "", "", appErr.New("refresh token was expired", http.StatusUnauthorized, err)
	}

	// check refresh token is match with request body
	if req.RefreshToken != refreshTokenExists.RefreshToken {
		return "", "", appErr.New("refresh token not found", http.StatusUnauthorized, err)
	}

	// generate new token
	token, err := jwt.CreateToken(userExists.ID, userExists.Username, s.cfg.JWTSecret)
	if err != nil {
		return "", "", appErr.New("failed creating token", http.StatusInternalServerError, err)
	}

	// delete old refresh token and generate new refresh token
	err = s.repository.DeleteRefreshTokenByUserID(ctx, userExists.ID)
	if err != nil {
		return "", "", appErr.New("failed deleting refresh token", http.StatusInternalServerError, err)
	}

	refreshToken, err := refreshtoken.GenerateRefreshToken()
	if err != nil {
		return "", "", appErr.New("failed generating refresh token", http.StatusInternalServerError, err)
	}

	now := time.Now()

	err = s.repository.StoreRefreshToken(ctx, &models.RefreshToken{
		UserID:       userExists.ID,
		RefreshToken: refreshToken,
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiredAt:    time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return "", "", appErr.New("failed saving refresh token", http.StatusInternalServerError, err)
	}

	// return
	return token, refreshToken, nil

}
