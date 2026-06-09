package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
	"markhub/internal/provider"
)

type HTTPClient = image.HTTPClient

type Options struct {
	Client HTTPClient
}

func NewRouter(options Options) *gin.Engine {
	client := options.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Recovery())

	router.GET("/github", provider.GitHubByID(client))
	router.GET("/github/:user", provider.GitHubByUser(client))
	router.GET("/gravatar/:hash", provider.Gravatar(client))
	router.GET("/qq/:number", provider.QQ(client))
	router.GET("/telegram/:user", provider.Telegram(client))
	router.GET("/opencollective/:user", provider.OpenCollective(client))
	router.GET("/favicon/:host", provider.Favicon(client))

	return router
}
