package router

import (
	"app/handler"
	"app/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.elastic.co/apm/module/apmfiber"
)

// SetupRoutes setup router api
func SetupRoutes(app *fiber.App) {
	// Middleware
	api := app.Group("/api", logger.New())
	api.Use(apmfiber.Middleware())
	api.Get("/", handler.Hello)
	api.Post("/create", handler.CreateOrder)
	api.Post("/payment", handler.CreateTransactionV1)
	api.Post("/transaction", handler.CreateTransaction)
	api.Get("/transactions", middleware.AdminOnly(true), handler.GetTransactions)
	api.Get("/transaction/:id", middleware.AdminOnly(true), handler.GetTransactionByID)
	api.Get("/check/:id", handler.CheckTrans)
	api.Post("/test-payment", handler.TestPayment)
	api.Get("/order/:appid/:token", handler.PaymentPage)
	api.Get("/success-payment/:msisdn/:token", handler.SuccessPage)

	merchant := api.Group("/merchant")
	merchant.Get("/transactions", handler.GetTransactionsMerchant)
	merchant.Get("/transaction/:id", middleware.AdminOnly(true), handler.GetTransactionMerchantByID)

	user := api.Group("/user")
	user.Post("/login", handler.Login)
	user.Post("/register", handler.CreateUser)
	user.Patch("/:id", middleware.Protected(), handler.UpdateUser)
	user.Delete("/:id", middleware.Protected(), handler.DeleteUser)

	admin := api.Group("/admin", middleware.Protected())
	admin.Get("/users", handler.GetUser)
	admin.Delete("/user/:id", handler.DeleteUser)
	admin.Post("/merchant", middleware.AdminOnly(false), handler.AddMerchant)
	admin.Put("/merchant/:clientID", middleware.AdminOnly(true), handler.UpdateMerchant)
	admin.Get("/merchants", middleware.AdminOnly(true), handler.GetAllMerchants)
	admin.Get("/merchant/:clientID", handler.GetMerchantByAppID)
	admin.Delete("/merchant/:clientID", middleware.AdminOnly(false), handler.DeleteMerchant)

}
