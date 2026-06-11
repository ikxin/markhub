package provider

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

const githubSourceSize = 460

var githubIDPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

func GitHub(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.Param("identifier")
		if _, ok := c.GetQuery("id"); ok && githubIDPattern.MatchString(identifier) {
			image.ProxyImage(c, client, fmt.Sprintf("https://avatars.githubusercontent.com/u/%s?size=%d", identifier, githubSourceSize), "github", image.OutputSize(c))
			return
		}

		image.ProxyImage(c, client, fmt.Sprintf("https://github.com/%s.png?size=%d", identifier, githubSourceSize), "github", image.OutputSize(c))
	}
}
