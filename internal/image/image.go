package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"

	"markhub/internal/assets"
)

const cacheControl = "max-age=2592000"

const DefaultImageSize = 100
const maxImageSize = 2048

type Format int

const (
	FormatWebP Format = iota
	FormatJPEG
	FormatPNG
	FormatAVIF
	FormatGIF
)

const defaultFormat = FormatWebP

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

func Fetch(ctx context.Context, client HTTPClient, fetchURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "markhub")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("upstream returned %d", response.StatusCode)
	}

	return io.ReadAll(response.Body)
}

func ProxyImage(c *gin.Context, client HTTPClient, fetchURL string, fallback string, size int, format Format) {
	data, err := Fetch(c.Request.Context(), client, fetchURL)
	if err != nil {
		WriteFallback(c, fallback, size, format)
		return
	}

	if err := writeConvertedImage(c, data, size, format); err != nil {
		WriteFallback(c, fallback, size, format)
	}
}

func WriteImage(c *gin.Context, data []byte, format Format) {
	c.Header("Cache-Control", cacheControl)
	c.Data(http.StatusOK, format.ContentType(), data)
}

func WriteFallback(c *gin.Context, name string, size int, format Format) {
	data, err := assets.Fallback(name)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err := writeConvertedImage(c, data, size, format); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func ResolveFormat(c *gin.Context, value string) (string, Format) {
	if trimmed, format, ok := TrimFormatSuffix(value); ok {
		return trimmed, format
	}
	return value, QueryFormat(c)
}

func TrimFormatSuffix(value string) (string, Format, bool) {
	ext := path.Ext(value)
	format, ok := ParseFormat(ext)
	if !ok {
		return value, defaultFormat, false
	}
	return strings.TrimSuffix(value, ext), format, true
}

func QueryFormat(c *gin.Context) Format {
	if value, ok := c.GetQuery("format"); ok {
		if format, ok := ParseFormat(value); ok {
			return format
		}
	}
	if value, ok := c.GetQuery("fmt"); ok {
		if format, ok := ParseFormat(value); ok {
			return format
		}
	}
	return defaultFormat
}

func ParseFormat(value string) (Format, bool) {
	value = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), ".")
	switch value {
	case "webp":
		return FormatWebP, true
	case "jpg", "jpeg":
		return FormatJPEG, true
	case "png":
		return FormatPNG, true
	case "avif":
		return FormatAVIF, true
	case "gif":
		return FormatGIF, true
	default:
		return defaultFormat, false
	}
}

func (format Format) ContentType() string {
	switch format {
	case FormatWebP:
		return "image/webp"
	case FormatJPEG:
		return "image/jpeg"
	case FormatPNG:
		return "image/png"
	case FormatAVIF:
		return "image/avif"
	case FormatGIF:
		return "image/gif"
	default:
		return "image/webp"
	}
}

func Resize(data []byte, size int, format Format) ([]byte, error) {
	return resize(data, size, size, format)
}

func OutputSize(c *gin.Context) int {
	if value, ok := c.GetQuery("size"); ok {
		return NormalizeImageSize(value)
	}
	if value, ok := c.GetQuery("s"); ok {
		return NormalizeImageSize(value)
	}
	return DefaultImageSize
}

func NormalizeImageSize(value string) int {
	size, err := strconv.Atoi(value)
	if err != nil || size <= 0 {
		return DefaultImageSize
	}
	if size > maxImageSize {
		return maxImageSize
	}
	return size
}

func writeConvertedImage(c *gin.Context, data []byte, size int, format Format) error {
	converted, err := Resize(data, size, format)
	if err != nil {
		return err
	}
	WriteImage(c, converted, format)
	return nil
}

func resize(input []byte, width, height int, format Format) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid resize dimensions %dx%d", width, height)
	}

	ref, err := vips.NewImageFromBuffer(input)
	if err != nil {
		return nil, err
	}
	defer ref.Close()

	return ResizeImageRef(ref, width, height, format)
}

func ResizeImageRef(ref *vips.ImageRef, width, height int, format Format) ([]byte, error) {
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

	return ExportImageRef(ref, format)
}

func ExportImageRef(ref *vips.ImageRef, format Format) ([]byte, error) {
	switch format {
	case FormatWebP:
		params := vips.NewWebpExportParams()
		params.StripMetadata = true
		params.Lossless = true
		out, _, err := ref.ExportWebp(params)
		return out, err
	case FormatJPEG:
		if ref.HasAlpha() {
			if err := ref.Flatten(&vips.Color{R: 255, G: 255, B: 255}); err != nil {
				return nil, err
			}
		}
		params := vips.NewJpegExportParams()
		params.StripMetadata = true
		out, _, err := ref.ExportJpeg(params)
		return out, err
	case FormatPNG:
		params := vips.NewPngExportParams()
		params.StripMetadata = true
		out, _, err := ref.ExportPng(params)
		return out, err
	case FormatAVIF:
		params := vips.NewAvifExportParams()
		params.StripMetadata = true
		out, _, err := ref.ExportAvif(params)
		return out, err
	case FormatGIF:
		params := vips.NewGifExportParams()
		params.StripMetadata = true
		out, _, err := ref.ExportGIF(params)
		return out, err
	default:
		return ExportImageRef(ref, defaultFormat)
	}
}
