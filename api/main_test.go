package api

import (
	"os"
	"testing"

	"time"

	"github.com/gin-gonic/gin"
	db "github.com/piyapong-mun/simplebank/db/sqlc"
	"github.com/piyapong-mun/simplebank/util"
)

func newTestServer(store db.Store) *Server {

	config, err := util.LoadConfig("../")
	if err != nil {
		return nil
	}

	config.TokenDuration = time.Minute
	config.SecretKey = util.RandomString(32)

	server, err := NewServer(config, store)
	if err != nil {
		return nil
	}
	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
