package provider

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

const openCollectiveSourceSize = 512

func OpenCollective(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, format := image.ResolveFormat(c, c.Param("user"))
		image.ProxyImage(c, client, fmt.Sprintf("https://images.opencollective.com/%s/avatar.png?width=%d&height=%d", user, openCollectiveSourceSize, openCollectiveSourceSize), "opencollective", image.OutputSize(c), format)
	}
}
