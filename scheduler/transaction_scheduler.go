package scheduler

import (
	"app/config"
	"app/dto/model"
	"app/repository"
	"app/service"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type TransactionScheduler struct {
	cron        *cron.Cron
	sftpService *service.SFTPService
}

func NewTransactionScheduler() *TransactionScheduler {
	// Gunakan timezone WIB untuk cron
	wibLocation, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Error loading WIB location: %v, using UTC", err)
		wibLocation = time.UTC
	}

	log.Printf("Scheduler timezone: %s", wibLocation.String())

	cronInstance := cron.New(cron.WithLocation(wibLocation))

	return &TransactionScheduler{
		cron:        cronInstance,
		sftpService: service.NewSFTPService(),
	}
}

func (ts *TransactionScheduler) Start() {
	// Jalankan scheduler setiap jam 09:00 WIB
	// Format cron: "00 09 * * *" (menit jam hari bulan hari_minggu)

	// Log waktu saat ini untuk debugging
	now := time.Now()
	log.Printf("Current time when starting scheduler: %s", now.Format("2006-01-02 15:04:05"))
	log.Printf("Current timezone: %s", now.Location().String())

	// Main scheduler job - berjalan setiap hari jam 09:00 WIB
	entryID, err := ts.cron.AddFunc("00 09 * * *", func() {
		log.Println("=== MAIN SCHEDULER RUNNING AT 09:00 WIB ===")
		ts.sendTransactionReport()
	})
	if err != nil {
		log.Printf("Error scheduling transaction report: %v", err)
		return
	}

	log.Printf("Transaction scheduler started with entry ID: %d - will run daily at 09:00 WIB", entryID)

	// Log waktu saat ini dan next run time
	entries := ts.cron.Entries()
	log.Printf("Total cron entries: %d", len(entries))
	for _, entry := range entries {
		log.Printf("Cron Entry ID: %d, Next Run: %s, Schedule: %v",
			entry.ID,
			entry.Next.Format("2006-01-02 15:04:05"),
			entry.Schedule)
	}

	ts.cron.Start()
}

func (ts *TransactionScheduler) Stop() {
	ts.cron.Stop()
}

func (ts *TransactionScheduler) GetStatus() map[string]interface{} {
	entries := ts.cron.Entries()
	status := make(map[string]interface{})

	for i, entry := range entries {
		status[fmt.Sprintf("entry_%d", i)] = map[string]interface{}{
			"id":       entry.ID,
			"next_run": entry.Next.Format("2006-01-02 15:04:05"),
			"schedule": fmt.Sprintf("%v", entry.Schedule),
		}
	}

	return status
}

func (ts *TransactionScheduler) calculateDateRange() (time.Time, time.Time) {
	// Ambil tanggal kemarin untuk laporan harian
	// Karena server UTC, kurangi 7 jam untuk mendapatkan waktu WIB yang benar
	utcNow := time.Now().UTC()
	wibNow := utcNow.Add(-7 * time.Hour)

	// Data transaksi dimulai dari jam 07:00 WIB
	// Jadi range: kemarin jam 07:00 WIB sampai hari ini jam 06:59 WIB
	// Konversi ke UTC: kemarin jam 17:00 UTC sampai hari ini jam 16:59 UTC
	yesterday := wibNow.AddDate(0, 0, -1)
	today := wibNow

	// Jika sekarang tanggal 8, maka:
	// - startDate: 2025-08-05 17:00:00 UTC (kemarin-1 jam 17:00 UTC)
	// - endDate: 2025-08-06 16:59:59 UTC (kemarin jam 16:59 UTC)
	startDate := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 17, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
	endDate := time.Date(today.Year(), today.Month(), today.Day(), 16, 59, 59, 999999999, time.UTC).AddDate(0, 0, -1)

	return startDate, endDate
}

