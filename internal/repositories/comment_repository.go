package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-tweets/internal/models"
	"strings"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *models.Comment) error
	GetCommentByID(ctx context.Context, commentID int64) (*models.Comment, error)
	IsUserAlreadyLikeComment(ctx context.Context, commentID, userID int64) (bool, error)
	DeleteLikeComment(ctx context.Context, commentID, userID int64) error
	CreateLike(ctx context.Context, commentLike *models.CommentLike) error
	GetCommentsByPostIDs(ctx context.Context, postIDs []int64) ([]models.Comment, error)
}

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepository{
		db: db,
	}
}

func (r *commentRepository) Create(ctx context.Context, comment *models.Comment) error {
	query := `INSERT INTO comments (post_id, user_id, content, created_at, updated_at)
		VALUES (?,?,?,?,?)`

	_, err := r.db.ExecContext(ctx, query, comment.PostID, comment.UserID, comment.Content, comment.CreatedAt, comment.UpdatedAt)

	return err
}

func (r *commentRepository) GetCommentByID(ctx context.Context, commentID int64) (*models.Comment, error) {
	query := `SELECT id, post_id, user_id, content, created_at, updated_at
		FROM comments
		WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, commentID)

	var comment models.Comment
	err := row.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &comment, nil
}

func (r *commentRepository) IsUserAlreadyLikeComment(ctx context.Context, commentID, userID int64) (bool, error) {
	query := `SELECT id FROM comment_likes
		WHERE comment_id = ?
		AND user_id = ?`

	row := r.db.QueryRowContext(ctx, query, commentID, userID)

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

func (r *commentRepository) DeleteLikeComment(ctx context.Context, commentID, userID int64) error {
	query := `DELETE FROM comment_likes
		WHERE comment_id = ?
		AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, commentID, userID)
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

func (r *commentRepository) CreateLike(ctx context.Context, commentLike *models.CommentLike) error {
	query := `INSERT INTO comment_likes (comment_id, user_id, created_at, updated_at)
		VALUES (?,?,?,?)`

	_, err := r.db.ExecContext(ctx, query, commentLike.CommentID, commentLike.UserID, commentLike.CreatedAt, commentLike.UpdatedAt)

	return err
}

func (r *commentRepository) GetCommentsByPostIDs(ctx context.Context, postIDs []int64) ([]models.Comment, error) {
	if len(postIDs) == 0 {
		return []models.Comment{}, nil
	}

	placeholders := make([]string, len(postIDs))
	args := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`SELECT c.id, c.post_id, c.user_id, u.username, c.content, c.created_at, c.updated_at, COUNT(cl.id) as like_count
	FROM comments as c
	JOIN users as u ON u.id = c.user_id
	LEFT JOIN comment_likes as cl ON cl.comment_id = c.id
	WHERE c.post_id IN(%s)
	GROUP BY c.id, c.post_id, c.user_id, u.username, c.content, c.created_at, c.updated_at
	ORDER BY like_count DESC`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return []models.Comment{}, err
	}
	defer rows.Close()

	comments := make([]models.Comment, 0)
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Username, &comment.Content, &comment.CreatedAt, &comment.UpdatedAt, &comment.LikeCount)
		if err != nil {
			return []models.Comment{}, err
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}
