package main

import (
	"app/config"
	"app/database"
	"app/lib"
	"app/middleware"
	"app/repository"
	"app/router"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	// "app/webhook"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	config.SetupEnvFile()
	SetupLogfile()

	app := fiber.New(fiber.Config{
		Prefork:       true,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Redpay",
	})

	engine := html.New(filepath.Join("..", "views"), ".html")
	app = fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(middleware.TrackMetrics())

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, appid, appkey",
	}))

	middleware.PrometheusInit()

	db := database.ConnectDB()
	go lib.ProcessPendingTransactions()
	// go repository.ProcessTransactions()
	go repository.ProcessCallbackQueue()

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	router.SetupRoutes(app, db)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := app.ListenTLS(":443", "/home/aldi/mydomain.csr", "/home/aldi/mydomain.key") // Ganti dengan path sertifikat yang benar
		if err != nil {
			log.Fatalf("Error starting HTTPS server: %v", err)
		}
	}()

	<-sigs
	log.Println("Shutting down server...")

	time.Sleep(2 * time.Second)

	log.Println("Server stopped gracefully.")
}

func SetupLogfile() {
	logFile, err := os.OpenFile("../logs/dcb-new.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}
