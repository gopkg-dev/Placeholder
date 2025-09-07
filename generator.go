package main

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"strconv"
	"unicode/utf8"

	"github.com/chai2010/webp"
	"github.com/fogleman/gg"
	"github.com/gopkg-dev/placeholder/cache"
)

//go:embed fonts/DouyinSansBold.otf
var embeddedFont []byte

const (
	// Maximum number of cached images in memory
	MAX_CACHE_ITEMS = 10000
	// Cache TTL in seconds (1 hour)
	CACHE_MAX_AGE = 3600
)

// ImageGenerator handles placeholder image generation with memory caching
type ImageGenerator struct {
	memoryCache *cache.LruCache // LRU cache for generated images
	fontPool    *FontPool       // Pool of pre-parsed fonts for reuse
}

// NewImageGenerator creates a new image generator with LRU cache and font pool
func NewImageGenerator() *ImageGenerator {
	return &ImageGenerator{
		memoryCache: cache.New(
			cache.WithSize(MAX_CACHE_ITEMS),
			cache.WithAge(CACHE_MAX_AGE),
			cache.WithUpdateAgeOnGet(),
		),
		fontPool: NewFontPool(FONT_POOL_SIZE),
	}
}

// GenerateImage creates or retrieves a cached placeholder image based on request parameters
func (ig *ImageGenerator) GenerateImage(req *ImageRequest) ([]byte, error) {
	cacheKey := ig.getCacheKey(req)

	// Check memory cache
	if data, found := ig.memoryCache.Get(cacheKey); found {
		if imageData, ok := data.([]byte); ok {
			return imageData, nil
		}
	}

	// Generate new image
	img, err := ig.createImage(req)
	if err != nil {
		return nil, err
	}

	data, err := ig.encodeImage(img, req.Type)
	if err != nil {
		return nil, err
	}

	// Store in memory cache
	ig.memoryCache.Set(cacheKey, data)
	return data, nil
}

// createImage generates a new placeholder image with text and colors
func (ig *ImageGenerator) createImage(req *ImageRequest) (image.Image, error) {
	dc := gg.NewContext(req.Width, req.Height)

	bgColor, err := parseHexColor(req.BgColor)
	if err != nil {
		return nil, err
	}
	dc.SetColor(bgColor)
	dc.Clear()

	fgColor, err := parseHexColor(req.FgColor)
	if err != nil {
		return nil, err
	}
	dc.SetColor(fgColor)

	fontSize := calculateOptimalFontSize(req.Width, req.Height, req.Text)

	fontFace := ig.fontPool.GetFont(fontSize)
	if fontFace != nil {
		dc.SetFontFace(*fontFace)
		defer ig.fontPool.PutFont(fontFace)
	} else {
		dc.LoadFontFace("", fontSize)
	}

	if !utf8.ValidString(req.Text) {
		req.Text = "Invalid UTF-8"
	}

	// Calculate center position with empirical adjustment for visual centering
	centerX := float64(req.Width) / 2

	// Use MeasureString to get text dimensions and calculate visual offset
	_, textHeight := dc.MeasureString(req.Text)

	// TODO: This 0.15 is an empirical guess based on typical font line spacing
	// A more precise approach would require accessing font metrics directly
	// For now, this value can be tuned based on visual testing results
	visualOffset := textHeight * 0.15
	centerY := float64(req.Height)/2 - visualOffset

	// Draw text with center anchors
	dc.DrawStringAnchored(req.Text, centerX, centerY, 0.5, 0.5)

	return dc.Image(), nil
}

// encodeImage converts image to specified format (PNG, JPG, GIF, WebP)
func (ig *ImageGenerator) encodeImage(img image.Image, imageType string) ([]byte, error) {
	var buf bytes.Buffer

	switch imageType {
	case "png":
		err := png.Encode(&buf, img)
		return buf.Bytes(), err
	case "jpg", "jpeg":
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
		return buf.Bytes(), err
	case "gif":
		err := gif.Encode(&buf, img, nil)
		return buf.Bytes(), err
	case "webp":
		err := webp.Encode(&buf, img, &webp.Options{Quality: 90})
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("unsupported image type: %s", imageType)
	}
}

// getCacheKey generates MD5 hash for cache key based on image parameters
func (ig *ImageGenerator) getCacheKey(req *ImageRequest) string {
	key := fmt.Sprintf("%dx%d_%s_%s_%s_%s", req.Width, req.Height, req.Type, req.BgColor, req.FgColor, req.Text)
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", hash)
}

// parseHexColor converts 6-character hex color string to RGBA color
func parseHexColor(hexColor string) (color.RGBA, error) {
	if len(hexColor) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color length: %s", hexColor)
	}

	r, err := strconv.ParseUint(hexColor[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	g, err := strconv.ParseUint(hexColor[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	b, err := strconv.ParseUint(hexColor[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}

	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}

// calculateOptimalFontSize computes appropriate font size based on image dimensions and text length
func calculateOptimalFontSize(width, height int, text string) float64 {
	minDim := float64(width)
	if height < width {
		minDim = float64(height)
	}

	runeCount := float64(utf8.RuneCountInString(text))
	if runeCount == 0 {
		runeCount = 1
	}

	// Enhanced algorithm with much more aggressive scaling for dramatic text sizes
	// Base font size calculation considering text length and available space
	targetWidthRatio := 0.85 // Use 85% of width for text (increased from 85%)
	avgCharWidth := (float64(width) * targetWidthRatio) / runeCount

	// More aggressive scale factor for bigger impact
	dimensionScale := (minDim / 200.0)                  // Normalize to 150px baseline (reduced from 200px for bigger fonts)
	baseFontSize := avgCharWidth * 0.8 * dimensionScale // Increased multiplier from 0.8 to 1.2

	// Much more aggressive dynamic limits for dramatic text size
	maxFontSize := minDim * 0.4  // Increased from 0.25 to 0.4 for much larger text
	minFontSize := minDim * 0.02 // Slightly increased minimum

	// Ensure reasonable bounds but allow much larger fonts
	if minFontSize < 12.0 {
		minFontSize = 12.0
	}
	if maxFontSize > 150.0 { // Much higher upper limit
		maxFontSize = 150.0
	}

	fontSize := baseFontSize
	if fontSize > maxFontSize {
		fontSize = maxFontSize
	}
	if fontSize < minFontSize {
		fontSize = minFontSize
	}

	return fontSize
}
