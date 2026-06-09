package image

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"golang.org/x/image/webp"
)

func TestMain(m *testing.M) {
	vips.LoggingSettings(nil, vips.LogLevelWarning)
	vips.Startup(nil)
	code := m.Run()
	vips.Shutdown()
	os.Exit(code)
}

func TestResizeToWebPWithEmbeddedPNGICO(t *testing.T) {
	input := readFixture(t, "toutiao.ico")

	output, err := ResizeToWebP(input, 100, 100)
	if err != nil {
		t.Fatalf("ResizeToWebP returned error: %v", err)
	}

	assertWebPSize(t, output, 100, 100)
}

func TestResizeToWebPWithDirectPNG(t *testing.T) {
	input := readFixture(t, "tiktok.ico")

	output, err := ResizeToWebP(input, 120, 80)
	if err != nil {
		t.Fatalf("ResizeToWebP returned error: %v", err)
	}

	assertWebPSize(t, output, 120, 80)
}

func TestResizeToWebPWithDIBICO(t *testing.T) {
	input := readFixture(t, "github.ico")

	output, err := ResizeToWebP(input, 100, 100)
	if err != nil {
		t.Fatalf("ResizeToWebP returned error: %v", err)
	}

	assertWebPSize(t, output, 100, 100)
}

func assertWebPSize(t *testing.T, data []byte, width, height int) {
	t.Helper()

	config, err := webp.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("data is not a valid WebP: %v", err)
	}
	if config.Width != width || config.Height != height {
		t.Fatalf("WebP size = %dx%d, want %dx%d", config.Width, config.Height, width, height)
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("..", "..", "test", "fixtures", "favicon", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}
