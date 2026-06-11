package provider

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

func Telegram(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, format := image.ResolveFormat(c, c.Param("user"))
		image.ProxyImage(c, client, fmt.Sprintf("https://t.me/i/userpic/320/%s.jpg", user), "telegram", image.OutputSize(c), format)
	}
}
