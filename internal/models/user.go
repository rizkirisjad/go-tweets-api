package models

import "time"

type (
	User struct {
		ID        int64
		Email     string
		Username  string
		Password  string
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	RefreshToken struct {
		ID           int64
		UserID       int64
		RefreshToken string
		ExpiredAt    time.Time
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}
)
