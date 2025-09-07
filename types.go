package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ImageRequest represents a request for generating a placeholder image
type ImageRequest struct {
	Width   int    // Image width in pixels
	Height  int    // Image height in pixels
	Type    string // Image format (png, jpg, gif, webp)
	BgColor string // Background color in hex format
	FgColor string // Foreground text color in hex format
	Text    string // Text to display on the image
}

// ImageSize represents image dimensions
type ImageSize struct {
	Width  int // Image width in pixels
	Height int // Image height in pixels
}

var (
	// Regular expressions for input validation
	sizeRegex  = regexp.MustCompile(`^(\d+)x(\d+)$`)   // Matches WxH format
	colorRegex = regexp.MustCompile(`^[a-fA-F0-9]{6}$`) // Matches 6-char hex color
)

const (
	MaxImageSize = 3000      // Maximum allowed image dimension
	DefaultBg    = "cccccc"  // Default background color (light gray)
	DefaultFg    = "666666"  // Default foreground color (dark gray)
	DefaultType  = "png"     // Default image format
)

// validTypes defines supported image formats
var validTypes = map[string]bool{
	"png":  true, // Portable Network Graphics
	"gif":  true, // Graphics Interchange Format
	"jpg":  true, // Joint Photographic Experts Group
	"jpeg": true, // Joint Photographic Experts Group (alt)
	"webp": true, // WebP format
}

// ParseSize parses size string in WxH format (e.g., "300x200") into ImageSize struct
func ParseSize(sizeStr string) (*ImageSize, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 3 {
		return nil, errors.New("invalid size format, expected WIDTHxHEIGHT")
	}

	width, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, errors.New("invalid width")
	}

	height, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, errors.New("invalid height")
	}

	if width <= 0 || height <= 0 || width > MaxImageSize || height > MaxImageSize {
		return nil, fmt.Errorf("size must be between 1x1 and %dx%d", MaxImageSize, MaxImageSize)
	}

	return &ImageSize{Width: width, Height: height}, nil
}

// ValidateType checks if the given image type is supported
func ValidateType(imageType string) bool {
	return validTypes[strings.ToLower(imageType)]
}

// ValidateColor validates hex color format (6 characters)
func ValidateColor(color string) bool {
	if color == "" {
		return true
	}
	return colorRegex.MatchString(color)
}

// applyDefault returns value if non-empty, otherwise returns defaultValue
func applyDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// NewImageRequest creates a new ImageRequest from parsed parameters with validation and defaults
func NewImageRequest(sizeStr, imageType string, p RequestQueryParams) (*ImageRequest, error) {
	// Parse and validate image dimensions
	size, err := ParseSize(sizeStr)
	if err != nil {
		return nil, err
	}

	// Apply defaults and validate image type
	if imageType == "" {
		imageType = DefaultType
	}
	if !ValidateType(imageType) {
		return nil, fmt.Errorf("unsupported image type: %s", imageType)
	}

	// Process colors with defaults and validation
	bg := applyDefault(p.Bg, DefaultBg)
	if !ValidateColor(bg) {
		return nil, fmt.Errorf("invalid background color: %s", bg)
	}

	fg := applyDefault(p.Fg, DefaultFg)  
	if !ValidateColor(fg) {
		return nil, fmt.Errorf("invalid foreground color: %s", fg)
	}

	// Set text with fallback to dimensions
	text := applyDefault(p.Text, fmt.Sprintf("%dx%d", size.Width, size.Height))

	return &ImageRequest{
		Width:   size.Width,
		Height:  size.Height,
		Type:    strings.ToLower(imageType),
		BgColor: bg,
		FgColor: fg,
		Text:    text,
	}, nil
}
