package api

import (
	"net/http"
	"time"

	database "github.com/Sahas001/pay-on/internal/database/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	PhoneNumber string  `json:"phone_number" binding:"required"`
	Email       *string `json:"email"`
	Password    string  `json:"password" binding:"required"`
}

type authResponse struct {
	AccessToken string  `json:"access_token"`
	UserID      string  `json:"user_id"`
	PhoneNumber string  `json:"phone_number"`
	Email       *string `json:"email"`
}

func (server *Server) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	user, err := server.store.CreateUser(c.Request.Context(), database.CreateUserParams{
		PhoneNumber:  req.PhoneNumber,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	token, err := server.newAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		AccessToken: token,
		UserID:      user.ID.String(),
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	})
}

type loginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Password    string `json:"password" binding:"required"`
}

func (server *Server) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	var user database.User
	var err error

	switch {
	case req.PhoneNumber != "":
		user, err = server.store.GetUserByPhone(c.Request.Context(), req.PhoneNumber)
	case req.Email != "":
		user, err = server.store.GetUserByEmail(c.Request.Context(), &req.Email)
	default:
		c.JSON(http.StatusBadRequest, errorResponse(errMissingQuery))
		return
	}
	if err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, errorResponse(errInvalidCredentials))
		return
	}

	token, err := server.newAccessToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, authResponse{
		AccessToken: token,
		UserID:      user.ID.String(),
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	})
}

func (server *Server) newAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(server.config.AccessTokenDuration)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(server.config.JWTSecret))
}
