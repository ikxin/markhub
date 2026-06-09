package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/gin-gonic/gin"

	"markhub/internal/assets"
)

const cacheControl = "max-age=2592000"
const imageContentType = "image/webp"

const DefaultImageSize = 100
const maxImageSize = 2048

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

func ProxyImage(c *gin.Context, client HTTPClient, fetchURL string, fallback string, size int) {
	data, err := Fetch(c.Request.Context(), client, fetchURL)
	if err != nil {
		WriteFallback(c, fallback, size)
		return
	}

	if err := writeConvertedImage(c, data, size); err != nil {
		WriteFallback(c, fallback, size)
	}
}

func WriteImage(c *gin.Context, data []byte) {
	c.Header("Cache-Control", cacheControl)
	c.Data(http.StatusOK, imageContentType, data)
}

func WriteFallback(c *gin.Context, name string, size int) {
	data, err := assets.Fallback(name)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err := writeConvertedImage(c, data, size); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func ResizeToWebP(data []byte, size int) ([]byte, error) {
	return resizeToWebP(data, size, size)
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

func writeConvertedImage(c *gin.Context, data []byte, size int) error {
	converted, err := ResizeToWebP(data, size)
	if err != nil {
		return err
	}
	WriteImage(c, converted)
	return nil
}

func resizeToWebP(input []byte, width, height int) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid resize dimensions %dx%d", width, height)
	}

	ref, err := vips.NewImageFromBuffer(input)
	if err != nil {
		return nil, err
	}
	defer ref.Close()

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
	if err != nil {
		return nil, err
	}

	return out, nil
}
