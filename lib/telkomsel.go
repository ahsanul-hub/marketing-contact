package lib

import (
	"app/config"
	"app/database"
	"app/helper"
	"app/repository"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type MOResponseTsel struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func generateXSignature(apiKey, secret string) string {
	// Mendapatkan timestamp saat ini
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	// Menghasilkan hash menggunakan MD5
	hash := md5.Sum([]byte(apiKey + secret + timeStamp))
	return fmt.Sprintf("%x", hash)
}

func RequestMoTsel(msisdn, itemID, itemDesc, transactionId string, denom string) (MOResponseTsel, string, int, error) {

	config, _ := config.GetGatewayConfig("telkomsel_airtime_sms")
	arrayOptions := config.Options
	moUrl := arrayOptions["moUrl"].(string)
	apiKey := arrayOptions["apikey"].(string)
	secret := arrayOptions["secret"].(string)

	signature := generateXSignature(apiKey, secret)

	denomConfig, exists := config.Denom[denom]
	if !exists {
		return MOResponseTsel{}, "", 0, fmt.Errorf("denom %s not found in gateway config", denom)
	}

	keyword := denomConfig["keyword"]
	price := denomConfig["price"]

	var otp int
	for {
		otp = rand.Intn(100000)
		if otp >= 10000 {
			break
		}
	}
	sms := fmt.Sprintf("Waspada Penipuan! Anda akan membeli coin dengan tarif %s (termasuk ppn). Balas \n \n %s %d \n \n Abaikan jika tdk membeli. CS: bit.ly/3AqzjzU", price, keyword, otp)

	params := url.Values{}
	params.Add("adn", "99899")
	params.Add("sms", sms)
	params.Add("msisdn", msisdn)
	params.Add("trx_id", transactionId)
	params.Add("cp_name", "redision_game")
	params.Add("pwd", "Redision_Game")
	params.Add("sid", "GAMGENRRGREDS350_IOD")

	queryString := params.Encode()

	fullURL := fmt.Sprintf("%s?%s", moUrl, queryString)

	// log.Println("fullURL:", fullURL)
	// log.Println("key", apiKey, signature)
	// Membuat request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return MOResponseTsel{}, "", 0, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("api_key", apiKey)
	req.Header.Add("x-signature", signature)

	// Membuat client HTTP
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return MOResponseTsel{}, "", 0, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return MOResponseTsel{}, "", 0, fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return MOResponseTsel{}, "", 0, fmt.Errorf("HTTP error: %s, Response body: %s", resp.Status, body)
	}
	now := time.Now()

	requestDate := &now

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionId, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionId, err)
	}

	// Insert ke Redis

	zeroMsisdn := helper.BeautifyIDNumber(msisdn, true)

	ctx := context.Background()
	cacheKey := fmt.Sprintf("tsel:tx:%s:%s:%d", zeroMsisdn, keyword, otp)

	cacheData := map[string]interface{}{
		"transaction_id": transactionId,
		"msisdn":         msisdn,
		"keyword":        keyword,
		"amount":         denom,
		"otp":            otp,
		"created_at":     now.Unix(),
	}
	jsonData, _ := json.Marshal(cacheData)

	log.Println("cacheKey request mo tsel", cacheKey)

	if database.RedisClient != nil {
		if err := database.RedisClient.Set(ctx, cacheKey, jsonData, 10*time.Minute).Err(); err != nil {
			log.Printf("Error saving transaction %s to Redis: %s", transactionId, err)
		}
	}

	// Decode response body
	var response MOResponseTsel

	return response, keyword, otp, nil

}

func RequestMtTsel(msisdn, transactionId string, denom string) (MOResponseTsel, error) {

	config, _ := config.GetGatewayConfig("telkomsel_airtime_sms")
	arrayOptions := config.Options
	mtUrl := arrayOptions["mtUrl"].(string)
	apiKey := arrayOptions["apikey"].(string)
	secret := arrayOptions["secret"].(string)

	signature := generateXSignature(apiKey, secret)

	denomConfig, exists := config.Denom[denom]
	if !exists {
		return MOResponseTsel{}, fmt.Errorf("denom %s not found in gateway config", denom)
	}

	// keyword := denomConfig["keyword"]

	var response MOResponseTsel

	price := denomConfig["price"]
	sid := denomConfig["sid"]
	tid := denomConfig["tid"]

	var otp int
	for {
		otp = rand.Intn(100000)
		if otp >= 10000 {
			break
		}
	}
	sms := fmt.Sprintf("Terima kasih. Anda telah melakukan pembelian coin seharga Rp. %s (incl.PPN) dari PT. Redision Teknologi Indonesia. CS: bit.ly/3AqzjzU", price)

	params := url.Values{}
	params.Add("sender", "99899")
	params.Add("cpid", "redision_game")
	params.Add("pwd", "R3dision!")
	params.Add("msisdn", msisdn)
	params.Add("sid", sid)
	params.Add("tid", tid)
	params.Add("sms", sms)
	params.Add("trx_id", transactionId)

	queryString := params.Encode()

	fullURL := fmt.Sprintf("%s?%s", mtUrl, queryString)

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return MOResponseTsel{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("api_key", apiKey)
	req.Header.Add("x-signature", signature)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return MOResponseTsel{}, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return MOResponseTsel{}, fmt.Errorf("error reading response body: %v", err)
	}

	// log.Println("res mt Tsel", string(body))

	resBody := string(body)

	// if resBody != "1" && resp.StatusCode != 202 {
	// 	return MOResponseTsel{}, fmt.Errorf("HTTP error: %s, Response body: %s", resp.Status, body)
	// }

	response = MOResponseTsel{
		Status:     resBody,
		StatusCode: resp.StatusCode,
		Message:    "success",
	}

	return response, nil

}
