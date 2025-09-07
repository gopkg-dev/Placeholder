package main

import (
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := fiber.New(fiber.Config{
		ServerHeader:          "placeholder-api",
		DisableStartupMessage: true,
		DisableKeepalive:      true,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return ctx.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(
		recover.New(),
		compress.New(),
		cors.New(cors.Config{
			AllowOrigins:  "*",
			AllowMethods:  "*",
			AllowHeaders:  "*",
			ExposeHeaders: "*",
			MaxAge:        86400,
		}),
		logger.New(logger.Config{
			TimeFormat: time.DateTime,
			Format:     "${time} | ${method} | ${url} | ${status} | ${latency} | ${ip} | ${error}\n",
		}),
	)

	// Rate limiting middleware - 50 requests per second per path
	app.Use(limiter.New(limiter.Config{
		Max:        50,
		Expiration: 1 * time.Second,
		KeyGenerator: func(ctx *fiber.Ctx) string {
			return ctx.Path()
		},
		LimitReached: func(ctx *fiber.Ctx) error {
			return ctx.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests",
			})
		},
	}))

	app.Get("/api/placeholder/:size", HandlerImage)

	// Display startup information and example URLs
	log.Println("🚀 Placeholder Image Generator started on http://localhost:8080")
	log.Println("")
	log.Println("📋 API Examples:")
	log.Println("   Basic usage:")
	log.Println("   • http://localhost:8080/api/placeholder/300x200")
	log.Println("   • http://localhost:8080/api/placeholder/400x300.png")
	log.Println("")
	log.Println("   Different formats:")
	log.Println("   • http://localhost:8080/api/placeholder/300x200.jpg")
	log.Println("   • http://localhost:8080/api/placeholder/250x150.gif")
	log.Println("   • http://localhost:8080/api/placeholder/500x300.webp")
	log.Println("")
	log.Println("   Custom colors:")
	log.Println("   • http://localhost:8080/api/placeholder/400x200?bg=ff0000&fg=ffffff")
	log.Println("   • http://localhost:8080/api/placeholder/300x300?bg=1e1e1e&fg=00ff00")
	log.Println("   • http://localhost:8080/api/placeholder/600x200?bg=0066cc&fg=ffffff")
	log.Println("")
	log.Println("   Custom text:")
	log.Println("   • http://localhost:8080/api/placeholder/400x250?text=Logo")
	log.Println("   • http://localhost:8080/api/placeholder/500x300?text=Hello%20World")
	log.Println("   • http://localhost:8080/api/placeholder/350x200.jpg?text=Sample&bg=navy&fg=white")
	log.Println("")
	log.Println("   Advanced examples:")
	log.Println("   • http://localhost:8080/api/placeholder/800x600.webp?bg=gradient&fg=gold&text=Premium")
	log.Println("   • http://localhost:8080/api/placeholder/1200x400.png?bg=f8f9fa&fg=343a40&text=Banner")
	log.Println("   • http://localhost:8080/api/placeholder/200x200.gif?bg=e91e63&fg=ffffff&text=Avatar")
	log.Println("")
	log.Println("⚡ Performance: <1ms cache hits, ~3ms cold generation")
	log.Println("🎯 Rate limit: 50 requests/second per endpoint")
	log.Println("💾 Cache: LRU memory cache with 1-hour TTL")
	log.Println("")

	err := app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
