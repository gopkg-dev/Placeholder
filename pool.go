package main

import (
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	// Size of the font pool for reusing parsed fonts
	FONT_POOL_SIZE = 32
)

var (
	parsedFont   *opentype.Font // Cached parsed font to avoid repeated parsing
	fontParseErr error          // Error from font parsing, cached for reuse
	fontOnce     sync.Once      // Ensures font is parsed only once
)

// FontPool manages a pool of pre-parsed font faces for efficient reuse
type FontPool struct {
	fonts chan *font.Face // Channel-based pool of font faces
}

// NewFontPool creates a new font pool and pre-populates it with parsed fonts
func NewFontPool(size int) *FontPool {
	pool := &FontPool{
		fonts: make(chan *font.Face, size),
	}

	for range size {
		if face, err := parseEmbeddedFont(24.0, 72); err == nil {
			pool.fonts <- face
		}
	}
	return pool
}

// GetFont creates a font face with the specified size (pool disabled for size accuracy)
func (fp *FontPool) GetFont(fontSize float64) *font.Face {
	// Always create font with exact requested size instead of using pre-cached fonts
	// This ensures fontSize parameter is respected rather than returning cached 24px fonts
	if face, err := parseEmbeddedFont(fontSize, 72); err == nil {
		return face
	}
	return nil
}

// PutFont returns a font face to the pool for reuse
func (fp *FontPool) PutFont(face *font.Face) {
	if face == nil {
		return
	}
	select {
	case fp.fonts <- face:
	default:
	}
}


// parseEmbeddedFont parses embedded font data and creates font face with specified size and DPI
func parseEmbeddedFont(fontSize, dpi float64) (*font.Face, error) {
	fontOnce.Do(func() {
		parsedFont, fontParseErr = opentype.Parse(embeddedFont)
	})

	if fontParseErr != nil {
		return nil, fontParseErr
	}

	otfFace, err := opentype.NewFace(parsedFont, &opentype.FaceOptions{
		Size: fontSize,
		DPI:  dpi,
	})
	if err != nil {
		return nil, err
	}
	return &otfFace, nil
}
