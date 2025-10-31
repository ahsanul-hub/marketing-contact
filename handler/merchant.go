package handler

import (
	"app/database"
	"app/dto/model"
	"app/pkg/response"
	"app/repository"
	"context"
	"fmt"
	"log"
	"regexp"

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
		log.Println("Error parsing request body:", err)
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

// CRUD Payment Method Route Fees
func CreateRouteFee(c *fiber.Ctx) error {
	var in struct {
		PaymentMethodSlug string  `json:"payment_method_slug"`
		Route             string  `json:"route"`
		Fee               float64 `json:"fee"`
	}
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	fee := model.PaymentMethodRouteFee{PaymentMethodSlug: in.PaymentMethodSlug, Route: in.Route, Fee: in.Fee}
	if err := repository.CreateRouteFee(&fee); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fee)
}

func UpdateRouteFee(c *fiber.Ctx) error {
	var in struct {
		Fee *float64 `json:"fee"`
	}
	if err := c.BodyParser(&in); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	var id uint
	if _, err := fmt.Sscan(c.Params("id"), &id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid id"})
	}
	updates := map[string]interface{}{}
	if in.Fee != nil {
		updates["fee"] = *in.Fee
	}
	if err := repository.UpdateRouteFee(id, updates); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true})
}

func DeleteRouteFee(c *fiber.Ctx) error {
	var id uint
	if _, err := fmt.Sscan(c.Params("id"), &id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid id"})
	}
	if err := repository.DeleteRouteFee(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func ListRouteFees(c *fiber.Ctx) error {
	slug := c.Query("slug")
	fees, err := repository.ListRouteFees(slug)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fees)
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

// UpdateClientProfile mengizinkan client mengupdate data mereka sendiri (email, alamat, callback URL)
func UpdateClientProfile(c *fiber.Ctx) error {
	// Ambil token dari context yang sudah divalidasi oleh middleware
	token := c.Locals("user")
	if token == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized access",
		})
	}

	// Ambil header appkey dan appid untuk identifikasi client
	appKey := c.Get("appkey")
	appID := c.Get("appid")

	if appKey == "" || appID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Missing required headers: appkey and appid",
		})
	}

	// Parse request body
	var requestData struct {
		Email      *string                 `json:"email,omitempty"`
		Address    *string                 `json:"address,omitempty"`
		ClientApps []model.ClientAppUpdate `json:"client_apps,omitempty"`
	}

	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Validasi bahwa setidaknya ada satu field yang diupdate
	if requestData.Email == nil && requestData.Address == nil && len(requestData.ClientApps) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "At least one field must be provided for update",
		})
	}

	// Ambil client dari context yang sudah divalidasi oleh middleware ClientAuth
	clientInterface := c.Locals("client")
	if clientInterface == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Unauthorized access",
		})
	}

	client, ok := clientInterface.(*model.Client)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Invalid client data",
		})
	}

	// Update data client
	updateData := map[string]interface{}{}

	if requestData.Email != nil {
		updateData["email"] = *requestData.Email
	}

	if requestData.Address != nil {
		updateData["address"] = *requestData.Address
	}

	// Update data client jika ada perubahan
	if err := repository.UpdateClientProfile(context.Background(), client.UID, updateData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update client data: " + err.Error(),
		})
	}

	// Update callback URL untuk setiap app yang diupdate
	if len(requestData.ClientApps) > 0 {
		if err := repository.UpdateClientApps(context.Background(), client.UID, requestData.ClientApps); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Failed to update client apps: " + err.Error(),
			})
		}
	}

	// Refresh cache untuk client yang diupdate
	cacheKey := fmt.Sprintf("client:%s:%s", appKey, appID)
	repository.ClearClientCache(cacheKey)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Client profile updated successfully",
		"data": fiber.Map{
			"client_id":   client.ClientID,
			"client_name": client.ClientName,
			"email":       client.Email,
			"address":     client.Address,
			"client_apps": client.ClientApps,
		},
	})
}

func GetCreditCardLogByFirst6(c *fiber.Ctx) error {
	first6 := c.Params("first6")

	appKey := c.Get("appkey")
	appID := c.Get("appid")
	if appKey == "" || appID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Missing required headers: appkey and appid",
		})
	}
	_, err := repository.FindClient(c.Context(), appKey, appID)
	if err != nil {
		log.Println("")

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Error get client data",
		})

	}

	if len(first6) < 4 || len(first6) > 6 || !regexp.MustCompile(`^\d{4,6}$`).MatchString(first6) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "first6 param harus 4-6 digit angka",
		})
	}
	logs, err := repository.FindCreditCardLogsByFirst6(context.Background(), database.DB, first6)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(logs)
}
