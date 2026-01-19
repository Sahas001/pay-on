package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const authUserIDKey = "auth_user_id"

func (server *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
			c.Abort()
			return
		}

		tokenStr := parts[1]
		token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(server.config.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*jwt.RegisteredClaims)
		if !ok || claims.Subject == "" {
			c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
			c.Abort()
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
			c.Abort()
			return
		}

		c.Set(authUserIDKey, userID)
		c.Next()
	}
}

func authUserID(c *gin.Context) (uuid.UUID, bool) {
	value, ok := c.Get(authUserIDKey)
	if !ok {
		return uuid.UUID{}, false
	}
	id, ok := value.(uuid.UUID)
	return id, ok
}
