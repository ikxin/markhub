package provider

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

const githubSourceSize = 460

func GitHubByID(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Query("id")
		image.ProxyImage(c, client, fmt.Sprintf("https://avatars.githubusercontent.com/u/%s?size=%d", id, githubSourceSize), "github", image.OutputSize(c))
	}
}

func GitHubByUser(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.Param("user")
		image.ProxyImage(c, client, fmt.Sprintf("https://github.com/%s.png?size=%d", user, githubSourceSize), "github", image.OutputSize(c))
	}
}
