package handler

import (
	"app/dto/model"
	"app/repository"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ReportResult struct {
	Summaries              []model.TransactionReport `json:"summaries"`
	AdditionalFee          uint                      `json:"additional_fee"`
	BhpUSO                 uint                      `json:"bhp_uso"`
	Tax23                  uint                      `json:"tax_23"`
	ServiceChargeharge     uint                      `json:"service_charge"`
	GrandTotalRedision     uint                      `json:"grand_total_redision"`
	TotalMerchant          uint                      `json:"total_merchant"`
	GrandTotal             uint                      `json:"grand_total"`
	TotalTransactionAmount uint                      `json:"total_transaction_amount"`
	TotalTransaction       uint                      `json:"total_transaction"`
	ShareRedision          float32                   `json:"share_redision"`
	Mdr                    string                    `json:"mdr"`
	ShareMerchant          float32                   `json:"share_merchant"`
}

func GetReport(c *fiber.Ctx) error {
	ctx := c.Context()

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	clientUID := c.Query("client_uid")
	appID := c.Query("app_id")

	var startDate, endDate *time.Time

	if startDateStr != "" {
		parsedStart, err := time.Parse(time.RFC1123, startDateStr)
		if err == nil {
			startDate = &parsedStart
		}
	}

	if endDateStr != "" {
		parsedEnd, err := time.Parse(time.RFC1123, endDateStr)
		if err == nil {
			endDate = &parsedEnd
		}
	}

	merchantNameStr := c.Query("merchant_name")
	var merchants []string
	if merchantNameStr != "" {
		merchants = strings.Split(merchantNameStr, ",")
	} else {
		merchants = []string{}
	}
	paymentMethods := c.Query("payment_method")

	summaries, settlementConfig, err := repository.GetTransactionReport(ctx, startDate, endDate, merchants, appID, clientUID, paymentMethods)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var (
		grandTotalRedision     uint
		totalMerchant          uint
		additionalFee          uint
		bhpUSO                 uint
		tax23                  uint
		totalTransactionAmount uint64
		totalTransaction       uint
	)

	// if settlementConfig != nil {
	// 	for i := range summaries {
	// 		totalAmount := summaries[i].TotalAmount
	// 		totalTransactionAmount += totalAmount
	// 		totalTransaction += uint(summaries[i].Count)

	// 		if settlementConfig.ShareRedision != nil {
	// 			shareRed := uint(math.Ceil((float64(totalAmount) * float64(*settlementConfig.ShareRedision)) / 100))
	// 			summaries[i].ShareRedision = shareRed
	// 			grandTotalRedision += shareRed

	// 			shareMerchant := totalAmount - uint64(shareRed)
	// 			summaries[i].ShareMerchant = uint(shareMerchant)
	// 			totalMerchant += uint(shareMerchant)
	// 		} else {
	// 			shareMerch := uint((totalAmount * uint64(*settlementConfig.SharePartner)) / 100)
	// 			summaries[i].ShareMerchant = shareMerch
	// 			totalMerchant += shareMerch
	// 		}
	// 	}

	// 	if strings.ToLower(settlementConfig.IsBhpuso) == "1" {
	// 		bhpUSO = uint(float64(totalMerchant) * 0.0175)
	// 	}

	if settlementConfig.AdditionalFee != nil && *settlementConfig.AdditionalFee == 1 {
		additionalFee = uint(float64(totalMerchant) * 0.05)
	}

	// 	if settlementConfig.Tax23 != nil && strings.ToLower(*settlementConfig.Tax23) == "1" {
	// 		tax23 = uint(float64(totalMerchant) * 0.02)
	// 	}
	// }

	if settlementConfig != nil {
		for i := range summaries {
			totalAmount := summaries[i].TotalAmount
			totalTransaction += uint(summaries[i].Count)

			if settlementConfig.ShareRedision != nil {
				shareRedFloat := float64(totalAmount) * float64(*settlementConfig.ShareRedision) / 100
				if settlementConfig.MdrType == "fix" && *settlementConfig.ShareRedision == 0 {
					// Attempt to parse MDR from string to float64
					mdrFloat, err := strconv.ParseFloat(settlementConfig.Mdr, 64)
					if err == nil {
						shareRedFloat = float64(summaries[i].Count) * mdrFloat
					} else {
						shareRedFloat = 0 // fallback if MDR can't be parsed
					}
				}

				shareRed := uint(math.Round(shareRedFloat))
				summaries[i].ShareRedision = shareRed
				grandTotalRedision += shareRed

				shareMerchant := totalAmount - uint64(shareRed)
				summaries[i].ShareMerchant = uint(shareMerchant)
				totalMerchant += uint(shareMerchant)
			} else if settlementConfig.SharePartner != nil {
				shareMerch := uint(math.Round(float64(totalAmount) * float64(*settlementConfig.SharePartner) / 100))
				summaries[i].ShareMerchant = shareMerch
				totalMerchant += shareMerch
			}
		}

		if settlementConfig.Mdr != "" && *settlementConfig.ShareRedision != 0 {
			mdrFloat, err := strconv.ParseFloat(settlementConfig.Mdr, 64)
			if err == nil {
				additionalFee = uint(float64(totalTransaction) * mdrFloat)
			} else {
				additionalFee = 0
			}
		}

		if strings.ToLower(settlementConfig.IsBhpuso) == "1" {
			bhpUSO = uint(float64(totalMerchant) * 0.0175)
		}

		if settlementConfig.AdditionalFee != nil && *settlementConfig.AdditionalFee == 1 {
			additionalFee = uint(float64(totalMerchant) * 0.05)
		}

		if settlementConfig.Tax23 != nil && strings.ToLower(*settlementConfig.Tax23) == "1" {
			tax23 = uint(float64(totalMerchant) * 0.02)
		}
	}

	var shareRedision, shareMerchant float32

	if settlementConfig.ShareRedision != nil {
		shareRedision = *settlementConfig.ShareRedision
	}

	if settlementConfig.SharePartner != nil {
		shareMerchant = *settlementConfig.SharePartner
	}

	result := ReportResult{
		AdditionalFee:          additionalFee,
		Summaries:              summaries,
		GrandTotalRedision:     grandTotalRedision,
		TotalMerchant:          totalMerchant,
		BhpUSO:                 bhpUSO,
		Mdr:                    settlementConfig.Mdr,
		Tax23:                  tax23,
		TotalTransactionAmount: uint(totalTransactionAmount),
		TotalTransaction:       totalTransaction,
		GrandTotal:             totalMerchant - bhpUSO - tax23 - additionalFee,
		ShareRedision:          shareRedision,
		ShareMerchant:          shareMerchant,
	}

	return c.JSON(result)
}

func GetReportMargin(c *fiber.Ctx) error {
	ctx := c.Context()

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	clientUID := c.Query("client_uid")
	appID := c.Query("app_id")

	var startDate, endDate *time.Time

	if startDateStr != "" {
		parsedStart, err := time.Parse(time.RFC1123, startDateStr)
		if err == nil {
			startDate = &parsedStart
		}
	}

	if endDateStr != "" {
		parsedEnd, err := time.Parse(time.RFC1123, endDateStr)
		if err == nil {
			endDate = &parsedEnd
		}
	}

	// log.Printf("Final parsed dates - startDate: %v, endDate: %v", startDate, endDate)
	// log.Printf("=== END DATE PARSING DEBUG ===")

	merchantNameStr := c.Query("merchant_name")
	var merchants []string
	if merchantNameStr != "" {
		merchants = strings.Split(merchantNameStr, ",")
	} else {
		merchants = []string{}
	}
	paymentMethods := c.Query("payment_method")

	summaries, err := repository.GetTransactionReportWithMargin(ctx, startDate, endDate, merchants, appID, clientUID, paymentMethods)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var (
		totalMargin        uint64
		totalShareRedision uint64
		totalTransaction   uint
		totalAmount        uint64
	)

	// Calculate totals
	for _, summary := range summaries {
		// log.Printf("Processing summary %d: Count=%d, TotalAmount=%d, TotalAmountTax=%d",
		// 	i+1, summary.Count, summary.TotalAmount, summary.TotalAmountTax)

		totalMargin += uint64(summary.Margin)
		totalShareRedision += uint64(summary.ShareRedision)
		totalTransaction += uint(summary.Count)
		totalAmount += summary.TotalAmount
	}

	result := fiber.Map{
		"summaries":    summaries,
		"total_amount": totalAmount,
		// "total_transaction":      totalTransaction,
		"total_margin": totalMargin,
		// "total_share_redision":   totalShareRedision,
		"calculation_formula":    "margin = shareRedision - (shareRedision * fee / 100)",
		"share_redision_formula": "shareRedision = amount - (amount * shareMerchantPercentage / 100)",
	}

	return c.JSON(result)
}
