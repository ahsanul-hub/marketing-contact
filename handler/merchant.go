package handler

import (
	"app/dto/model"
	"app/pkg/response"
	"app/repository"
	"context"

	"github.com/gofiber/fiber/v2"
)

func AddMerchant(c *fiber.Ctx) error {

	var requestData model.InputClientRequest
	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	err := repository.AddMerchant(context.Background(), &requestData)

	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, nil)
}

func UpdateMerchant(c *fiber.Ctx) error {

	var input model.InputClientRequest
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	clientID := c.Params("clientID")
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Client ID is required",
		})
	}

	err := repository.UpdateMerchant(context.Background(), clientID, &input)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, "Client updated successfully")
}

func GetMerchantByAppID(c *fiber.Ctx) error {
	clientAppID := c.Params("clientID")
	if clientAppID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Client App ID is required",
		})
	}

	client, err := repository.GetByClientAppID(clientAppID)
	if err != nil {
		return response.Response(c, fiber.StatusNotFound, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, client)
}

func GetAllMerchants(c *fiber.Ctx) error {
	clients, err := repository.GetAllClients()
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, clients)
}

func DeleteMerchant(c *fiber.Ctx) error {

	clientAppID := c.Params("clientID")
	if clientAppID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Client ID is required",
		})
	}

	err := repository.DeleteMerchant(clientAppID)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, "Client deleted successfully")
}
