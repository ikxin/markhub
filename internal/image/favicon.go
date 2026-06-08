package image

import (
	"errors"
	"fmt"

	"github.com/davidbyttow/govips/v2/vips"
)

func ResizeToPNG(input []byte, width, height int) ([]byte, error) {
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

	params := vips.NewPngExportParams()
	params.StripMetadata = true
	out, _, err := ref.ExportPng(params)
	if err != nil {
		return nil, err
	}

	return out, nil
}
