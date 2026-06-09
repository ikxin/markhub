package provider

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

func QQ(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		size := qqOutputSize(c)
		number := c.Param("number")
		image.ProxyImage(c, client, fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%s&s=%d", number, size), "qq", size)
	}
}

func qqOutputSize(c *gin.Context) int {
	if value, ok := c.GetQuery("s"); ok {
		return image.NormalizeImageSize(value)
	}
	if value, ok := c.GetQuery("spec"); ok {
		return image.NormalizeImageSize(value)
	}
	return image.DefaultImageSize
}
