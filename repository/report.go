package repository

import (
	"app/database"
	"app/dto/model"
	"context"
	"fmt"
	"log"
	"time"
)

func GetTransactionReport(ctx context.Context, startDate, endDate *time.Time, merchants []string, appID, clientUid, paymentMethods string) ([]model.TransactionReport, *model.SettlementClient, error) {
	var summaries []model.TransactionReport

	settlementConfig, err := GetSettlementConfig(clientUid)
	if err != nil {
		log.Println("Error GetSettlementConfig:", err)
	}

	var selectedSettlement *model.SettlementClient
	for _, settlement := range settlementConfig {
		if settlement.Name == paymentMethods {
			selectedSettlement = &settlement
			break
		}
	}

	if selectedSettlement == nil {
		log.Println("selectedSettlement nil, check paymentMethod:", paymentMethods)
	}

	query := database.DB.Model(&model.Transactions{}).
		Select(`
			merchant_name,
			payment_method,
			amount,
			price as amount_tax,
			COUNT(*) as count,
			amount * COUNT(*) as total_amount,
			price * COUNT(*) as total_amount_tax
		`).
		Where("status_code = ?", 1000).
		Group("merchant_name, payment_method, amount, price")

	if startDate != nil && endDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", *startDate, *endDate)
	}
	if len(merchants) > 0 {
		query = query.Where("merchant_name IN ?", merchants)
	} else if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	if len(paymentMethods) > 0 {
		query = query.Where("payment_method = ?", paymentMethods)
	}

	if err := query.Scan(&summaries).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to get transaction summary per client: %w", err)
	}

	return summaries, selectedSettlement, nil
}

