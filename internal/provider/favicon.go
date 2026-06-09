package provider

import (
	"context"
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

func Favicon(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		size := image.DefaultImageSize
		host, ok := normalizeHost(c.Param("host"))
		if !ok {
			image.WriteFallback(c, "favicon", size)
			return
		}

		if data, err := faviconByLinkTag(c.Request.Context(), client, host, size); err == nil {
			image.WriteImage(c, data)
			return
		}

		if data, err := faviconByDefaultPath(c.Request.Context(), client, host, size); err == nil {
			image.WriteImage(c, data)
			return
		}

		image.WriteFallback(c, "favicon", size)
	}
}

func faviconByLinkTag(ctx context.Context, client image.HTTPClient, host string, size int) ([]byte, error) {
	pageURL := httpURL(host, "/")
	page, err := image.Fetch(ctx, client, pageURL)
	if err != nil {
		return nil, err
	}

	href, ok := findIconHref(string(page))
	if !ok {
		return nil, errors.New("icon link not found")
	}

	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}
	iconURL, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	resolved := base.ResolveReference(iconURL)
	resolved.RawQuery = ""

	data, err := image.Fetch(ctx, client, resolved.String())
	if err != nil {
		return nil, err
	}
	return image.ResizeToWebP(data, size)
}

func faviconByDefaultPath(ctx context.Context, client image.HTTPClient, host string, size int) ([]byte, error) {
	data, err := image.Fetch(ctx, client, httpURL(host, "/favicon.ico"))
	if err != nil {
		return nil, err
	}
	return image.ResizeToWebP(data, size)
}

var linkTagPattern = regexp.MustCompile(`(?is)<link\b[^>]*\brel\s*=\s*["']?(icon|shortcut icon|alternate icon|apple-touch-icon)["']?[^>]*>`)
var hrefPattern = regexp.MustCompile(`(?is)\bhref\s*=\s*["']([^"']+)["']`)

func findIconHref(html string) (string, bool) {
	tag := linkTagPattern.FindString(html)
	if tag == "" {
		return "", false
	}

	matches := hrefPattern.FindStringSubmatch(tag)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

var hostnamePattern = regexp.MustCompile(`(?i)^([a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.?$`)

func normalizeHost(host string) (string, bool) {
	host = strings.TrimSpace(host)
	if host == "" || strings.ContainsAny(host, `/\`) || strings.ContainsRune(host, 0) {
		return "", false
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		ip := net.ParseIP(strings.Trim(host, "[]"))
		if ip == nil || ip.To4() != nil {
			return "", false
		}
		return host, true
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() == nil {
			return "[" + host + "]", true
		}
		return host, true
	}

	if len(host) > 253 || !hostnamePattern.MatchString(host) {
		return "", false
	}
	return host, true
}

func httpURL(host string, path string) string {
	return "http://" + host + path
}
