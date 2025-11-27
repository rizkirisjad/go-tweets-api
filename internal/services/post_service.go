package services

import (
	"context"
	"go-tweets/internal/config"
	"go-tweets/internal/dto"
	"go-tweets/internal/models"
	"go-tweets/internal/repositories"
	appErr "go-tweets/pkg/errors"
	"math"
	"net/http"
	"time"
)

type PostService interface {
	CreatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, userID int64) (int64, error)
	UpdatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, postID, userID int64) error
	DeletePost(ctx context.Context, postID, userID int64) error
	LikeOrUnlikePost(ctx context.Context, postID, userID int64) error
	DetailPost(ctx context.Context, postID int64) (*dto.DetailPostResponse, error)
	GetAllPosts(ctx context.Context, param *dto.GetAllPostsRequest) (*dto.GetAllPostsResponse, error)
}

type postService struct {
	cfg               *config.Config
	repository        repositories.PostRepository
	commentRepository repositories.CommentRepository
}

func NewPostService(cfg *config.Config, repository repositories.PostRepository, commentRepository repositories.CommentRepository) PostService {
	return &postService{
		cfg:               cfg,
		repository:        repository,
		commentRepository: commentRepository,
	}
}

func (s *postService) CreatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, userID int64) (int64, error) {
	// store post
	now := time.Now()
	post := models.Post{
		UserID:    userID,
		Title:     req.Title,
		Content:   req.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.repository.Create(ctx, &post)
	if err != nil {
		return 0, appErr.New("failed creating post", http.StatusInternalServerError, err)
	}

	return post.ID, nil
}

func (s *postService) UpdatePost(ctx context.Context, req *dto.CreateOrUpdatePostRequest, postID, userID int64) error {
	// Check Post by ID
	postExists, err := s.repository.GetPostByID(ctx, postID)

	if err != nil {
		return appErr.New("error checking tweet", http.StatusInternalServerError, err)
	}

	if postExists == nil {
		return appErr.New("post not found", http.StatusNotFound, nil)
	}

	if postExists.UserID != userID {
		return appErr.New("post not found", http.StatusNotFound, err)
	}

	// update post
	err = s.repository.Update(ctx, &models.Post{
		Title:     req.Title,
		Content:   req.Content,
		UpdatedAt: time.Now(),
	}, postID)

	if err != nil {
		return appErr.New("failed update post", http.StatusInternalServerError, err)
	}

	return nil
}

func (s *postService) DeletePost(ctx context.Context, postID, userID int64) error {
	// check post
	postExists, err := s.repository.GetPostByID(ctx, postID)

	if err != nil {
		return appErr.New("error checking post", http.StatusInternalServerError, err)
	}

	if postExists == nil {
		return appErr.New("post not found", http.StatusNotFound, nil)
	}

	if postExists.UserID != userID {
		return appErr.New("post not found", http.StatusNotFound, err)
	}

	// soft delete post
	err = s.repository.SoftDelete(ctx, postID, time.Now())
	if err != nil {
		return appErr.New("failed delete post", http.StatusInternalServerError, err)
	}

	// return
	return nil

}

func (s *postService) LikeOrUnlikePost(ctx context.Context, postID, userID int64) error {
	// check post
	postExists, err := s.repository.GetPostByID(ctx, postID)

	if err != nil {
		return appErr.New("failed checking post", http.StatusInternalServerError, err)
	}

	if postExists == nil {
		return appErr.New("post not found", http.StatusNotFound, nil)
	}

	// check user already like or not
	isUserAlreadyLikePost, err := s.repository.IsUserAlreadyLikePost(ctx, postID, userID)
	if err != nil {
		appErr.New("failed checking like status", http.StatusInternalServerError, err)
	}

	// if user already like, delete data
	if isUserAlreadyLikePost {
		err := s.repository.DeleteLikePost(ctx, postID, userID)
		if err != nil {
			return appErr.New("failed to unlike post", http.StatusInternalServerError, err)
		}
	} else {
		// else, store data
		now := time.Now()
		err := s.repository.StoreLikePost(ctx, &models.PostLike{
			PostID:    postID,
			UserID:    userID,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return appErr.New("failed to like post", http.StatusInternalServerError, err)
		}
	}

	//return
	return nil
}

func (s *postService) DetailPost(ctx context.Context, postID int64) (*dto.DetailPostResponse, error) {
	// get post by id
	post, err := s.repository.GetPostByID(ctx, postID)

	if err != nil {
		return nil, appErr.New("failed checking post", http.StatusInternalServerError, err)
	}

	if post == nil {
		return nil, appErr.New("post not found", http.StatusNotFound, nil)
	}

	// get all comments related with post
	postIDs := []int64{post.ID}
	comments, err := s.commentRepository.GetCommentsByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, appErr.New("failed to get comments", http.StatusInternalServerError, err)
	}

	// mapping comments with post
	commentDTOs := make([]dto.Comment, 0)
	for _, comment := range comments {
		commentDTOs = append(commentDTOs, dto.Comment{
			ID:        comment.ID,
			Username:  comment.Username,
			Content:   comment.Content,
			LikeCount: comment.LikeCount,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
		})
	}

	// set response
	return &dto.DetailPostResponse{
		ID:        post.ID,
		Username:  post.Username,
		Title:     post.Title,
		Content:   post.Content,
		LikeCount: post.LikeCount,
		Comments:  commentDTOs,
		CreatedAt: post.CreatedAt.Format(time.RFC3339),
		UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *postService) GetAllPosts(ctx context.Context, param *dto.GetAllPostsRequest) (*dto.GetAllPostsResponse, error) {
	// get total post
	totalPosts, err := s.repository.TotalPost(ctx)
	if err != nil {
		return nil, appErr.New("failed to get total posts", http.StatusInternalServerError, err)
	}

	// get all post
	offset := param.Limit * (param.Page - 1)
	posts, err := s.repository.GetAllPosts(ctx, param, int(offset))
	if err != nil {
		return nil, appErr.New("failed to get posts", http.StatusInternalServerError, err)
	}

	// get all comments related with posts
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	comments, err := s.commentRepository.GetCommentsByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, appErr.New("failed to get comments", http.StatusInternalServerError, err)
	}

	// mapping all comments with post based on key value of post_id
	commentDTOs := make(map[int64][]dto.Comment)
	for _, comment := range comments {
		commentDTOs[comment.PostID] = append(commentDTOs[comment.PostID], dto.Comment{
			ID:        comment.ID,
			Username:  comment.Username,
			Content:   comment.Content,
			LikeCount: comment.LikeCount,
			CreatedAt: comment.CreatedAt.Format(time.RFC3339),
			UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
		})
	}

	// mapping response
	var data []dto.DetailPostResponse
	for _, post := range posts {
		comments := commentDTOs[post.ID]
		if comments == nil {
			comments = []dto.Comment{}
		}

		data = append(data, dto.DetailPostResponse{
			ID:        post.ID,
			Username:  post.Username,
			Title:     post.Title,
			Content:   post.Content,
			LikeCount: post.LikeCount,
			Comments:  comments,
			CreatedAt: post.CreatedAt.Format(time.RFC3339),
			UpdatedAt: post.UpdatedAt.Format(time.RFC3339),
		})
	}

	totalPages := int64(math.Ceil(float64(totalPosts) / float64(param.Limit)))

	// return
	result := dto.GetAllPostsResponse{
		TotalPage:   totalPages,
		CurrentPage: param.Page,
		Limit:       param.Limit,
		Data:        data,
	}

	return &result, nil
}
