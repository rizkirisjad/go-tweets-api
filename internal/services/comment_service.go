package services

import (
	"context"
	"go-tweets/internal/config"
	"go-tweets/internal/dto"
	"go-tweets/internal/models"
	"go-tweets/internal/repositories"
	appErr "go-tweets/pkg/errors"
	"net/http"
	"time"
)

type CommentService interface {
	CreateComment(ctx context.Context, req *dto.CommentRequest, userID int64) error
	LikeOrUnlikeComment(ctx context.Context, commentID, userID int64) error
}

type commentService struct {
	cfg               *config.Config
	commentRepository repositories.CommentRepository
	postRepository    repositories.PostRepository
}

func NewCommentService(cfg *config.Config, commentRepository repositories.CommentRepository, postRepository repositories.PostRepository) CommentService {
	return &commentService{
		cfg:               cfg,
		commentRepository: commentRepository,
		postRepository:    postRepository,
	}
}

func (s *commentService) CreateComment(ctx context.Context, req *dto.CommentRequest, userID int64) error {
	// check tweet is exist
	postExists, err := s.postRepository.GetPostByID(ctx, req.PostID)
	if err != nil {
		return appErr.New("failed checking post", http.StatusInternalServerError, err)
	}

	if postExists == nil {
		return appErr.New("post not found", http.StatusNotFound, nil)
	}

	// store comment
	now := time.Now()
	err = s.commentRepository.Create(ctx, &models.Comment{
		PostID:    postExists.ID,
		UserID:    userID,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	})

	if err != nil {
		return appErr.New("failed posting comment", http.StatusInternalServerError, err)
	}

	// return
	return nil
}

func (s *commentService) LikeOrUnlikeComment(ctx context.Context, commentID, userID int64) error {
	// check comment
	commentExists, err := s.commentRepository.GetCommentByID(ctx, commentID)
	if err != nil {
		return appErr.New("failed checking comment", http.StatusInternalServerError, err)
	}

	if commentExists == nil {
		return appErr.New("comment not found", http.StatusNotFound, nil)
	}

	// check user already like comment or not
	isUserAlreadyLikeComment, err := s.commentRepository.IsUserAlreadyLikeComment(ctx, commentID, userID)
	if err != nil {
		appErr.New("failed checking like status", http.StatusInternalServerError, err)
	}

	// if user already liked comment, delete data
	if isUserAlreadyLikeComment {
		err := s.commentRepository.DeleteLikeComment(ctx, commentID, userID)
		if err != nil {
			return appErr.New("failed to unlike comment", http.StatusInternalServerError, err)
		}
	} else {
		// else, store data
		now := time.Now()
		err := s.commentRepository.CreateLike(ctx, &models.CommentLike{
			CommentID: commentID,
			UserID:    userID,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return appErr.New("failed to like comment", http.StatusInternalServerError, err)
		}
	}

	// return
	return nil
}
