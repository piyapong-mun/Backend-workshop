package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/piyapong-mun/simplebank/db/sqlc"
	"github.com/piyapong-mun/simplebank/token"
)

type createEntryRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Amount    int64  `json:"amount" binding:"required"`
}

func (server *Server) createEntry(ctx *gin.Context) {
	var req createEntryRequest

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

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	arg := db.CreateEntryParams{
		AccountID: accountID,
		Amount:    req.Amount,
	}

	entry, err := server.store.CreateEntry(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

type getEntryRequest struct {
	ID string `uri:"id" binding:"required"`
}

func (server *Server) getEntry(ctx *gin.Context) {
	var req getEntryRequest

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

	entryID, err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, responseError(err))
		return
	}

	entry, err := server.store.GetEntry(ctx, entryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	ctx.JSON(http.StatusOK, entry)
}

type listEntriesRequest struct {
	PageID   int64 `form:"page_id" binding:"required,min=1"`
	PageSize int64 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listEntries(ctx *gin.Context) {
	var req listEntriesRequest

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

	arg := db.ListEntriesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	entries, err := server.store.ListEntries(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, responseError(err))
		return
	}

	if entries == nil {
		entries = []db.Entry{}
	}

	ctx.JSON(http.StatusOK, entries)
}

