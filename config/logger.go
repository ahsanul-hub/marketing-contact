package config

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func SetupLogfile() {
	logDir := "../logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	// Tentukan nama file berdasarkan tahun, bulan, dan minggu
	currentTime := time.Now()
	year, month, _ := currentTime.Date()
	_, week := currentTime.ISOWeek()

	logFilename := filepath.Join(logDir,
		fmt.Sprintf("dcb-new-%d-%02d-week%d.log", year, month, week))

	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	// MultiWriter untuk stdout & file log
	mw := io.MultiWriter(os.Stdout, logFile)

	// Mengarahkan log default Golang ke MultiWriter
	log.SetOutput(mw)

	// Menambahkan format log dengan timestamp dan lokasi file
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Logging initialized") // Ini akan muncul di terminal & file log
}
