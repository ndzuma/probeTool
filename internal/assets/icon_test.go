package assets

import (
	"testing"
)

func TestIconEmbedded(t *testing.T) {
	if len(Icon) == 0 {
		t.Error("Icon should be embedded and non-empty")
	}
}

func TestIconIsPNG(t *testing.T) {
	if len(Icon) < 8 {
		t.Fatal("Icon too small to be a valid PNG")
	}

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range pngHeader {
		if Icon[i] != b {
			t.Errorf("Icon does not have valid PNG header at byte %d: got %x, want %x", i, Icon[i], b)
		}
	}
}

func TestIconSize(t *testing.T) {
	minSize := 1000
	if len(Icon) < minSize {
		t.Errorf("Icon size %d bytes is suspiciously small, expected at least %d", len(Icon), minSize)
	}
}
