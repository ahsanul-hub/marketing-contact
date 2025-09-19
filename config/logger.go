// package config

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"time"
// )

// func SetupLogfile() {
// 	logDir := "../logs"
// 	err := os.MkdirAll(logDir, 0755)
// 	if err != nil {
// 		fmt.Printf("Failed to create log directory: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Tentukan nama file berdasarkan tahun, bulan, dan minggu
// 	currentTime := time.Now()
// 	year, month, _ := currentTime.Date()
// 	_, week := currentTime.ISOWeek()

// 	logFilename := filepath.Join(logDir,
// 		fmt.Sprintf("dcb-new-%d-%02d-week%d.log", year, month, week))

// 	logFile, err := os.OpenFile(logFilename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
// 	if err != nil {
// 		fmt.Printf("Failed to open log file: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// MultiWriter untuk stdout & file log
// 	mw := io.MultiWriter(os.Stdout, logFile)

// 	// Mengarahkan log default Golang ke MultiWriter
// 	log.SetOutput(mw)

// 	// Menambahkan format log dengan timestamp dan lokasi file
// 	log.SetFlags(log.LstdFlags | log.Lshortfile)

// 	log.Println("Logging initialized") // Ini akan muncul di terminal & file log
// }

package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	MainLogger  *log.Logger
	mainLogFile *os.File
)

// ANSI color codes for console
var (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func InitLoggers() error {
	logDir := LogsBaseDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	currentTime := time.Now()
	year, month, _ := currentTime.Date()
	_, week := currentTime.ISOWeek()

	mainLogName := fmt.Sprintf("dcb-new-%d-%02d-week%d.log", year, month, week)

	var err error
	mainLogFile, err = os.OpenFile(filepath.Join(logDir, mainLogName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	MainLogger = log.New(mainLogFile, "", 0)

	// Initialize payment loggers
	if err := InitPaymentLoggers(); err != nil {
		log.Printf("Failed to initialize payment loggers: %v", err)
		// Continue without payment loggers
	}

	return nil
}

// LogsBaseDir returns absolute path to logs directory at project root
func LogsBaseDir() string {
	// Start from current working directory and walk up to find go.mod
	cwd, err := os.Getwd()
	if err != nil {
		return "logs"
	}
	dir := cwd
	for i := 0; i < 6; i++ { // limit upward traversal to avoid infinite loops
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return filepath.Join(dir, "logs")
		}
		parent := filepath.Dir(dir)
		if parent == dir { // reached filesystem root
			break
		}
		dir = parent
	}
	// Fallback: if running from cmd/, go one level up
	if filepath.Base(cwd) == "cmd" {
		return filepath.Join("..", "logs")
	}
	return "logs"
}

// Set default logger output for log.SetOutput
func SetDefaultLoggerOutput() {
	mw := io.MultiWriter(os.Stdout, mainLogFile)
	log.SetOutput(mw)
}

// LogWithLevel logs with level, file:line, and color for console, JSON for file
func LogWithLevel(level, format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	shortFile := file
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		shortFile = file[idx+1:]
	}
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	msg := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("%s [%s] %s:%d: %s", timestamp, strings.ToUpper(level), shortFile, line, msg)

	// Write to file (no color, plain text) - hanya jika MainLogger sudah diinisialisasi
	if MainLogger != nil && mainLogFile != nil {
		MainLogger.Println(logLine)
		mainLogFile.Sync()
	}

	// Write to console (with color)
	var color string
	switch strings.ToUpper(level) {
	case "INFO":
		color = colorGreen
	case "WARN":
		color = colorYellow
	case "ERROR":
		color = colorRed
	case "DEBUG":
		color = colorCyan
	default:
		color = colorWhite
	}
	fmt.Printf("%s%s%s\n", color, logLine, colorReset)
}

// LogJSONWithLevel logs a JSON entry with level and message to the given logger
func LogJSONWithLevel(logger *log.Logger, level, message string, data map[string]interface{}) {
	entry := map[string]interface{}{
		"level":     level,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	for k, v := range data {
		entry[k] = v
	}
	b, _ := json.Marshal(entry)
	logger.Println(string(b))
}

// Convenience wrappers
func LogUtama(level, message string, data map[string]interface{}) {
	LogJSONWithLevel(MainLogger, level, message, data)
}

// Helper for slow response warning
func LogIfSlowResponse(start time.Time, endpoint, method string, threshold time.Duration) {
	dur := time.Since(start)
	if dur > threshold {
		LogWithLevel("WARN", "Slow response: %.2fs for %s %s", dur.Seconds(), method, endpoint)
	} else {
		LogWithLevel("INFO", "Response: %.2fs for %s %s", dur.Seconds(), method, endpoint)
	}
}
