package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	db "github.com/piyapong-mun/simplebank/db/sqlc"
	"github.com/piyapong-mun/simplebank/token"

	// sql
	pq "github.com/lib/pq"

	"github.com/google/uuid"
)

type createAccountRequest struct {
	// Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		// 🔒 ปรับให้เป็น Safe Type Assertion เพื่อป้องกันการ Panic
		if pqErr, ok := err.(*pq.Error); ok && pqErr != nil {
			// แนะนำให้เช็คจาก pqErr.Code.Name() จะแม่นยำกว่าการสแกน String ใน Message ครับ
			if pqErr.Code.Name() == "foreign_key_violation" || strings.Contains(pqErr.Message, "foreign key constraint") {
				ctx.JSON(http.StatusForbidden, responseError(fmt.Errorf("owner does not exist")))
				return
			}
			if pqErr.Code.Name() == "unique_violation" || strings.Contains(pqErr.Message, "unique constraint") {
				ctx.JSON(http.StatusForbidden, responseError(fmt.Errorf("account already exists")))
				return
			}
		}

		// หากไม่ใช่ pq.Error หรือไม่เข้าเงื่อนไขด้านบน จะหลุดมาที่ Internal Server Error อย่างปลอดภัย
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	accountID, err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	if account.Owner != authPayload.Username {
		ctx.JSON(http.StatusUnauthorized, responseError(fmt.Errorf("account does not belong to the authenticated user")))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountsRequest struct {
	PageID   int64 `form:"page_id" binding:"required,min=1"`
	PageSize int64 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccounts(ctx *gin.Context) {
	var req listAccountsRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	authPlayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.ListAccountsParams{
		Owner:  authPlayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	if accounts == nil {
		accounts = []db.Account{}
	}

	ctx.JSON(http.StatusOK, accounts)
}
