package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	pq "github.com/lib/pq"
	db "github.com/piyapong-mun/simplebank/db/sqlc"
	"github.com/piyapong-mun/simplebank/token"
	"github.com/piyapong-mun/simplebank/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username         string `json:"username"`
	FullName         string `json:"full_name"`
	Email            string `json:"email"`
	PasswordChangeAt string `json:"password_change_at"`
	CreateAt         string `json:"create_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != "admin" {
		err := fmt.Errorf("only admin is allowed to perform this action")
		ctx.JSON(http.StatusForbidden, responseError(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if strings.Contains(pqErr.Message, "unique constraint") {
				ctx.JSON(http.StatusForbidden, responseError(fmt.Errorf("username or email already exists")))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	response := userResponse{
		Username:         user.Username,
		FullName:         user.FullName,
		Email:            user.Email,
		PasswordChangeAt: user.PasswordChangeAt.String(),
		CreateAt:         user.CreateAt.String(),
	}

	ctx.JSON(http.StatusCreated, response)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string `json:"access_token"`
	User        userResponse
}

// Login
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if strings.Contains(pqErr.Message, "no rows in result set") {
				ctx.JSON(http.StatusNotFound, responseError(fmt.Errorf("user not found")))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, responseError(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(user.Username, server.config.TokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	response := loginUserResponse{
		AccessToken: accessToken,
		User: userResponse{
			Username:         user.Username,
			FullName:         user.FullName,
			Email:            user.Email,
			PasswordChangeAt: user.PasswordChangeAt.String(),
			CreateAt:         user.CreateAt.String(),
		},
	}

	ctx.JSON(http.StatusOK, response)
}
