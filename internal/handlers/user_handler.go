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

type UserHandler struct {
	api      *gin.Engine
	validate *validator.Validate
	service  services.UserService
}

func NewUserHandler(api *gin.Engine, validate *validator.Validate, service services.UserService) *UserHandler {
	return &UserHandler{
		api:      api,
		validate: validate,
		service:  service,
	}
}

func (h *UserHandler) RouteList(secretKey string) {
	authRoute := h.api.Group("/auth")
	authRoute.POST("/register", h.Register)
	authRoute.POST("/login", h.Login)

	refreshRoute := h.api.Group("/auth")
	refreshRoute.Use(middlewares.AuthRefreshTokenMiddleware(secretKey))
	refreshRoute.POST("/refresh", h.RefreshToken)
}

func (h *UserHandler) Register(c *gin.Context) {
	var (
		req dto.RegisterRequest
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

	userID, err := h.service.Register(ctx, &req)
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

	c.JSON(http.StatusCreated, dto.RegisterResponse{
		ID: userID,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var (
		req dto.LoginRequest
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

	token, refreshToken, err := h.service.Login(ctx, &req)
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

	// ---------------------------
	//   SET COOKIE TOKEN
	// ---------------------------

	c.SetSameSite(http.SameSiteStrictMode)

	c.SetCookie(
		"token",
		token,
		3600, // expired 1 jam
		"/",
		"",    // domain
		false, // secure
		true,  // httpOnly
	)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		3600*24*7,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var (
		req dto.RefreshTokenRequest
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

	token, refreshToken, err := h.service.RefreshToken(ctx, &req, userID)
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

	// ---------------------------
	//   SET COOKIE TOKEN
	// ---------------------------

	c.SetSameSite(http.SameSiteStrictMode)

	c.SetCookie(
		"token",
		token,
		3600, // expired 1 jam
		"/",
		"",    // domain
		false, // secure
		true,  // httpOnly
	)

	c.SetCookie(
		"refresh_token",
		refreshToken,
		3600*24*7,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
	})
}
