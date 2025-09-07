package main

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Global image generator instance
var generator *ImageGenerator

// Initialize the image generator with memory cache and font pool
func init() {
	generator = NewImageGenerator()
}

// RequestQueryParams represents HTTP request parameters for image generation
type RequestQueryParams struct {
	Size string `params:"size"` // Image size in WxH format
	Bg   string `query:"bg"`    // Background color (hex)
	Fg   string `query:"fg"`    // Foreground color (hex)
	Text string `query:"text"`  // Custom text to display
}

// HandlerImage processes placeholder image requests and returns generated images
func HandlerImage(c *fiber.Ctx) error {
	var params RequestQueryParams

	// Parse path parameters
	if err := c.ParamsParser(&params); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Parse query parameters
	if err := c.QueryParser(&params); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	parts := strings.Split(params.Size, ".")
	size := parts[0]
	imageType := ""

	if len(parts) == 2 {
		imageType = parts[1]
	}

	if params.Text != "" {
		if decodedText, err := url.QueryUnescape(params.Text); err == nil {
			params.Text = decodedText
		}
	}

	req, err := NewImageRequest(size, imageType, params)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	imageData, err := generator.GenerateImage(req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	contentType := getContentType(req.Type)
	c.Set("Content-Type", contentType)
	c.Set("Cache-Control", "public, max-age=31536000")

	return c.Send(imageData)
}

// getContentType returns appropriate MIME type for image format
func getContentType(imageType string) string {
	switch imageType {
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	default:
		return "image/png"
	}
}
