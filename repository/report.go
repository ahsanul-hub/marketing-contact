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

	jakartaLocation, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {

		log.Printf("Warning: Could not load 'Asia/Jakarta' timezone, falling back to fixed offset. Error: %v", err)
		jakartaLocation = time.FixedZone("WIB", 7*60*60)
	}

	nowWIB := time.Now().In(jakartaLocation)

	if startDate.IsZero() {
		startDate = time.Date(
			nowWIB.Year(), nowWIB.Month(), nowWIB.Day()-7,
			0, 0, 0, 0, jakartaLocation,
		)
	} else {
		startDate = time.Date(
			startDate.In(jakartaLocation).Year(),
			startDate.In(jakartaLocation).Month(),
			startDate.In(jakartaLocation).Day(),
			0, 0, 0, 0,
			jakartaLocation,
		)
	}

	if endDate.IsZero() {
		endDate = time.Date(
			nowWIB.Year(), nowWIB.Month(), nowWIB.Day()+1,
			0, 0, 0, -1, jakartaLocation,
		)
	} else {
		endDate = time.Date(
			endDate.In(jakartaLocation).Year(),
			endDate.In(jakartaLocation).Month(),
			endDate.In(jakartaLocation).Day()+1,
			0, 0, 0, -1,
			jakartaLocation,
		)
	}

	// log.Println("startDate", startDate)
	// log.Println("endDate", endDate)
	query := database.DB.Table("transactions").Select(`
    DATE_TRUNC('day', created_at AT TIME ZONE 'Asia/Jakarta') AS date,
    status_code,
    payment_method,
    amount,
    route,
    merchant_name,
    COUNT(*) AS total,
    SUM(amount) AS revenue,
    MIN(created_at) AS first_created_at,
    MAX(created_at) AS last_created_at
`)

	query = query.Where("created_at BETWEEN ? AND ?", startDate, endDate)

	if merchantName != "" {
		query = query.Where("merchant_name = ?", merchantName)
	}
	if status != "" {
		statusCode := -1
		switch status {
		case "success":
			statusCode = 1000
		case "pending":
			statusCode = 1001
		case "failed":
			statusCode = 1005
		}
		if statusCode != -1 {
			query = query.Where("status_code = ?", statusCode)
		} else {
			log.Printf("Warning: Invalid status string provided: %s", status)
		}
	}
	if paymentMethod != "" {
		query = query.Where("payment_method = ?", paymentMethod)
	}
	if route != "" {
		query = query.Where("route = ?", route)
	}

	query = query.Group(`
    DATE_TRUNC('day', created_at AT TIME ZONE 'Asia/Jakarta'),
    status_code,
    payment_method,
    amount,
    route,
    merchant_name
`).Order(`
    DATE_TRUNC('day', created_at AT TIME ZONE 'Asia/Jakarta') DESC,
    merchant_name,
    payment_method,
    amount,
    status_code
`)

	// Define a struct to scan the raw query results into
	type Result struct {
		Date           time.Time `gorm:"column:date"`
		StatusCode     int       `gorm:"column:status_code"`
		PaymentMethod  string    `gorm:"column:payment_method"`
		Amount         uint      `gorm:"column:amount"`
		Route          string    `gorm:"column:route"`
		MerchantName   string    `gorm:"column:merchant_name"`
		Total          int       `gorm:"column:total"`
		Revenue        float64   `gorm:"column:revenue"`
		FirstCreatedAt time.Time `gorm:"column:first_created_at"`
		LastCreatedAt  time.Time `gorm:"column:last_created_at"`
	}

	var results []Result
	if err := query.Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get daily transaction summary: %w", err)
	}

	for _, r := range results {
		summaries = append(summaries, model.TransactionDailySummary{
			Date:          r.Date.Format(time.RFC3339),
			Status:        MapStatusCodeToString(r.StatusCode),
			PaymentMethod: r.PaymentMethod,
			Amount:        r.Amount,
			Route:         r.Route,
			MerchantName:  r.MerchantName,
			Total:         r.Total,
			Revenue:       r.Revenue,
			// FirstCreatedAt: r.FirstCreatedAt,
			// LastCreatedAt:  r.LastCreatedAt,
		})
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

func MapStatusCodeToString(statusCode int) string {
	switch statusCode {
	case 1000:
		return "success"
	case 1003:
		return "success"
	case 1001:
		return "pending"
	case 1005:
		return "failed"
	default:
		return fmt.Sprintf("unknown (%d)", statusCode)
	}
}
