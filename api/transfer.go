package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/piyapong-mun/simplebank/db/sqlc"
	"github.com/piyapong-mun/simplebank/token"
)

type createTransferRequest struct {
	FromAccountID string `json:"from_account_id" binding:"required"`
	ToAccountID   string `json:"to_account_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest

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

	arg := db.CreateTransfersParams{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        req.Amount,
	}

	transfer, err := server.store.CreateTransfers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	ctx.JSON(http.StatusOK, transfer)
}

type getTransferRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (server *Server) getTransfer(ctx *gin.Context) {
	var req getTransferRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != "admin" {
		err := fmt.Errorf("only admin is allowed to perform this action")
		ctx.JSON(http.StatusForbidden, responseError(err))
		return
	}

	transferID, err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	transfer, err := server.store.GetTransfers(ctx, transferID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	ctx.JSON(http.StatusOK, transfer)
}

type listTransfersRequest struct {
	PageID   int64 `form:"page_id" binding:"required,min=1"`
	PageSize int64 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listTransfers(ctx *gin.Context) {
	var req listTransfersRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if authPayload.Username != "admin" {
		err := fmt.Errorf("only admin is allowed to perform this action")
		ctx.JSON(http.StatusForbidden, responseError(err))
		return
	}

	arg := db.ListTransfersParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	transfers, err := server.store.ListTransfers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	if transfers == nil {
		transfers = []db.Transfer{}
	}

	ctx.JSON(http.StatusOK, transfers)
}

