package repository

import (
	"app/database"
	"app/dto/model"
	"context"
	"fmt"
	"log"
	"math"
	"strings"
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
	price,
    route,
    merchant_name,
    COUNT(*) AS total,
    SUM(price) AS revenue,
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
	price,
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
		Price          uint      `gorm:"column:price"`
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
			Date:          r.Date,
			Status:        MapStatusCodeToString(r.StatusCode),
			PaymentMethod: r.PaymentMethod,
			Amount:        r.Amount,
			Price:         r.Price,
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

func GetRouteFee(ctx context.Context, clientID, paymentMethod, route string) (float64, error) {
	// log.Printf("=== GET ROUTE FEE DEBUG ===")
	// log.Printf("Parameters - clientID: %s, paymentMethod: %s, route: '%s'", clientID, paymentMethod, route)

	// Check if route is empty
	if route == "" {
		// log.Printf("Route is empty, trying to get fee from PaymentMethodClient")
		var paymentMethodClient model.PaymentMethodClient
		err := database.DB.Where("client_id = ? AND name = ?", clientID, paymentMethod).First(&paymentMethodClient).Error
		if err != nil {
			log.Printf("Failed to get PaymentMethodClient: %v", err)
			return 0, err
		}
		return paymentMethodClient.Fee, nil
	}

	// Try to get from ChannelRouteWeight first
	var routeWeight model.ChannelRouteWeight
	err := database.DB.Where("client_id = ? AND payment_method = ? AND route = ?",
		clientID, paymentMethod, route).First(&routeWeight).Error

	if err != nil {
		// Jika tidak ada ChannelRouteWeight, coba ambil dari PaymentMethodClient sebagai fallback
		var paymentMethodClient model.PaymentMethodClient
		err2 := database.DB.Where("client_id = ? AND name = ?",
			clientID, paymentMethod).First(&paymentMethodClient).Error

		if err2 != nil {
			log.Printf("PaymentMethodClient fallback also failed: %v", err2)
			return 0, err2
		}

		//log.Printf("Using PaymentMethodClient fallback fee: %.2f", paymentMethodClient.Fee)
		return paymentMethodClient.Fee, nil
	}

	return routeWeight.Fee, nil
}

// GetTransactionReportWithMargin mendapatkan report transaksi dengan perhitungan margin
func GetTransactionReportWithMargin(ctx context.Context, startDate, endDate *time.Time, merchants []string, appID, clientUID, paymentMethods string) ([]model.TransactionMarginReport, error) {
	var summaries []model.TransactionMarginReport

	// log.Printf("=== NEW TRANSACTION REPORT WITH MARGIN ===")
	// log.Printf("Filters - merchants: %v, appID: %s, paymentMethods: %s, clientUID: %s", merchants, appID, paymentMethods, clientUID)
	// log.Printf("Date parameters - startDate: %v, endDate: %v", startDate, endDate)

	query := database.DB.Model(&model.Transactions{}).
		Select(`
			merchant_name,
			payment_method,
			route,
			clients.uid AS client_uid,
			COUNT(*) as count,
			SUM(amount) as total_amount,
			SUM(price) as total_amount_tax
		`).
		Joins("JOIN client_apps ON client_apps.app_id = transactions.app_id").
		Joins("JOIN clients ON clients.uid = client_apps.client_id").
		Where("status_code = ?", 1000).
		Group("merchant_name, payment_method, route, clients.uid").
		Having("COUNT(*) > 0")

	if startDate != nil && endDate != nil {
		query = query.Where("transactions.created_at BETWEEN ? AND ?", *startDate, *endDate)
	}

	if len(merchants) > 0 {
		query = query.Where("merchant_name IN ?", merchants)
		// log.Printf("Merchant filter: %v", merchants)
	} else if appID != "" {
		query = query.Where("app_id = ?", appID)
		// log.Printf("AppID filter: %s", appID)
	}

	if paymentMethods != "" {
		query = query.Where("payment_method = ?", paymentMethods)
		log.Printf("Payment method filter: %s", paymentMethods)
	}

	// Debug: Print the SQL query
	// sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
	// 	return tx.Find(&[]model.Transactions{})
	// })

	// log.Printf("Generated SQL: %s", sql)

	var totalCount int64
	database.DB.Model(&model.Transactions{}).Count(&totalCount)
	// log.Printf("Total transactions in table: %d", totalCount)

	var successCount int64
	database.DB.Model(&model.Transactions{}).Where("status_code = ?", 1000).Count(&successCount)

	// Execute query
	if err := query.Scan(&summaries).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction summary: %w", err)
	}

	// log.Printf("Found %d transaction summaries", len(summaries))
	// for i, summary := range summaries {
	// 	log.Printf("Summary %d: Merchant=%s, PaymentMethod=%s, Route=%s, Count=%d, TotalAmount=%d, TotalAmountTax=%d",
	// 		i+1, summary.MerchantName, summary.PaymentMethod, summary.Route, summary.Count, summary.TotalAmount, summary.TotalAmountTax)
	// }

	configsByClient := make(map[string][]model.SettlementClient)
	uniqueClients := make(map[string]struct{})
	for _, s := range summaries {
		if s.ClientUID != "" {
			uniqueClients[s.ClientUID] = struct{}{}
		}
	}
	for uid := range uniqueClients {
		settlementConfigs, err := GetSettlementConfig(uid)
		if err != nil {
			// log.Printf("Warning: Failed to get settlement config for client %s: %v", uid, err)
			continue
		}
		configsByClient[uid] = settlementConfigs
		// log.Printf("Loaded %d settlement configs for client %s", len(settlementConfigs), uid)
	}

	// Calculate margin for each summary
	for i := range summaries {
		// log.Printf("Processing summary %d: %s - %s - %s", i+1, summaries[i].MerchantName, summaries[i].PaymentMethod, summaries[i].Route)

		// Debug: Check what routes exist for this payment method (per client row)
		var existingRoutes []model.ChannelRouteWeight
		database.DB.Where("client_id = ? AND payment_method = ?", summaries[i].ClientUID, summaries[i].PaymentMethod).Find(&existingRoutes)
		// log.Printf("Existing routes for %s: %v", summaries[i].PaymentMethod, existingRoutes)

		// Get fee for this payment method and route
		fee, err := GetRouteFee(ctx, summaries[i].ClientUID, summaries[i].PaymentMethod, summaries[i].Route)
		if err != nil {
			fee = 0
			log.Printf("Fee not found for %s-%s, using 0", summaries[i].PaymentMethod, summaries[i].Route)
		}

		// Find settlement config for this specific payment method and client in this row
		var settlementConfig *model.SettlementClient
		var shareMerchantPercentage float32
		clientConfigs := configsByClient[summaries[i].ClientUID]
		for _, settlement := range clientConfigs {
			if settlement.Name == summaries[i].PaymentMethod {
				settlementConfig = &settlement
				if settlement.SharePartner != nil {
					shareMerchantPercentage = *settlement.SharePartner
				}
				break
			}
		}

		if settlementConfig == nil {
			log.Printf("No settlement config found for client %s and payment method %s, using 0%%", summaries[i].ClientUID, summaries[i].PaymentMethod)
		}

		// Calculate share redision using new formula: amount - (amount * shareMerchant)
		shareRedision := calculateShareRedisionNew(summaries[i].TotalAmount, shareMerchantPercentage)

		// Calculate share merchant amount
		shareMerchantAmountGross := summaries[i].TotalAmount - shareRedision

		// Set ShareRedisionPercentage safely
		var shareRedisionPercentage float32
		if settlementConfig != nil && settlementConfig.ShareRedision != nil {
			shareRedisionPercentage = *settlementConfig.ShareRedision
		}

		var (
			additionalFee uint
			bhpUSO        uint
			tax23         uint
			shareSupplier uint
		)
		if strings.ToLower(settlementConfig.IsBhpuso) == "1" {
			bhpUSO = uint(float64(shareMerchantAmountGross) * 0.0175)
		}

		if settlementConfig.AdditionalFee != nil && *settlementConfig.AdditionalFee == 1 {
			additionalFee = uint(float64(shareMerchantAmountGross) * 0.05)
		}

		if settlementConfig.Tax23 != nil && strings.ToLower(*settlementConfig.Tax23) == "1" {
			tax23 = uint(float64(shareMerchantAmountGross) * 0.02)
		}

		// Perbaiki tipe data agar tidak terjadi mismatched types (uint64 dan float64)
		shareSupplier = uint(summaries[i].TotalAmount - uint64(math.Round(float64(summaries[i].TotalAmount)*fee/100)))
		// shareSupplierInc adalah shareSupplier ditambah 11%
		shareSupplierInc := shareSupplier + (shareSupplier*11)/100

		// Perhitungan BHP USO supplier dan PPH supplier berdasarkan payment method
		var bhpUsoSupplier uint
		var pphSupplier uint

		paymentMethod := strings.ToLower(summaries[i].PaymentMethod)

		// Untuk telkomsel_airtime dan xl_airtime: gunakan BHP USO supplier + PPH supplier
		if paymentMethod == "telkomsel_airtime" || paymentMethod == "xl_airtime" {
			// bhpUsoSupplier: 1.75% dari shareSupplier -> (shareSupplier * 175) / 10000
			bhpUsoSupplier = (shareSupplier * 175) / 10000
			// pphSupplier: 2% dari shareSupplier -> (shareSupplier * 2) / 100
			pphSupplier = (shareSupplier * 2) / 100
		} else if paymentMethod == "smartfren_airtime" || paymentMethod == "indosat_airtime" || paymentMethod == "three_airtime" {
			// Untuk smartfren_airtime, indosat_airtime, dan three_airtime: hanya BHP USO supplier
			bhpUsoSupplier = (shareSupplier * 175) / 10000
			pphSupplier = 0
		} else {
			// Untuk payment method lainnya: tidak menggunakan keduanya
			bhpUsoSupplier = 0
			pphSupplier = 0
		}

		shareSupplierNett := shareSupplierInc - bhpUsoSupplier - pphSupplier
		shareSupplierNettExc := (shareSupplierNett * 100) / 111
		shareMerchantAmountNett := uint(shareMerchantAmountGross) - bhpUSO - additionalFee - tax23

		// Izinkan margin negatif: simpan sebagai int64
		margin := int64(shareSupplierNettExc) - int64(shareMerchantAmountNett)
		summaries[i].Margin = margin
		summaries[i].ShareSupplier = shareSupplier
		summaries[i].ShareSupplierInc = shareSupplierInc
		summaries[i].BhpUsoSupplier = bhpUsoSupplier
		summaries[i].PphSupplier = pphSupplier
		summaries[i].ShareSupplierNett = shareSupplierNett
		summaries[i].ShareRedision = uint(shareRedision)
		summaries[i].ShareRedisionPercentage = shareRedisionPercentage
		summaries[i].ShareMerchantPercentage = shareMerchantPercentage
		summaries[i].ShareMerchant = uint(shareMerchantAmountNett)
		summaries[i].Fee = fee

	}

	return summaries, nil
}

// calculateShareRedision menghitung share redision berdasarkan amount dan share percentage
func calculateShareRedision(amount uint64, shareRedisionPercentage float32) uint64 {
	if shareRedisionPercentage <= 0 {
		return amount
	}

	shareRedision := float64(amount) - (float64(amount) * float64(shareRedisionPercentage) / 100)
	return uint64(math.Round(shareRedision))
}

// calculateShareRedisionNew menghitung share redision menggunakan formula baru: amount - (amount * shareMerchant)
func calculateShareRedisionNew(amount uint64, shareMerchantPercentage float32) uint64 {
	if shareMerchantPercentage <= 0 {
		return amount
	}

	shareRedision := float64(amount) - (float64(amount) * float64(shareMerchantPercentage) / 100)
	return uint64(math.Round(shareRedision))
}

// calculateMargin menghitung margin berdasarkan share redision dan fee
func calculateMargin(shareRedision uint64, fee float64) uint64 {
	if fee <= 0 {
		return shareRedision
	}

	margin := float64(shareRedision) - (float64(shareRedision) * fee / 100)
	return uint64(math.Round(margin))
}
