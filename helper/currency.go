package helper

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatCurrencyIDR(amount uint) string {
	amountStr := strconv.FormatUint(uint64(amount), 10)

	if len(amountStr) <= 3 {
		return amountStr
	}

	reversed := reverseString(amountStr)
	var result strings.Builder

	for i, char := range reversed {
		if i > 0 && i%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(char)
	}

	// Balik kembali hasilnya
	return reverseString(result.String())
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ValidateCurrency(currency string) (string, error) {
	if currency == "" {
		return "IDR", nil
	}

	validCurrencies := []string{"IDR", "USD", "PHP"}

	normalized := strings.ToUpper(currency)
	for _, valid := range validCurrencies {
		if normalized == valid {
			return valid, nil
		}
	}

	return "", fmt.Errorf("unsupported currency: %s", currency)
}
