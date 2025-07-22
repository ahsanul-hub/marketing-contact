package helper

import (
	"errors"
	"strings"
)

// Daftar currency yang diperbolehkan
var allowedCurrencies = map[string]bool{
	"IDR": true,
	"USD": true,
	"PHP": true,
}

func ValidateCurrency(input string) (string, error) {
	currency := strings.ToUpper(strings.TrimSpace(input))
	if currency == "" {
		return "IDR", nil
	}

	if !allowedCurrencies[currency] {
		return "", errors.New("invalid currency. Allowed currencies: IDR, USD, PHP")
	}

	return currency, nil
}
