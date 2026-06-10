package provider

import (
	"context"
	"errors"
	"fmt"
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
	return resizeFaviconToWebP(data, size)
}

func faviconByDefaultPath(ctx context.Context, client image.HTTPClient, host string, size int) ([]byte, error) {
	data, err := image.Fetch(ctx, client, httpURL(host, "/favicon.ico"))
	if err != nil {
		return nil, err
	}
	return resizeFaviconToWebP(data, size)
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

func resizeFaviconToWebP(input []byte, size int) ([]byte, error) {
	if !isICO(input) {
		return image.ResizeToWebP(input, size)
	}

	ref, err := loadLargestICOPage(input)
	if err != nil {
		return nil, err
	}
	defer ref.Close()

	return resizeImageRefToWebP(ref, size, size)
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

func resizeImageRefToWebP(ref *vips.ImageRef, width, height int) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid resize dimensions %dx%d", width, height)
	}
	if ref.Width() <= 0 || ref.Height() <= 0 {
		return nil, errors.New("image has invalid dimensions")
	}

	hScale := float64(width) / float64(ref.Width())
	vScale := float64(height) / float64(ref.Height())
	if err := ref.ResizeWithVScale(hScale, vScale, vips.KernelLanczos3); err != nil {
		return nil, err
	}

	params := vips.NewWebpExportParams()
	params.StripMetadata = true
	params.Lossless = true
	out, _, err := ref.ExportWebp(params)
	return out, err
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
