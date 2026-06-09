package provider

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

func GitHubByID(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")
		image.ProxyImage(c, client, fmt.Sprintf("https://avatars.githubusercontent.com/u/%s?size=100", id), "github", image.DefaultImageSize)
	}
}

func GitHubByUser(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.Param("user")
		image.ProxyImage(c, client, fmt.Sprintf("https://github.com/%s.png?size=100", user), "github", image.DefaultImageSize)
	}
}
