package router

import (
	"app/config"
	"app/handler"
	"app/middleware"
	"app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.elastic.co/apm/module/apmfiber"
	"gorm.io/gorm"
)

func SetupRoutes(app *fiber.App, db *gorm.DB) {
	// Serve static assets (e.g., images) from /assets using absolute path
	dirAsset := config.Config("DIR_ASSETS", "/home/aldi/dcb-backend/assets")
	app.Static("/assets", dirAsset)

	api := app.Group("/api", logger.New())

	paymentMethodRepo := repository.NewPaymentMethodRepository(db)
	paymentMethodHandler := handler.NewPaymentMethodHandler(paymentMethodRepo)

	api.Use(apmfiber.Middleware())
	api.Get("/", handler.Hello)
	api.Post("/create", handler.CreateOrder)
	api.Post("/v2/create", handler.CreateOrderLegacy)
	api.Post("/payment", handler.CreateTransactionV1)
	api.Post("/payment-va", handler.CreateTransactionVa)
	api.Post("/payment/notelco", handler.CreateTransactionNonTelco)
	api.Post("/transaction", handler.CreateTransaction)
	api.Get("/transactions", middleware.Protected(), middleware.AdminOnly(false), handler.GetTransactions)
	api.Get("/export", middleware.Protected(), middleware.AdminOnly(false), handler.ExportTransactions)
	api.Get("/export/transactions-merchant", handler.ExportTransactionsMerchant)
	api.Get("/transaction/:id", middleware.Protected(), middleware.AdminOnly(false), handler.GetTransactionByID)
	api.Post("/manual-callback/:id", middleware.Protected(), middleware.AdminOnly(false), handler.ManualCallback)
	api.Post("/merchant/manual-callback/:id", handler.ManualCallbackClient)
	// api.Get("/mark-paid/:id", handler.MakePaid)
	// api.Get("/mark-failed/:id", handler.MakeFailed)
	api.Get("/check/:id", handler.CheckTrans)
	api.Post("/receive-callback1", handler.ReceiveCallback)
	api.Get("/order/:appid/:token", handler.PaymentPage)
	api.Get("/v2/order/:appid/:token", handler.CreateTransactionLegacy)
	api.Get("/v1/order/:appid/:token", handler.PaymentPageLegacy)
	api.Get("/payment-qris", handler.PaymentQrisRedirect)
	api.Get("/payment-qris/:id", handler.QrisPage)
	api.Get("/callback-triyakom", handler.CallbackTriyakom)
	api.Get("/callback/midtrans", handler.MidtransCallback)
	api.Post("/callback/harsya", handler.CallbackHarsya)
	api.Post("/callback/midtrans", handler.MidtransCallback)
	api.Post("/notify/xl", handler.XLCallback)
	api.Post("/callback/digiph", handler.DigiphCallback)
	api.Get("/mo/telkomsel", handler.MoTelkomsel)
	api.Get("/return/dana", handler.PayReturnSuccess)
	api.Get("/check-status/dana/:id", handler.CheckStatusDana)
	api.Get("/check-status/dana-faspay/:id", handler.CheckStatusDanaFaspay)
	api.Get("/check-status/ovo/:id", handler.CheckStatusOvo)
	api.Get("/check-status/qris-harsya/:id", handler.CheckStatusQrisHarsya)
	api.Get("/checkstatus/:id", handler.CheckTransactionStatus)
	api.Post("/v1/checkstatus", handler.CheckTransactionStatusLegacy)

	api.Get("/summary/transaction", middleware.Protected(), handler.GetTransactionSummary)
	api.Get("/report/merchant", middleware.Protected(), middleware.AdminOnly(false), handler.GetReport)
	api.Get("/report/merchant/margin", middleware.Protected(), middleware.AdminOnly(false), handler.GetReportMargin)
	api.Get("/test-email", middleware.Protected(), middleware.AdminOnly(true), handler.TestEmailService)
	api.Get("/test-sftp", middleware.Protected(), middleware.AdminOnly(true), handler.TestSFTPConnection)

	// Traffic Monitoring endpoints
	api.Get("/traffic/monitoring", middleware.Protected(), middleware.AdminOnly(false), handler.GetTrafficMonitoringChart)
	api.Get("/traffic/summary", middleware.Protected(), middleware.AdminOnly(false), handler.GetTrafficSummary)

	// app.Get("/cached-transactions", handler.GetAllCachedTransactions)

	api.Post("/callback/dana", handler.DanaCallback)
	api.Post("/callback/faspay", handler.DanaFaspayCallback)
	api.Get("/success-payment/:msisdn/:token", handler.SuccessPage)
	api.Get("/v1/success-payment/:msisdn/:token", handler.SuccessPageLegacy)
	api.Get("/success-otp/:token", handler.SuccessPageOTP)
	api.Get("/va-payment/:va", handler.VaPage)
	api.Get("/input-otp/:ximpayid/:token", handler.InputOTPSF)
	api.Post("/mt-smartfren/:token", handler.MTSmartfren)
	api.Post("/smartfren/otp", handler.MTSmartfren)
	api.Post("/block-mdn", middleware.Protected(), middleware.AdminOnly(false), handler.BlockMDNHandler)
	api.Post("/block-userId", middleware.Protected(), middleware.AdminOnly(false), handler.BlockUserIdHandler)
	api.Post("/unblock-mdn", middleware.Protected(), middleware.AdminOnly(false), handler.UnblockMDNHandler)
	api.Post("/unblock-userId", middleware.Protected(), middleware.AdminOnly(false), handler.UnblockUserIDHandler)
	api.Post("/bca/inquiry", handler.InquiryBca)
	api.Post("/bca/payment", handler.PaymentBca)
	api.Post("/bca/token", handler.TokenBca)

	api.Get("/credit-card-bin/:first4", handler.GetCreditCardLogByFirst4)

	merchant := api.Group("/merchant")
	merchant.Get("/transactions", handler.GetTransactionsMerchant)
	merchant.Get("/transaction/:id", handler.GetTransactionMerchantByID)
	merchant.Get("/detail", middleware.Protected(), handler.GetMerchantByAppID)
	merchant.Put("/profile", middleware.Protected(), middleware.ClientAuth(), handler.UpdateClientProfile)

	user := api.Group("/user")
	user.Post("/login", handler.Login)
	user.Post("/register", handler.CreateUser)
	user.Patch("/:id", middleware.Protected(), handler.UpdateUser)
	user.Delete("/:id", middleware.Protected(), handler.DeleteUser)

	admin := api.Group("/admin", middleware.Protected())

	admin.Get("/users", handler.GetUser)
	admin.Delete("/user/:id", handler.DeleteUser)

	admin.Post("/payment-methods", middleware.AdminOnly(true), paymentMethodHandler.CreatePaymentMethod)
	// admin.Get("/payment-methods", middleware.AdminOnly(false), paymentMethodHandler.GetPaymentMethods)
	admin.Get("/payment-methods/:slug", middleware.AdminOnly(false), paymentMethodHandler.GetPaymentMethodByID)
	admin.Put("/payment-methods/:slug", middleware.AdminOnly(true), paymentMethodHandler.UpdatePaymentMethod)
	admin.Delete("/payment-methods/:slug", middleware.AdminOnly(true), paymentMethodHandler.DeletePaymentMethod)

	admin.Get("/payment-methods", middleware.AdminOnly(false), handler.GetAvailablePaymentMethods)
	admin.Get("/payment-method-routes/:slug", middleware.AdminOnly(false), handler.GetPaymentMethodRoutes)

	// Route fee CRUD
	admin.Post("/route-fees", middleware.AdminOnly(true), handler.CreateRouteFee)
	admin.Put("/route-fees/:id", middleware.AdminOnly(true), handler.UpdateRouteFee)
	admin.Get("/route-fees", middleware.AdminOnly(false), handler.ListRouteFees)
	admin.Delete("/route-fees/:id", middleware.AdminOnly(true), handler.DeleteRouteFee)

	// admin.Post("/bodysign/generate", middleware.AdminOnly(false), handler.GenerateBodySign)
	// admin.Post("/bodysign/validate", middleware.AdminOnly(false), handler.ValidateBodySign)

	// admin.Post("/admin/channel-route", middleware.AdminOnly(false), handler.AddChannelRouteWeight)

	admin.Post("/merchant", middleware.AdminOnly(false), handler.AddMerchant)
	admin.Post("/merchant/v2", middleware.AdminOnly(true), handler.AddMerchantV2)
	admin.Put("/merchant/:clientID", middleware.AdminOnly(true), handler.UpdateMerchant)
	admin.Put("/merchant/v2/:clientID", middleware.AdminOnly(true), handler.UpdateMerchantV2)
	admin.Get("/merchants", middleware.AdminOnly(false), handler.GetAllMerchants)
	admin.Get("/merchant/:clientID", middleware.AdminOnly(false), handler.GetMerchantByID)
	admin.Delete("/merchant/:clientID", middleware.AdminOnly(true), handler.DeleteMerchant)
}
