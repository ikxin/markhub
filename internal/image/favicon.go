package image

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/davidbyttow/govips/v2/vips"
)

var (
	ErrInvalidICO     = errors.New("invalid ICO file")
	ErrUnsupportedICO = errors.New("unsupported ICO entry")
)

func ResizeToPNG(input []byte, width, height int) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid resize dimensions %dx%d", width, height)
	}

	imageData := input
	if isICO(input) {
		entry, err := SelectICOEntry(input, width)
		if err != nil {
			return nil, err
		}
		if !isEmbeddedImage(entry) {
			return nil, ErrUnsupportedICO
		}
		imageData = entry
	}

	ref, err := vips.NewImageFromBuffer(imageData)
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

func SelectICOEntry(input []byte, preferredSize int) ([]byte, error) {
	if len(input) < 6 ||
		binary.LittleEndian.Uint16(input[0:2]) != 0 ||
		binary.LittleEndian.Uint16(input[2:4]) != 1 {
		return nil, ErrInvalidICO
	}

	count := int(binary.LittleEndian.Uint16(input[4:6]))
	if count == 0 {
		return nil, errors.New("ICO contains no images")
	}
	if len(input) < 6+count*16 {
		return nil, ErrInvalidICO
	}

	bestIndex := 0
	bestScore := math.Inf(-1)
	for i := range count {
		entryOffset := 6 + i*16
		width := int(input[entryOffset])
		if width == 0 {
			width = 256
		}

		score := float64(width)
		if preferredSize > 0 {
			score = -math.Abs(float64(width-preferredSize))*0x10000 + float64(width)
		}

		if score > bestScore {
			bestScore = score
			bestIndex = i
		}
	}

	entryOffset := 6 + bestIndex*16
	imageSize := int(binary.LittleEndian.Uint32(input[entryOffset+8 : entryOffset+12]))
	imageOffset := int(binary.LittleEndian.Uint32(input[entryOffset+12 : entryOffset+16]))
	if imageSize < 0 || imageOffset < 0 || imageOffset > len(input) || imageSize > len(input)-imageOffset {
		return nil, ErrInvalidICO
	}

	return input[imageOffset : imageOffset+imageSize], nil
}

func isICO(input []byte) bool {
	return len(input) >= 6 &&
		binary.LittleEndian.Uint16(input[0:2]) == 0 &&
		binary.LittleEndian.Uint16(input[2:4]) == 1
}

func isEmbeddedImage(input []byte) bool {
	return isPNG(input) || isJPEG(input) || isGIF(input) || isWebP(input)
}

func isPNG(input []byte) bool {
	return len(input) >= 8 &&
		input[0] == 0x89 &&
		input[1] == 0x50 &&
		input[2] == 0x4e &&
		input[3] == 0x47 &&
		input[4] == 0x0d &&
		input[5] == 0x0a &&
		input[6] == 0x1a &&
		input[7] == 0x0a
}

func isJPEG(input []byte) bool {
	return len(input) >= 3 &&
		input[0] == 0xff &&
		input[1] == 0xd8 &&
		input[2] == 0xff
}

func isGIF(input []byte) bool {
	return len(input) >= 3 &&
		input[0] == 0x47 &&
		input[1] == 0x49 &&
		input[2] == 0x46
}

func isWebP(input []byte) bool {
	return len(input) >= 12 &&
		input[0] == 0x52 &&
		input[1] == 0x49 &&
		input[2] == 0x46 &&
		input[3] == 0x46 &&
		input[8] == 0x57 &&
		input[9] == 0x45 &&
		input[10] == 0x42 &&
		input[11] == 0x50
}
