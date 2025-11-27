package models

import "time"

type (
	Post struct {
		ID        int64
		UserID    int64
		Title     string
		Content   string
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt time.Time
	}

	PostLike struct {
		ID        int64
		PostID    int64
		UserID    int64
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	PostWithUser struct {
		ID        int64
		UserID    int64
		Username  string
		Title     string
		Content   string
		LikeCount int64
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt time.Time
	}
)
