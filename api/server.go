package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/piyapong-mun/simplebank/db/sqlc"
	"github.com/piyapong-mun/simplebank/token"
	"github.com/piyapong-mun/simplebank/util"
)

type Server struct {
	tokenMaker token.Maker
	config     util.Config
	store      db.Store
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {

	tokenMaker, err := token.NewPasetoMaker(config.SecretKey)
	if err != nil {
		return nil, err
	}

	server := &Server{tokenMaker: tokenMaker, config: config, store: store}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	server.router = router

	// Define Router

	// Users
	router.POST("/user/login", server.loginUser)
	router.POST("/users", server.createUser)

	// Tokens
	router.POST("/token/renewaccesstoken", server.renewAccessToken)

	routerGroup := router.Group("/").Use(authMiddleware(tokenMaker))

	// Accounts
	routerGroup.POST("/account", server.createAccount)
	routerGroup.GET("/account/:id", server.getAccount)
	routerGroup.GET("/accounts", server.listAccounts)

	// Entries
	routerGroup.POST("/entry", server.createEntry)
	routerGroup.GET("/entry/:id", server.getEntry)
	routerGroup.GET("/entries", server.listEntries)

	// Transfers
	routerGroup.POST("/transfer", server.createTransfer)
	routerGroup.GET("/transfer/:id", server.getTransfer)
	routerGroup.GET("/transfers", server.listTransfers)

	// TransferMoney (transactional)
	routerGroup.POST("/transfermoney", server.transferMoney)

	return server, nil
}

// Start the server
func (server *Server) Start(address string) {
	server.router.Run(address)
}

func responseError(err error) gin.H {
	return gin.H{"error": err.Error()}
}
