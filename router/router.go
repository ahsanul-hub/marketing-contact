package router

import (
	"app/handler"
	"app/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())
	api.Get("/", handler.Hello)
	api.Post("/create", handler.CreateOrder)
	api.Post("/payment", handler.CreateTransactionV1)
	api.Post("/transaction", handler.CreateTransaction)
	api.Get("/transactions", handler.GetTransactions)
	api.Get("/transaction/:id", handler.GetTransactionByID)
	api.Get("/check/:id", handler.CheckTrans)
	api.Post("/test-payment", handler.TestPayment)
	api.Get("/order/:appid/:token", handler.PaymentPage)
	api.Get("/success-payment/:msisdn/:token", handler.SuccessPage)

	// Auth
	// auth := api.Group("/auth")

	// User

	user := api.Group("/user")
	user.Post("/login", handler.Login)
	user.Post("/register", handler.CreateUser)
	user.Patch("/:id", middleware.Protected(), handler.UpdateUser)
	user.Delete("/:id", middleware.Protected(), handler.DeleteUser)

	admin := api.Group("/admin", middleware.Protected())
	admin.Get("/users", handler.GetUser)
	admin.Delete("/user/:id", handler.DeleteUser)
	admin.Post("/merchant", middleware.AdminOnly(false), handler.AddMerchant)
	admin.Put("/merchant/:clientID", handler.UpdateMerchant)
	admin.Get("/merchants", handler.GetAllMerchants)
	admin.Get("/merchant/:clientID", handler.GetMerchantByAppID)
	admin.Delete("/merchant/:clientID", handler.DeleteMerchant)

}
