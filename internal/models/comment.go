package models

import "time"

type (
	Comment struct {
		ID        int64
		PostID    int64
		UserID    int64
		Username  string
		Content   string
		LikeCount int64
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	CommentLike struct {
		ID        int64
		CommentID int64
		UserID    int64
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)
