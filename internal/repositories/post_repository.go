package repositories

import (
	"context"
	"database/sql"
	"errors"
	"go-tweets/internal/dto"
	"go-tweets/internal/models"
	"time"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetPostByID(ctx context.Context, postID int64) (*models.PostWithUser, error)
	Update(ctx context.Context, post *models.Post, postID int64) error
	SoftDelete(ctx context.Context, postID int64, now time.Time) error
	IsUserAlreadyLikePost(ctx context.Context, postID, userID int64) (bool, error)
	DeleteLikePost(ctx context.Context, postId, userID int64) error
	StoreLikePost(ctx context.Context, posLike *models.PostLike) error
	TotalPost(ctx context.Context) (int64, error)
	GetAllPosts(ctx context.Context, param *dto.GetAllPostsRequest, offset int) ([]models.PostWithUser, error)
}

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) PostRepository {
	return &postRepository{
		db: db,
	}
}

func (r *postRepository) Create(ctx context.Context, post *models.Post) error {
	query := `INSERT INTO posts (user_id, title, content, created_at, updated_at)
		VALUES (?,?,?,?,?)`

	result, err := r.db.ExecContext(ctx, query, post.UserID, post.Title, post.Content, post.CreatedAt, post.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	post.ID = id

	return nil

}

func (r *postRepository) GetPostByID(ctx context.Context, postID int64) (*models.PostWithUser, error) {
	query := `SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at, u.username, COUNT(pl.id) as like_count
		FROM posts as p
		JOIN users as u ON p.user_id = u.id
		LEFT JOIN post_likes as pl ON pl.post_id = p.id
		WHERE p.id = ?
		AND p.deleted_at IS NULL
		GROUP BY p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at, u.username`

	row := r.db.QueryRowContext(ctx, query, postID)

	var post models.PostWithUser
	err := row.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.UpdatedAt, &post.Username, &post.LikeCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &post, nil
}

func (r *postRepository) Update(ctx context.Context, post *models.Post, postID int64) error {
	query := `UPDATE posts SET title = ?, content = ?, updated_at = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, post.UpdatedAt, postID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("nothing to update")
	}

	return nil
}

func (r *postRepository) SoftDelete(ctx context.Context, postID int64, now time.Time) error {
	query := `UPDATE posts SET deleted_at = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, now, postID)
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

func (r *postRepository) IsUserAlreadyLikePost(ctx context.Context, postID, userID int64) (bool, error) {
	query := `SELECT id FROM post_likes
		WHERE post_id = ?
		AND user_id = ?`

	row := r.db.QueryRowContext(ctx, query, postID, userID)

	var id int64
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *postRepository) DeleteLikePost(ctx context.Context, postId, userID int64) error {
	query := `DELETE FROM post_likes
		WHERE post_id = ?
		AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, postId, userID)
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

func (r postRepository) StoreLikePost(ctx context.Context, postLike *models.PostLike) error {
	query := `INSERT INTO post_likes (post_id, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, postLike.PostID, postLike.UserID, postLike.CreatedAt, postLike.UpdatedAt)

	return err
}

func (r *postRepository) TotalPost(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(id) FROM posts
		WHERE deleted_at IS NULL`

	var total int64
	err := r.db.QueryRowContext(ctx, query).Scan(&total)

	return total, err
}

func (r *postRepository) GetAllPosts(ctx context.Context, param *dto.GetAllPostsRequest, offset int) ([]models.PostWithUser, error) {
	query := `SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at, u.username, COUNT(pl.id) as like_count
		FROM posts as p
		JOIN users as u ON p.user_id = u.id
		LEFT JOIN post_likes as pl ON pl.post_id = p.id
		WHERE p.deleted_at IS NULL
		GROUP BY p.id, p.user_id, p.title, p.content, p.created_at, p.updated_at, u.username
		ORDER BY created_at DESC
		LIMIT ?
		OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, param.Limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.PostWithUser, 0)
	for rows.Next() {
		var data models.PostWithUser
		err := rows.Scan(&data.ID, &data.UserID, &data.Title, &data.Content, &data.CreatedAt, &data.UpdatedAt, &data.Username, &data.LikeCount)
		if err != nil {
			return nil, err
		}

		result = append(result, data)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
