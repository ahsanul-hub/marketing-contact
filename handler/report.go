package handler

import (
	"app/dto/model"
	"app/repository"
	"math"
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

	// 	if strings.ToLower(settlementConfig.IsBhpuso) == "1" {
	// 		additionalFee = uint(float64(totalMerchant) * 0.05)
	// 	}

	// 	if settlementConfig.Tax23 != nil && strings.ToLower(*settlementConfig.Tax23) == "1" {
	// 		tax23 = uint(float64(totalMerchant) * 0.02)
	// 	}
	// }

	if settlementConfig != nil {
		for i := range summaries {
			totalAmount := summaries[i].TotalAmount

			if settlementConfig.ShareRedision != nil {
				shareRedFloat := float64(totalAmount) * float64(*settlementConfig.ShareRedision) / 100
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

		if strings.ToLower(settlementConfig.IsBhpuso) == "1" {
			bhpUSO = uint(float64(totalMerchant) * 0.0175)
		}

		if strings.ToLower(settlementConfig.IsBhpuso) == "1" {
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
		Tax23:                  tax23,
		TotalTransactionAmount: uint(totalTransactionAmount),
		TotalTransaction:       totalTransaction,
		GrandTotal:             totalMerchant - bhpUSO - tax23 - additionalFee,
		ShareRedision:          shareRedision,
		ShareMerchant:          shareMerchant,
	}

	return c.JSON(result)
}