func GetTransactionSummaryDaily(startDate, endDate time.Time, merchantName, status, paymentMethod, route string) ([]model.TransactionDailySummary, error) {
	var summaries []model.TransactionDailySummary

	locJakarta, _ := time.LoadLocation("Asia/Jakarta")

	now := time.Now().In(locJakarta)

	// Set default tanggal
	if startDate.IsZero() {
		startDate = now.AddDate(0, 0, -4)
	}
	if endDate.IsZero() {
		endDate = now
	}

	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, locJakarta)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, locJakarta)

	query := database.DB.Model(&model.Transactions{}).
		Select(`
		DATE(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Jakarta') AS date,
		CASE
			WHEN status_code = 1001 THEN 'pending'
			WHEN status_code IN (1000, 1003) THEN 'success'
			ELSE 'failed'
		END AS status,
		payment_method,
		amount,
		route,
		merchant_name,
		COUNT(*) AS total,
		CAST(amount * COUNT(*) AS FLOAT) AS revenue
	`).
		Where(`
		(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Jakarta') BETWEEN ? AND ?
	`, startDate, endDate).
		Group(`
		DATE(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Jakarta'),
		status,
		payment_method,
		amount,
		route,
		merchant_name
	`)

	if merchantName != "" {
		query = query.Where("merchant_name = ?", merchantName)
	}
	if status != "" {
		// Ubah status string ke status_code sesuai mapping
		switch status {
		case "pending":
			query = query.Where("status_code = ?", 1001)
		case "success":
			query = query.Where("status_code IN ?", []int{1000, 1003})
		case "failed":
			query = query.Where("status_code NOT IN ?", []int{1000, 1001, 1003})
		}
	}
	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}
	if route != "" {
		query = query.Where("route = ?", route)
	}

	err := query.Scan(&summaries).Error
	if err != nil {
		return nil, err
	}

	for i, summary := range summaries {
		// Parse summary.Date (yyyy-mm-dd) ke time.Time
		fmt.Printf("DEBUG summary.Date: %v\n", summary.Date)

		summaryDate, err := time.Parse("2006-01-02", summary.Date)
		if err != nil {
			continue
		}

		startOfDayJakarta := time.Date(summaryDate.Year(), summaryDate.Month(), summaryDate.Day(), 0, 0, 0, 0, locJakarta)
		endOfDayJakarta := time.Date(summaryDate.Year(), summaryDate.Month(), summaryDate.Day(), 23, 59, 59, 999999999, locJakarta)

		startOfDay := startOfDayJakarta.UTC()
		endOfDay := endOfDayJakarta.UTC()

		fmt.Printf("[DEBUG] Checking summary for %s | Start: %s | End: %s\n", summary.Date, startOfDay, endOfDay)

		// Mapping status ke status_code
		var statusCodes []int
		switch summary.Status {
		case "success":
			statusCodes = []int{1000, 1003}
		case "pending":
			statusCodes = []int{1001}
		default:
			statusCodes = []int{} // berarti NOT IN (1000,1001,1003)
		}

		fmt.Printf("[DEBUG] Querying range %s - %s for %s, %s, amount: %d, route: %s\n",
			startOfDay.UTC().Format(time.RFC3339),
			endOfDay.UTC().Format(time.RFC3339),
			summary.MerchantName,
			summary.PaymentMethod,
			summary.Amount,
			summary.Route,
		)

		// Query First
		var first model.Transactions
		firstQuery := database.DB.Model(&model.Transactions{}).
			Where("merchant_name = ? AND payment_method = ? AND amount = ? AND route = ? AND created_at BETWEEN ? AND ?", summary.MerchantName, summary.PaymentMethod, summary.Amount, summary.Route, startOfDay.UTC(), endOfDay.UTC()).
			Order("created_at ASC").
			Limit(1)

		if len(statusCodes) > 0 {
			firstQuery = firstQuery.Where("status_code IN ?", statusCodes)
		} else {
			firstQuery = firstQuery.Where("status_code NOT IN ?", []int{1000, 1001, 1003})
		}

		firstQuery.Find(&first)

		fmt.Printf("[DEBUG] first.ID = %v, first.CreatedAt = %v\n", first.ID, first.CreatedAt)

		// Query Last
		var last model.Transactions
		lastQuery := database.DB.Model(&model.Transactions{}).
			Where("merchant_name = ? AND payment_method = ? AND amount = ? AND route = ? AND created_at BETWEEN ? AND ?", summary.MerchantName, summary.PaymentMethod, summary.Amount, summary.Route, startOfDay.UTC(), endOfDay.UTC()).
			Order("created_at DESC").
			Limit(1)

		if len(statusCodes) > 0 {
			lastQuery = lastQuery.Where("status_code IN ?", statusCodes)
		} else {
			lastQuery = lastQuery.Where("status_code NOT IN ?", []int{1000, 1001, 1003})
		}

		lastQuery.Find(&last)

		// Assign ke summary
		summaries[i].FirstCreatedAt = first.CreatedAt
		summaries[i].LastCreatedAt = last.CreatedAt
		summaries[i].FirstTransactionID = first.ID
		summaries[i].LastTransactionID = last.ID
	}

	return summaries, nil
}

// func GetTransactionReportDaily() ([]model.TransactionDailyReport, error) {
// 	var reports []model.TransactionDailyReport

// 	rows, err := database.DB.Raw(`
// 		SELECT
// 			merchant_name,
// 			payment_method,
// 			COUNT(CASE WHEN status_code = 200 THEN 1 END) AS success,
// 			COUNT(CASE WHEN status_code IN (0, 100) THEN 1 END) AS pending,
// 			COUNT(CASE WHEN status_code NOT IN (200, 0, 100) THEN 1 END) AS failed
// 		FROM transactions
// 		GROUP BY merchant_name, payment_method
// 		ORDER BY merchant_name, payment_method
// 	`).Rows()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var row model.TransactionDailyReport
// 		if err := rows.Scan(&row.MerchantName, &row.PaymentMethod, &row.SuccessCount, &row.PendingCount, &row.FailedCount); err != nil {
// 			log.Println("Scan error:", err)
// 			continue
// 		}
// 		reports = append(reports, row)
// 	}

// 	return reports, nil
// }
