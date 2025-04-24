package handler

import (
	"app/lib"
	"app/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func CheckStatusDana(c *fiber.Ctx) error {
	id := c.Params("id")

	res, err := lib.CheckOrderDana(id)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}
