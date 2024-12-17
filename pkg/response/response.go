package response

import "github.com/gofiber/fiber/v2"

func Response(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": message,
	})
}

func ResponseSuccess(c *fiber.Ctx, status int, data interface{}) error {

	if data != nil {
		return c.Status(status).JSON(fiber.Map{
			"success": true,
			"data":    data,
		})
	}

	return c.Status(status).JSON(fiber.Map{
		"success": true,
	})
}
