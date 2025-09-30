package helper

import (
	"app/dto/model"
	"log"
	"math"
	"strconv"
	"strings"
)

// FindSettlementByPaymentMethod mencari konfigurasi settlement berdasarkan nama payment method
func FindSettlementByPaymentMethod(settlementConfig *[]model.SettlementClient, paymentMethod string) *model.SettlementClient {
	if settlementConfig == nil {
		log.Printf("FindSettlementByPaymentMethod: paymentMethod=%s, configs=nil", paymentMethod)
		return nil
	}
	log.Printf("FindSettlementByPaymentMethod: paymentMethod=%s, total_configs=%d", paymentMethod, len(*settlementConfig))
	for i := range *settlementConfig {
		// log each name for debugging minimal
		name := (*settlementConfig)[i].Name
		if name == paymentMethod {
			log.Printf("FindSettlementByPaymentMethod: found match for paymentMethod=%s", paymentMethod)
			return &(*settlementConfig)[i]
		}
		if (*settlementConfig)[i].Name == paymentMethod {
			return &(*settlementConfig)[i]
		}
	}
	log.Printf("FindSettlementByPaymentMethod: NOT FOUND for paymentMethod=%s", paymentMethod)
	return nil
}

// ComputeFeeFromSettlement menghitung fee berdasarkan MDR dan MDR Type
// - mdr_type: "fix"/"fixed" -> nilai fix (dibulatkan ke atas)
// - selain itu -> dianggap persen, fee = ceil(price * mdr / 100)
func ComputeFeeFromSettlement(price uint, settlement *model.SettlementClient) uint {
	if settlement == nil {
		return 0
	}
	mdrStr := strings.TrimSpace(settlement.Mdr)
	if mdrStr == "" {
		return 0
	}
	mdrType := strings.ToLower(strings.TrimSpace(settlement.MdrType))
	if mdrType == "fix" || mdrType == "fixed" {
		if val, err := strconv.ParseFloat(mdrStr, 64); err == nil {
			return uint(math.Ceil(val))
		}
		return 0
	}
	if val, err := strconv.ParseFloat(mdrStr, 64); err == nil {
		return uint(math.Ceil(float64(price) * val / 100.0))
	}
	return 0
}
