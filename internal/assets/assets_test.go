package assets

import (
	"bytes"
	"image/png"
	"testing"
)

func TestFallback(t *testing.T) {
	data, err := Fallback("github")
	if err != nil {
		t.Fatalf("Fallback returned error: %v", err)
	}

	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("fallback is not a valid PNG: %v", err)
	}
	if config.Width != 100 || config.Height != 100 {
		t.Fatalf("fallback size = %dx%d, want 100x100", config.Width, config.Height)
	}
}
