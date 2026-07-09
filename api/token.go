package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	pq "github.com/lib/pq"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken    string `json:"access_token"`
	TokenExpiredAt string `json:"token_expired_at"`
}

// Login
func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	// Verify the refresh token
	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, responseError(err))
		return
	}

	// Check if the refresh token is in the database and valid
	session, err := server.store.GetSessions(ctx, refreshPayload.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if strings.Contains(pqErr.Message, "no rows in result set") {
				ctx.JSON(http.StatusNotFound, responseError(fmt.Errorf("session not found")))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	// Check if the session is blocked
	if session.IsBlocked {
		ctx.JSON(http.StatusUnauthorized, responseError(fmt.Errorf("blocked session")))
		return
	}

	// Check if the session belongs to the user who owns the refresh token
	if session.Username != refreshPayload.Username {
		ctx.JSON(http.StatusUnauthorized, responseError(fmt.Errorf("incorrect session user")))
		return
	}

	// Check if the session has expired
	if session.ExpiredAt.Before(time.Now()) {
		ctx.JSON(http.StatusUnauthorized, responseError(fmt.Errorf("expired session")))
		return
	}

	// Create a new access token for the user
	accessToken, accessPayload, err := server.tokenMaker.CreateToken(refreshPayload.Username, server.config.TokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	if accessPayload == nil {
		ctx.JSON(http.StatusInternalServerError, responseError(fmt.Errorf("failed to create token payload")))
		return
	}

	// Send the new access token and its expiration time in the response
	response := renewAccessTokenResponse{
		AccessToken:    accessToken,
		TokenExpiredAt: accessPayload.ExpiredAt.String(),
	}

	ctx.JSON(http.StatusOK, response)
}
