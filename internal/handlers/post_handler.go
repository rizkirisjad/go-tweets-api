package handlers

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/helpers"
	"go-tweets/internal/middlewares"
	"go-tweets/internal/services"
	appErr "go-tweets/pkg/errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type PostHandler struct {
	api      *gin.Engine
	validate *validator.Validate
	service  services.PostService
}

func NewPostHandler(api *gin.Engine, validate *validator.Validate, service services.PostService) *PostHandler {
	return &PostHandler{
		api:      api,
		validate: validate,
		service:  service,
	}
}

func (h *PostHandler) RouteList(secretKey string) {
	tweetRoute := h.api.Group("/tweets")
	tweetRoute.Use(middlewares.AuthMiddleware(secretKey))
	tweetRoute.POST("/", h.CreatePost)
	tweetRoute.PUT("/:post_id", h.UpdatePost)
	tweetRoute.DELETE("/:post_id", h.DeletePost)
	tweetRoute.POST("/action", h.LikeOrUnlikePost)

	tweetRouteWithoutAuth := h.api.Group("/tweets")
	tweetRouteWithoutAuth.GET("/", h.GetAllPosts)
	tweetRouteWithoutAuth.GET("/:post_id", h.DetailPost)
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var (
		req dto.CreateOrUpdatePostRequest
		ctx = c.Request.Context()
	)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": helpers.FormatValidationError(errs, req),
			})
			return
		}
	}

	userID := c.GetInt64("userID")

	postID, err := h.service.CreatePost(ctx, &req, userID)
	if err != nil {
		// if it's an AppError, log detailed and return its status & message
		if appErr, ok := err.(*appErr.AppError); ok {
			appErr.Log() // uses utils.ErrorHandler under the hood
			c.JSON(appErr.StatusCode, gin.H{"message": appErr.Message})
			return
		}

		// fallback for unexpected errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected server error"})
		return
	}

	c.JSON(http.StatusCreated, dto.CreateOrUpdatePostResponse{
		ID: postID,
	})
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	var (
		req dto.CreateOrUpdatePostRequest
		ctx = c.Request.Context()
	)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": helpers.FormatValidationError(errs, req),
			})
			return
		}
	}

	userID := c.GetInt64("userID")
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = h.service.UpdatePost(ctx, &req, postID, userID)
	if err != nil {
		// if it's an AppError, log detailed and return its status & message
		if appErr, ok := err.(*appErr.AppError); ok {
			appErr.Log() // uses utils.ErrorHandler under the hood
			c.JSON(appErr.StatusCode, gin.H{"message": appErr.Message})
			return
		}

		// fallback for unexpected errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected server error"})
		return
	}

	c.JSON(http.StatusOK, dto.CreateOrUpdatePostResponse{
		ID: postID,
	})
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	var (
		ctx       = c.Request.Context()
		userID    = c.GetInt64("userID")
		postIDStr = c.Param("post_id")
	)

	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = h.service.DeletePost(ctx, postID, userID)
	if err != nil {
		// if it's an AppError, log detailed and return its status & message
		if appErr, ok := err.(*appErr.AppError); ok {
			appErr.Log() // uses utils.ErrorHandler under the hood
			c.JSON(appErr.StatusCode, gin.H{"message": appErr.Message})
			return
		}

		// fallback for unexpected errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully deleted tweet",
	})
}

func (h *PostHandler) LikeOrUnlikePost(c *gin.Context) {
	var (
		req dto.LikeOrUnlikePostRequest
		ctx = c.Request.Context()
	)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := h.validate.Struct(req); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"errors": helpers.FormatValidationError(errs, req),
			})
			return
		}
	}

	userID := c.GetInt64("userID")

	err := h.service.LikeOrUnlikePost(ctx, req.PostID, userID)
	if err != nil {
		// if it's an AppError, log detailed and return its status & message
		if appErr, ok := err.(*appErr.AppError); ok {
			appErr.Log() // uses utils.ErrorHandler under the hood
			c.JSON(appErr.StatusCode, gin.H{"message": appErr.Message})
			return
		}

		// fallback for unexpected errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func (h *PostHandler) DetailPost(c *gin.Context) {
	ctx := c.Request.Context()
	postIDStr := c.Param("post_id")
	postID, err := strconv.ParseInt(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.DetailPost(ctx, postID)
	if err != nil {
		// if it's an AppError, log detailed and return its status & message
		if appErr, ok := err.(*appErr.AppError); ok {
			appErr.Log() // uses utils.ErrorHandler under the hood
			c.JSON(appErr.StatusCode, gin.H{"message": appErr.Message})
			return
		}

		// fallback for unexpected errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *PostHandler) GetAllPosts(c *gin.Context) {
	ctx := c.Request.Context()
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "5")

	page, _ := strconv.ParseInt(pageStr, 10, 64)
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	param := dto.GetAllPostsRequest{
		Limit: limit,
		Page:  page,
	}

	result, err := h.service.GetAllPosts(ctx, &param)
	if err != nil {
		// if it's an AppError, log detailed and return its status & message
		if appErr, ok := err.(*appErr.AppError); ok {
			appErr.Log() // uses utils.ErrorHandler under the hood
			c.JSON(appErr.StatusCode, gin.H{"message": appErr.Message})
			return
		}

		// fallback for unexpected errors
		c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected server error"})
		return
	}

	c.JSON(http.StatusOK, result)
}
