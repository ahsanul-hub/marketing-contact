package main

import (
	"app/config"
	"app/database"
	"app/lib"
	"app/repository"
	"app/router"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	// "app/webhook"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	// "github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config.SetupEnvFile()

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

	database.ConnectDB()
	go lib.ProcessPendingTransactions()
	go repository.ProcessCallbackQueue()
	go repository.ProcessTransactions()

	// Mengatur rute sebelum memulai server
	router.SetupRoutes(app)

	// Menangani shutdown graceful
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Jalankan server dalam goroutine
	go func() {
		if err := app.Listen(":4000"); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Tunggu sinyal untuk shutdown
	<-sigs
	log.Println("Shutting down server...")

	// Memberikan waktu untuk menyelesaikan permintaan yang sedang berjalan
	time.Sleep(2 * time.Second) // Atau gunakan mekanisme lain untuk menunggu

	log.Println("Server stopped gracefully.")
}
