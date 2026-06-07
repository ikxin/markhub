package assets

import (
	"embed"
	"fmt"
)

//go:embed fallback/*.png
var fallbackFS embed.FS

func Fallback(name string) ([]byte, error) {
	data, err := fallbackFS.ReadFile("fallback/" + name + ".png")
	if err != nil {
		return nil, fmt.Errorf("read fallback %q: %w", name, err)
	}
	return data, nil
}