func (ts *TransactionScheduler) sendTransactionReport() {
	log.Println("=== STARTING SCHEDULED TRANSACTION REPORT ===")
	log.Println("Starting scheduled transaction report generation and SFTP upload...")

	// Hitung range waktu yang benar
	startDate, endDate := ts.calculateDateRange()

	// Log informasi timezone
	utcNow := time.Now().UTC()
	wibNow := utcNow.Add(-7 * time.Hour)
	yesterday := wibNow.AddDate(0, 0, -1)

	log.Printf("=== TIMEZONE INFO ===")
	log.Printf("UTC Now: %s", utcNow.Format("2006-01-02 15:04:05"))
	log.Printf("WIB Now: %s", wibNow.Format("2006-01-02 15:04:05"))
	log.Printf("Yesterday WIB: %s", yesterday.Format("2006-01-02 15:04:05"))
	log.Printf("Date range: %s to %s (UTC time for WIB 07:00-06:59)", startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05"))

	// Ambil data transaksi untuk semua merchant yang memerlukan laporan SFTP
	merchants := ts.getMerchantsWithSFTP()
	log.Printf("Found %d merchants configured for SFTP", len(merchants))

	for _, merchant := range merchants {
		log.Printf("Processing merchant: %s", merchant.ClientName)
		go ts.processMerchantReport(merchant, startDate, endDate)
	}

	log.Println("=== SCHEDULED TRANSACTION REPORT COMPLETED ===")
}

func (ts *TransactionScheduler) getMerchantsWithSFTP() []service.MerchantSFTPConfig {
	// Daftar merchant yang memerlukan laporan SFTP
	// Ini bisa diambil dari database atau environment variable

	// Generate folder name berdasarkan tahun dan bulan saat ini
	currentTime := time.Now()
	folderName := currentTime.Format("200601") // Format: YYYYMM

	return []service.MerchantSFTPConfig{
		{
			ClientName: "Zingplay International PTE,. LTD",
			AppID:      "CKxpZpt29Cx3BjOJ0CItnQ",
			SFTPHost:   config.Config("SFTP_HOST_1", ""),
			SFTPPort:   config.Config("SFTP_PORT_1", "22"),
			SFTPUser:   config.Config("SFTP_USER_1", ""),
			SFTPPass:   config.Config("SFTP_PASS_1", ""),
			RemotePath: fmt.Sprintf("/%s/", folderName),
			FileName:   "Zingplay-%s.xlsx", // Format: Zingplay-20250807.xlsx
		},
	}
}

func (ts *TransactionScheduler) processMerchantReport(merchant service.MerchantSFTPConfig, startDate, endDate time.Time) {
	log.Printf("Processing report for merchant: %s", merchant.ClientName)

	ctx := context.Background()

	// Ambil data transaksi untuk merchant ini
	transactions, err := repository.GetTransactionsByDateRange(
		ctx,
		0, // status 0 = semua status
		&startDate,
		&endDate,
		"", // payment method kosong = semua payment method
		[]string{merchant.ClientName},
		[]string{merchant.AppID}, // appID kosong = semua app
	)

	if err != nil {
		log.Printf("Error getting transactions for merchant %s: %v", merchant.ClientName, err)
		return
	}

	if len(transactions) == 0 {
		log.Printf("No transactions found for merchant %s on %s", merchant.ClientName, startDate.Format("2006-01-02"))
		return
	}

	folderName := startDate.Format("200601") // Format: YYYYMM
	merchant.RemotePath = fmt.Sprintf("/%s/", folderName)

	log.Printf("Using folder: %s for transactions on %s", merchant.RemotePath, startDate.Format("2006-01-02"))

	// Generate Excel file
	excelData, err := ts.generateExcelReport(transactions, merchant.ClientName)
	if err != nil {
		log.Printf("Error generating Excel report for merchant %s: %v", merchant.ClientName, err)
		return
	}

	// Upload ke SFTP
	fileName := fmt.Sprintf(merchant.FileName, startDate.Format("20060102")) // Format: Zingplay-20250807.xlsx
	err = ts.sftpService.UploadFile(merchant, fileName, excelData)
	if err != nil {
		log.Printf("Error uploading file to SFTP for merchant %s: %v", merchant.ClientName, err)
		return
	}

	log.Printf("Successfully uploaded transaction report for merchant %s: %s", merchant.ClientName, fileName)
}

func (ts *TransactionScheduler) generateExcelReport(transactions []model.Transactions, merchantName string) ([]byte, error) {
	// Gunakan fungsi yang sudah ada di service
	return service.GenerateExcelReport(transactions, merchantName)
}
