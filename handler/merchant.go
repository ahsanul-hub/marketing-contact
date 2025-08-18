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

	// Validate that at least one field is provided for update
	if input.ClientName == nil && input.AppName == nil && input.Mobile == nil &&
		input.ClientStatus == nil && input.Testing == nil && input.Lang == nil &&
		input.Phone == nil && input.Email == nil && input.CallbackURL == nil &&
		input.FailCallback == nil && input.Isdcb == nil && len(input.PaymentMethods) == 0 &&
		len(input.Settlements) == 0 && len(input.ClientApp) == 0 && len(input.ChannelRouteWeight) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one field must be provided for update",
		})
	}

	err := repository.UpdateMerchant(context.Background(), clientID, &input)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, "Client updated successfully")
}

func UpdateMerchantV2(c *fiber.Ctx) error {
	var input model.InputClientRequestV2
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

	// Validate that at least one field is provided for update
	if input.ClientName == nil && input.AppName == nil && input.Mobile == nil &&
		input.ClientStatus == nil && input.Testing == nil && input.Lang == nil &&
		input.Phone == nil && input.Email == nil && input.CallbackURL == nil &&
		input.FailCallback == nil && input.Isdcb == nil && len(input.SelectedPaymentMethods) == 0 &&
		len(input.ClientApp) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one field must be provided for update",
		})
	}

	err := repository.UpdateMerchantV2(context.Background(), clientID, &input)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, "Client updated successfully")
}

func GetMerchantByID(c *fiber.Ctx) error {
	clientID := c.Params("clientID")
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

func GetMerchantByAppID(c *fiber.Ctx) error {
	appID := c.Query("app_id")
	appKey := c.Query("app_key")

	if appID == "" || appKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "App ID and App Key are required",
		})
	}

	client, err := repository.FindClient(context.Background(), appKey, appID)
	if err != nil {
		return response.Response(c, fiber.StatusNotFound, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, client)
}

func GetAvailablePaymentMethods(c *fiber.Ctx) error {
	repo := repository.PaymentMethodRepository{DB: database.DB}
	paymentMethods, err := repo.GetAll()
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	// Transform to include only necessary fields for client selection
	var availableMethods []map[string]interface{}
	for _, pm := range paymentMethods {
		method := map[string]interface{}{
			"id":          pm.ID,
			"slug":        pm.Slug,
			"description": pm.Description,
			"type":        pm.Type,
			"route":       pm.Route,
			"flexible":    pm.Flexible,
			"status":      pm.Status,
			"min_denom":   pm.MinimumDenom,
			"denom":       pm.Denom,
			"prefix":      pm.Prefix,
		}
		availableMethods = append(availableMethods, method)
	}

	return response.ResponseSuccess(c, fiber.StatusOK, availableMethods)
}

func GetPaymentMethodRoutes(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Payment method slug is required",
		})
	}

	repo := repository.PaymentMethodRepository{DB: database.DB}
	paymentMethod, err := repo.GetBySlug(slug)
	if err != nil {
		return response.Response(c, fiber.StatusNotFound, "Payment method not found")
	}

	routes := map[string]interface{}{
		"slug":   paymentMethod.Slug,
		"routes": paymentMethod.Route,
		"type":   paymentMethod.Type,
	}

	return response.ResponseSuccess(c, fiber.StatusOK, routes)
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

func AddMerchantV2(c *fiber.Ctx) error {
	var requestData model.InputClientRequestV2
	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid input",
		})
	}

	err := repository.AddMerchantV2(context.Background(), &requestData)
	if err != nil {
		return response.Response(c, fiber.StatusInternalServerError, err.Error())
	}

	return response.ResponseSuccess(c, fiber.StatusOK, nil)
}
