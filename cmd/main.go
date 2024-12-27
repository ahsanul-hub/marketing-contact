package main

import (
	"app/config"
	"app/database"
	"app/lib"
	"app/repository"
	"app/router"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
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

	database.ConnectDB()
	go lib.ProcessPendingTransactions()
	go repository.ProcessCallbackQueue()
	go repository.ProcessTransactions()

	router.SetupRoutes(app)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Menggunakan HTTPS
	go func() {
		err := app.ListenTLS(":443", "/home/aldi/mydomain.crt", "/home/aldi/mydomain.key") // Ganti dengan path sertifikat yang benar
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
