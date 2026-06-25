package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/piyapong-mun/simplebank/token"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, responseError(err))
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, responseError(err))
			return
		}

		if strings.ToLower(fields[0]) != authorizationTypeBearer {
			err := errors.New("unsupported authorization type")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, responseError(err))
			return
		}

		accessToken, err := tokenMaker.VerifyToken(fields[1])
		if err != nil {
			err := errors.New("invalid access token")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, responseError(err))
			return
		}

		ctx.Set(authorizationPayloadKey, accessToken)
		ctx.Next()
	}
}
