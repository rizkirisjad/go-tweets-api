package middlewares

import (
	"errors"
	"go-tweets/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		header = strings.TrimSpace(header)

		if header == "" {
			c.AbortWithError(http.StatusUnauthorized, errors.New("missing token"))
			return
		}

		userID, username, err := jwt.ValidateToken(header, secretKey, true)

		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, errors.New("missing token"))
			return
		}

		c.Set("userID", userID)
		c.Set("username", username)
		c.Next()
	}
}

func AuthRefreshTokenMiddleware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		header = strings.TrimSpace(header)

		if header == "" {
			c.AbortWithError(http.StatusUnauthorized, errors.New("missing token"))
			return
		}

		userID, username, err := jwt.ValidateToken(header, secretKey, false)

		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, errors.New("missing token"))
			return
		}

		c.Set("userID", userID)
		c.Set("username", username)
		c.Next()
	}
}
