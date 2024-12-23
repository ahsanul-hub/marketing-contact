package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Hello handle api status
func Hello(c *fiber.Ctx) error {
	fmt.Println("Hello endpoint reached") // Debug log
	return c.SendString("Hello, World updated!")
}
