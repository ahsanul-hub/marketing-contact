package lib

import (
	"app/config"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type MOResponseTsel struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func generateXSignature(apiKey, secret string) string {
	// Mendapatkan timestamp saat ini
	timeStamp := fmt.Sprintf("%d", time.Now().Unix())
	// Menghasilkan hash menggunakan MD5
	hash := md5.Sum([]byte(apiKey + secret + timeStamp))
	return fmt.Sprintf("%x", hash)
}

func RequestMoTsel(msisdn, itemID, itemDesc, transactionId string, denom string) (MOResponseTsel, string, int, error) {

	config, _ := config.GetGatewayConfig("telkomsel_airtime")
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
	sms := fmt.Sprintf("Waspada Penipuan! Anda akan membeli coin dengan tarif %s (termasuk ppn). Balas %s %d Abaikan jika tdk membeli. CS: bit.ly/3AqzjzU", price, keyword, otp)

	params := url.Values{}
	params.Add("adn", "99899")
	params.Add("sms", sms)
	params.Add("msisdn", msisdn)
	params.Add("trx_id", transactionId)
	params.Add("cp_name", "redision_game")
	params.Add("pwd", "Redision_Game")
	params.Add("sid", fmt.Sprintf("GAMGENRRG%s_IOD", keyword))

	// Menggabungkan parameter query menjadi string
	queryString := params.Encode()

	// Menggabungkan URL dengan parameter query
	fullURL := fmt.Sprintf("%s?%s", moUrl, queryString)

	// Membuat request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return MOResponseTsel{}, "", 0, fmt.Errorf("error creating request: %v", err)
	}

	// Menambahkan header
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

	// Decode response body
	var response MOResponseTsel

	return response, keyword, otp, nil

}

func RequestMtTsel(msisdn, itemID, itemDesc, transactionId string, denom string) (MOResponseTsel, error) {

	config, _ := config.GetGatewayConfig("telkomsel_airtime")
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
	price := denomConfig["price"]
	sid := denomConfig["sid"]

	var otp int
	for {
		otp = rand.Intn(100000)
		if otp >= 10000 {
			break
		}
	}
	sms := fmt.Sprintf("Terima kasih. Anda telah melakukan pembelian coin seharga Rp. %s (incl.PPN) dari PT. Redision Teknologi Indonesia. CS: bit.ly/3AqzjzU", price)

	params := url.Values{}
	params.Add("adn", "99899")
	params.Add("sms", sms)
	params.Add("msisdn", msisdn)
	params.Add("trx_id", transactionId)
	params.Add("cpid", "redision_game")
	params.Add("pwd", "Redision_Game")
	params.Add("sid", sid)

	// Menggabungkan parameter query menjadi string
	queryString := params.Encode()

	// Menggabungkan URL dengan parameter query
	fullURL := fmt.Sprintf("%s?%s", mtUrl, queryString)

	// Membuat request
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return MOResponseTsel{}, fmt.Errorf("error creating request: %v", err)
	}

	// Menambahkan header
	req.Header.Add("api_key", apiKey)
	req.Header.Add("x-signature", signature)

	// Membuat client HTTP
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

	if resp.StatusCode != http.StatusOK {
		return MOResponseTsel{}, fmt.Errorf("HTTP error: %s, Response body: %s", resp.Status, body)
	}

	// Decode response body
	var response MOResponseTsel

	return response, nil

}
