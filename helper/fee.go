package helper

import (
	"app/dto/model"
	"log"
	"math"
	"strconv"
)

func CalculateFee(amount uint, payment_method string, config *model.SettlementClient) (fee uint, err error) {
	if config == nil {
		log.Printf("CalculateFee: config is nil for payment_method: %s", payment_method)
		return 0, nil
	}

	if config.ShareRedision == nil {
		log.Printf("CalculateFee: ShareRedision is nil for payment_method: %s", payment_method)
		return 0, nil
	}

	if config.MdrType == "fix" && *config.ShareRedision == 0 {
		mdrFloat, parseErr := strconv.ParseFloat(config.Mdr, 64)
		if parseErr != nil {
			log.Printf("CalculateFee: failed to parse MDR '%s' for payment_method: %s, error: %v", config.Mdr, payment_method, parseErr)
			return 0, parseErr
		}
		fee = uint(math.Round(mdrFloat))
	} else if config.MdrType == "fix" && *config.ShareRedision > 0 {
		mdrFloat, parseErr := strconv.ParseFloat(config.Mdr, 64)
		if parseErr != nil {
			log.Printf("CalculateFee: failed to parse MDR '%s' for payment_method: %s, error: %v", config.Mdr, payment_method, parseErr)
			return 0, parseErr
		}
		fee = uint(math.Round(mdrFloat))
		shareRedisionFee := uint(math.Ceil(float64(amount) * float64(*config.ShareRedision) / 100))

		fee += shareRedisionFee
	} else {
		fee = uint(math.Ceil(float64(amount) * float64(*config.ShareRedision) / 100.0))
	}

	return fee, nil
}
