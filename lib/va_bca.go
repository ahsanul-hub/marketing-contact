package lib

import (
	"app/config"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
)

type ResGenerateVa struct {
	VaNumber    string `json:"va_number"`
	ExpiredTime string `json:"expired_time"`
}

type VaRedpayTokenRequest struct {
	GrantType string `json:"grant_type"`
}

type RedpayVaTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	Scope       string `json:"scope"`
}

type CachedToken struct {
	Token     string
	ExpiredAt time.Time
}

var RedpayTokenCache = cache.New(5*time.Minute, 10*time.Minute)

func GenerateVA() (ResGenerateVa, error) {

	config, _ := config.GetGatewayConfig("va_bca_direct")
	arrayOptions := config.Options["production"].(map[string]interface{})

	prefix := arrayOptions["prefix"].(string)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	vaNumber := prefix + GenerateRandomString(10)
	expiredTime := now.Add(1 * time.Hour).Format("2006-01-02 15:04:05")

	response := ResGenerateVa{
		VaNumber:    vaNumber,
		ExpiredTime: expiredTime,
	}

	return response, nil
}

func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	numbers := "0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = numbers[rand.Intn(len(numbers))]
	}
	return string(result)
}

func RequestTokenVaBCARedpay() (*RedpayVaTokenResp, error) {
	// config, _ := config.GetGatewayConfig("xl_twt")
	// arrayOptions := config.Options["development"].(map[string]interface{})

	requestBody := VaRedpayTokenRequest{
		GrantType: "client_credentials",
	}

	url := "https://payment.redpay.co.id/api/v1/bca/token"
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic UjNkMXMxMG46YXRkc1Vxcml3MTQxQVQzTDlQNFo=")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s", resp.Status)
	}

	var tokenResp RedpayVaTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	expiresInSec, err := strconv.Atoi(tokenResp.ExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("invalid expires_in format: %w", err)
	}

	expiredAt := time.Now().Add(time.Duration(expiresInSec) * time.Second)

	cached := CachedToken{
		Token:     tokenResp.AccessToken,
		ExpiredAt: expiredAt,
	}

	RedpayTokenCache.Set("redpay_token", cached, time.Duration(expiresInSec)*time.Second)

	return &tokenResp, nil
}
