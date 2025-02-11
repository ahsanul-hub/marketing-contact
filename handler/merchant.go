package handler

import (
	"app/database"
	"app/dto/model"
	"app/pkg/response"
	"app/repository"
	"context"

	"github.com/gofiber/fiber/v2"
)

type PaymentMethodHandler struct {
	Repo *repository.PaymentMethodRepository
}

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

func GetMerchantByID(c *fiber.Ctx) error {
	clientID := c.Params("id")
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Client App ID is required",
		})
	}

	client, err := repository.GetByClientID(clientID)
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

func NewPaymentMethodHandler(repo *repository.PaymentMethodRepository) *PaymentMethodHandler {
	return &PaymentMethodHandler{Repo: repo}
}

// CreatePaymentMethod untuk membuat payment method baru
func (h *PaymentMethodHandler) CreatePaymentMethod(c *fiber.Ctx) error {
	var paymentMethod model.PaymentMethod
	if err := c.BodyParser(&paymentMethod); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.Repo.Create(&paymentMethod); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(paymentMethod)
}

// GetPaymentMethods untuk mendapatkan semua payment method
func (h *PaymentMethodHandler) GetPaymentMethods(c *fiber.Ctx) error {
	paymentMethods, err := h.Repo.GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(paymentMethods)
}

// GetPaymentMethodByID untuk mendapatkan payment method berdasarkan ID
func (h *PaymentMethodHandler) GetPaymentMethodByID(c *fiber.Ctx) error {
	slug := c.Params("slug")                                    // Ambil slug dari parameter URL
	repo := repository.PaymentMethodRepository{DB: database.DB} // Inisialisasi repository

	paymentMethod, err := repo.GetBySlug(slug) // Panggil fungsi GetBySlug
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Payment method not found",
		})
	}

	return c.JSON(paymentMethod)
}

// UpdatePaymentMethod untuk memperbarui payment method
func (h *PaymentMethodHandler) UpdatePaymentMethod(c *fiber.Ctx) error {
	slug := c.Params("slug")
	paymentMethod, err := h.Repo.GetBySlug(slug)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment method not found"})
	}

	if err := c.BodyParser(paymentMethod); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.Repo.Update(paymentMethod); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(paymentMethod)
}

// DeletePaymentMethod untuk menghapus payment method
func (h *PaymentMethodHandler) DeletePaymentMethod(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if err := h.Repo.Delete(slug); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment method not found"})
	}

	return c.Status(fiber.StatusNoContent).JSON(nil)
}
