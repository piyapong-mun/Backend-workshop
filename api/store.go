package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/piyapong-mun/simplebank/db/sqlc"
	"github.com/piyapong-mun/simplebank/token"
)

type transferRequest struct {
	FromAccountID string `json:"from_account_id" binding:"required"`
	ToAccountID   string `json:"to_account_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) transferMoney(ctx *gin.Context) {
	var req transferRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	fromAccountID, err := uuid.Parse(req.FromAccountID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	toAccountID, err := uuid.Parse(req.ToAccountID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	fromAccount, valid := server.validAccount(ctx, fromAccountID, req.Currency)
	if !valid {
		return
	}

	authPlayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	if fromAccount.Owner != authPlayload.Username {
		err := fmt.Errorf("from account does not belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, responseError(err))
		return
	}

	_, valid = server.validAccount(ctx, toAccountID, req.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(ctx *gin.Context, accountID uuid.UUID, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, responseError(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return account, false
	}

	if account.Currency != currency {
		err := fmt.Errorf("account [%s] currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return account, false
	}

	return account, true
}
