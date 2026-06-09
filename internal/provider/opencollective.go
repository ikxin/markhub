package provider

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

func OpenCollective(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := c.Param("user")
		image.ProxyImage(c, client, fmt.Sprintf("https://images.opencollective.com/%s/avatar.png?width=100&height=100", user), "opencollective", image.DefaultImageSize)
	}
}
