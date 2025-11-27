package handlers

import (
	"go-tweets/internal/dto"
	"go-tweets/internal/helpers"
	"go-tweets/internal/middlewares"
	"go-tweets/internal/services"
	appErr "go-tweets/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CommentHandler struct {
	api      *gin.Engine
	validate *validator.Validate
	service  services.CommentService
}

func NewCommentHandler(api *gin.Engine, validate *validator.Validate, service services.CommentService) *CommentHandler {
	return &CommentHandler{
		api:      api,
		validate: validate,
		service:  service,
	}
}

func (h *CommentHandler) RouteList(secretKey string) {
	commentAuth := h.api.Group("/comment")
	commentAuth.Use(middlewares.AuthMiddleware(secretKey))
	commentAuth.POST("/", h.CreateComment)
	commentAuth.POST("/action", h.LikeOrUnlikeComment)

}

func (h *CommentHandler) CreateComment(c *gin.Context) {
	var (
		req dto.CommentRequest
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

	err := h.service.CreateComment(ctx, &req, userID)
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "success",
	})
}

func (h *CommentHandler) LikeOrUnlikeComment(c *gin.Context) {
	var (
		req dto.LikeOrUnlikeCommentRequest
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

	err := h.service.LikeOrUnlikeComment(ctx, req.CommentID, userID)
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
