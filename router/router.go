package router

import (
	"app/handler"
	"app/middleware"
	"app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.elastic.co/apm/module/apmfiber"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {

	api := app.Group("/api", logger.New())

	paymentMethodRepo := repository.NewPaymentMethodRepository(db)
	paymentMethodHandler := handler.NewPaymentMethodHandler(paymentMethodRepo)

	api.Use(apmfiber.Middleware())
	api.Get("/", handler.Hello)
	api.Post("/create", handler.CreateOrder)
	api.Post("/payment", handler.CreateTransactionV1)
	api.Post("/transaction", handler.CreateTransaction)
	api.Get("/transactions", middleware.Protected(), middleware.AdminOnly(true), handler.GetTransactions)
	api.Get("/export", middleware.Protected(), middleware.AdminOnly(true), handler.ExportTransactions)
	api.Get("/transaction/:id", middleware.Protected(), middleware.AdminOnly(true), handler.GetTransactionByID)
	api.Post("/manual-callback/:id", middleware.Protected(), middleware.AdminOnly(true), handler.ManualCallback)
	api.Get("/check/:id", handler.CheckTrans)
	api.Post("/test-payment", handler.TestPayment)
	api.Post("/receive-callback1", handler.ReceiveCallback)
	api.Get("/order/:appid/:token", handler.PaymentPage)
	api.Get("/callback-triyakom", handler.CallbackTriyakom)
	api.Get("/success-payment/:msisdn/:token", handler.SuccessPage)
	api.Get("/success-payment/otp/:token", handler.SuccessPageOTP)
	api.Get("/input-otp/:ximpayid/:token", handler.InputOTPSF)
	api.Post("/mt-smartfren/:token", handler.MTSmartfren)

	merchant := api.Group("/merchant")
	merchant.Get("/transactions", handler.GetTransactionsMerchant)
	merchant.Get("/transaction/:id", handler.GetTransactionMerchantByID)

	user := api.Group("/user")
	user.Post("/login", handler.Login)
	user.Post("/register", handler.CreateUser)
	user.Patch("/:id", middleware.Protected(), handler.UpdateUser)
	user.Delete("/:id", middleware.Protected(), handler.DeleteUser)

	admin := api.Group("/admin", middleware.Protected())

	admin.Get("/users", handler.GetUser)
	admin.Delete("/user/:id", handler.DeleteUser)

	admin.Post("/payment-methods", middleware.AdminOnly(false), paymentMethodHandler.CreatePaymentMethod)
	admin.Get("/payment-methods", middleware.AdminOnly(false), paymentMethodHandler.GetPaymentMethods)
	admin.Get("/payment-methods/:slug", middleware.AdminOnly(false), paymentMethodHandler.GetPaymentMethodByID)
	admin.Put("/payment-methods/:slug", middleware.AdminOnly(false), paymentMethodHandler.UpdatePaymentMethod)
	admin.Delete("/payment-methods/:slug", middleware.AdminOnly(false), paymentMethodHandler.DeletePaymentMethod)

	admin.Post("/merchant", middleware.AdminOnly(false), handler.AddMerchant)
	admin.Put("/merchant/:clientID", middleware.AdminOnly(true), handler.UpdateMerchant)
	admin.Get("/merchants", middleware.AdminOnly(true), handler.GetAllMerchants)
	admin.Get("/merchant/:clientID", handler.GetMerchantByID)
	admin.Delete("/merchant/:clientID", middleware.AdminOnly(false), handler.DeleteMerchant)
}
