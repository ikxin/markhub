package provider

import (
	"context"
	"errors"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"

	"markhub/internal/image"
)

const maxICOPageScan = 256

func Favicon(client image.HTTPClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		size := image.OutputSize(c)
		host, format := image.ResolveFormat(c, c.Param("host"))
		host, ok := normalizeHost(host)
		if !ok {
			image.WriteFallback(c, "favicon", size, format)
			return
		}

		if data, err := faviconByLinkTag(c.Request.Context(), client, host, size, format); err == nil {
			image.WriteImage(c, data, format)
			return
		}

		if data, err := faviconByDefaultPath(c.Request.Context(), client, host, size, format); err == nil {
			image.WriteImage(c, data, format)
			return
		}

		image.WriteFallback(c, "favicon", size, format)
	}
}

func faviconByLinkTag(ctx context.Context, client image.HTTPClient, host string, size int, format image.Format) ([]byte, error) {
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
	return resizeFavicon(data, size, format)
}

func faviconByDefaultPath(ctx context.Context, client image.HTTPClient, host string, size int, format image.Format) ([]byte, error) {
	data, err := image.Fetch(ctx, client, httpURL(host, "/favicon.ico"))
	if err != nil {
		return nil, err
	}
	return resizeFavicon(data, size, format)
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

func resizeFavicon(input []byte, size int, format image.Format) ([]byte, error) {
	if !isICO(input) {
		return image.Resize(input, size, format)
	}

	ref, err := loadLargestICOPage(input)
	if err != nil {
		return nil, err
	}
	defer ref.Close()

	return image.ResizeImageRef(ref, size, size, format)
}

func loadLargestICOPage(input []byte) (*vips.ImageRef, error) {
	var bestRef *vips.ImageRef
	bestPixels := 0

	for page := 0; page < maxICOPageScan; page++ {
		params := vips.NewImportParams()
		params.Page.Set(page)

		ref, err := vips.LoadImageFromBuffer(input, params)
		if err != nil {
			if page == 0 {
				return nil, err
			}
			break
		}

		pixels := ref.Width() * ref.Height()
		if pixels > bestPixels {
			if bestRef != nil {
				bestRef.Close()
			}
			bestRef = ref
			bestPixels = pixels
			continue
		}

		ref.Close()
	}

	if bestRef == nil {
		return nil, errors.New("ico has invalid dimensions")
	}
	return bestRef, nil
}

func isICO(input []byte) bool {
	return len(input) >= 4 && input[0] == 0 && input[1] == 0 && input[2] == 1 && input[3] == 0
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
