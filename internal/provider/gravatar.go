package provider

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

const gravatarSourceSize = 640

func Gravatar(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		size := image.OutputSize(c)
		value, format := image.ResolveFormat(c, c.Param("hash"))
		hash, ok := gravatarHash(value)
		if !ok {
			image.WriteFallback(c, "gravatar", size, format)
			return
		}

		query := gravatarQuery(c)
		fetchURL := fmt.Sprintf("https://secure.gravatar.com/avatar/%s?%s", hash, query.Encode())
		image.ProxyImage(c, client, fetchURL, "gravatar", size, format)
	}
}

var hashPattern = regexp.MustCompile(`(?i)^([a-f0-9]{32}|[a-f0-9]{64})$`)

func gravatarHash(value string) (string, bool) {
	if hashPattern.MatchString(value) {
		return value, true
	}

	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", false
	}

	sum := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(value))))
	return hex.EncodeToString(sum[:]), true
}

func gravatarQuery(c *gin.Context) url.Values {
	values := url.Values{}
	values.Set("s", strconv.Itoa(gravatarSourceSize))
	setOptionalQuery(c, values, "default")
	setOptionalQuery(c, values, "f")
	setOptionalQuery(c, values, "forcedefault")
	setQueryDefault(c, values, "r", "g")
	setOptionalQuery(c, values, "rating")
	setOptionalQuery(c, values, "initials")
	setOptionalQuery(c, values, "name")
	return values
}

func setQueryDefault(c *gin.Context, values url.Values, key string, fallback string) {
	if value, ok := c.GetQuery(key); ok {
		values.Set(key, value)
		return
	}
	values.Set(key, fallback)
}

func setOptionalQuery(c *gin.Context, values url.Values, key string) {
	if value, ok := c.GetQuery(key); ok {
		values.Set(key, value)
	}
}
