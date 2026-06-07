package image

import (
	"bytes"
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestMain(m *testing.M) {
	vips.LoggingSettings(nil, vips.LogLevelWarning)
	vips.Startup(nil)
	code := m.Run()
	vips.Shutdown()
	os.Exit(code)
}

func TestResizeToPNGWithEmbeddedPNGICO(t *testing.T) {
	input := readFixture(t, "toutiao.ico")

	output, err := ResizeToPNG(input, 100, 100)
	if err != nil {
		t.Fatalf("ResizeToPNG returned error: %v", err)
	}

	config, err := png.DecodeConfig(bytes.NewReader(output))
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}
	if config.Width != 100 || config.Height != 100 {
		t.Fatalf("output size = %dx%d, want 100x100", config.Width, config.Height)
	}
}

func TestResizeToPNGWithDirectPNG(t *testing.T) {
	input := readFixture(t, "tiktok.ico")

	output, err := ResizeToPNG(input, 100, 100)
	if err != nil {
		t.Fatalf("ResizeToPNG returned error: %v", err)
	}

	config, err := png.DecodeConfig(bytes.NewReader(output))
	if err != nil {
		t.Fatalf("output is not a valid PNG: %v", err)
	}
	if config.Width != 100 || config.Height != 100 {
		t.Fatalf("output size = %dx%d, want 100x100", config.Width, config.Height)
	}
}

func TestResizeToPNGWithDIBICO(t *testing.T) {
	input := readFixture(t, "github.ico")

	_, err := ResizeToPNG(input, 100, 100)
	if !errors.Is(err, ErrUnsupportedICO) {
		t.Fatalf("ResizeToPNG error = %v, want ErrUnsupportedICO", err)
	}
}

func TestSelectICOEntryRespectsPreferredSize(t *testing.T) {
	input := readFixture(t, "github.ico")

	entry16, err := SelectICOEntry(input, 16)
	if err != nil {
		t.Fatalf("SelectICOEntry(16) returned error: %v", err)
	}
	entry32, err := SelectICOEntry(input, 32)
	if err != nil {
		t.Fatalf("SelectICOEntry(32) returned error: %v", err)
	}

	if bytes.Equal(entry16, entry32) {
		t.Fatal("SelectICOEntry returned the same entry for 16px and 32px preferences")
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("..", "..", "test", "fixtures", "ico-to-sharp", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}
