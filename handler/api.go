package handler

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

// Hello handle api status
func Hello(c *fiber.Ctx) error {
	fmt.Println("Hello endpoint reached") // Debug log
	return c.SendString("Hello, from Redpay API update schema!")
}

func ReceiveCallback(c *fiber.Ctx) error {
	// Log headers
	log.Println("Headers:")
	for k, v := range c.GetReqHeaders() {
		log.Printf("%s: %s\n", k, v)
	}

	// Log query parameters
	log.Println("Query Parameters:")
	for k, v := range c.Queries() {
		log.Printf("%s: %s\n", k, v)
	}

	// Log request body
	body := c.Body()
	log.Println("Raw Request Body:", string(body))

	// Parse body ke map
	var requestData map[string]interface{}
	if err := c.BodyParser(&requestData); err != nil {
		log.Println("Error parsing body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Invalid request body"})
	}

	log.Println("Parsed Request Body:", requestData)

	return c.JSON(fiber.Map{"success": true, "message": "Callback received successfully"})
}
