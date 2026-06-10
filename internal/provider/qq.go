package provider

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

const qqSourceSize = 640

func QQ(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		number := c.Param("number")
		image.ProxyImage(c, client, fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=%d", number, qqSourceSize), "qq", image.OutputSize(c))
	}
}
