package repositories

import (
	"context"
	"database/sql"
	"errors"
	"go-tweets/internal/models"
	"time"
)

type UserRepository interface {
	FindByEmailOrUsername(ctx context.Context, email, username string) (*models.User, error)
	Create(ctx context.Context, model *models.User) error
	GetRefreshToken(ctx context.Context, userID int64, now time.Time) (*models.RefreshToken, error)
	StoreRefreshToken(ctx context.Context, model *models.RefreshToken) error
	FindByID(ctx context.Context, userID int64) (*models.User, error)
	DeleteRefreshTokenByUserID(ctx context.Context, userID int64) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) FindByEmailOrUsername(ctx context.Context, email, username string) (*models.User, error) {
	query := `SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE email = ?
		OR username = ? 
	`
	row := r.db.QueryRowContext(ctx, query, email, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (email, username, password, created_at, updated_at)
		VALUES (?,?,?,?,?)`

	result, err := r.db.ExecContext(ctx, query, user.Email, user.Username, user.Password, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id

	return nil
}

func (r *userRepository) GetRefreshToken(ctx context.Context, userID int64, now time.Time) (*models.RefreshToken, error) {
	query := `SELECT id, user_id, refresh_token, expired_at
	FROM refresh_tokens
	WHERE user_id = ? AND expired_at >= ?`

	row := r.db.QueryRowContext(ctx, query, userID, now)

	var refreshToken models.RefreshToken
	err := row.Scan(&refreshToken.ID, &refreshToken.UserID, &refreshToken.RefreshToken, &refreshToken.ExpiredAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &refreshToken, nil
}

func (r *userRepository) StoreRefreshToken(ctx context.Context, refreshToken *models.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (user_id, refresh_token, created_at, updated_at, expired_at)
		VALUES (?,?,?,?,?)`

	_, err := r.db.ExecContext(ctx, query, refreshToken.UserID, refreshToken.RefreshToken, refreshToken.CreatedAt, refreshToken.UpdatedAt, refreshToken.ExpiredAt)

	return err
}

func (r *userRepository) FindByID(ctx context.Context, userID int64) (*models.User, error) {
	query := `SELECT id, username, email, created_at, updated_at
		FROM users WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, userID)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) DeleteRefreshTokenByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM refresh_tokens
		WHERE user_id = ?`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("nothing to delete")
	}

	return nil
}
