package lib

import (
	"app/config"
	"app/repository"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type XimpayItem struct {
	ItemName  string `json:"item_name"`
	ItemDesc  string `json:"item_desc"`
	AmountExc int    `json:"amount_exc"`
	Name      string `json:"name"`
	Price     string `json:"price"`
}

type XimpayTransaction struct {
	ResponseCode int          `json:"responsecode"`
	XimpayID     string       `json:"ximpayid"`
	XimpayItem   []XimpayItem `json:"ximpayitem"`
}

type XimpayResponse struct {
	XimpayTransaction []XimpayTransaction `json:"ximpaytransaction"`
}

type DoMTRequest struct {
	XimpayID    string `json:"ximpayid"`
	CodePin     string `json:"codepin"`
	XimpayToken string `json:"ximpaytoken"`
}

type DoMTResponse struct {
	XimpayTransaction []struct {
		ResponseCode int    `json:"responsecode"`
		XimpayID     string `json:"ximpayid"`
		Pin          string `json:"pin"`
	} `json:"ximpaytransaction"`
}

func RequestChargingSfTriyakom(msisdn, itemName, transactionId string, amount uint) (string, error) {
	config, _ := config.GetGatewayConfig("smartfren_triyakom")
	arrayOptions := config.Options["production"].(map[string]interface{})
	currentTime := time.Now()

	partnerID := arrayOptions["partnerid"].(string)
	cbParam := fmt.Sprintf("r%s", transactionId)
	date := currentTime.Format("1/2/2006")
	secretKey := arrayOptions["seckey"].(string)

	joinedString := fmt.Sprintf("%s%d%s%s%s", partnerID, amount, cbParam, date, secretKey)
	lowerCaseString := strings.ToLower(joinedString)
	token := fmt.Sprintf("%x", md5.Sum([]byte(lowerCaseString)))

	arrBody := map[string]interface{}{
		"partnerid":  strings.ToLower(partnerID),
		"amount_exc": amount,
		"item_name":  itemName,
		"item_desc":  itemName,
		"cbparam":    cbParam,
		"token":      token,
		"op":         "SF",
		"msisdn":     msisdn,
	}

	jsonBody, err := json.Marshal(arrBody)
	if err != nil {
		return "", fmt.Errorf("error marshalling body: %v", err)
	}

	requestURL := arrayOptions["requestUrl"].(string)
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// log.Println("resp:", string(body))

	defer resp.Body.Close()

	var response XimpayResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	responseCode := response.XimpayTransaction[0].ResponseCode
	ximpayID := response.XimpayTransaction[0].XimpayID

	if responseCode != 1 {
		log.Printf("error request charging with code: %d", responseCode)
		return "", fmt.Errorf("error request charging with code: %d", responseCode)
	}

	now := time.Now()

	requestDate := &now

	err = repository.UpdateTransactionTimestamps(context.Background(), transactionId, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionId, err)
	}

	// switch responseCode {
	// case :
	// 	responseCode = v // Jika sudah string, langsung gunakan
	// case float64:
	// 	responseCode = fmt.Sprintf("%.0f", v) // Konversi float64 ke string
	// default:
	// 	return fmt.Errorf("unexpected type for responsecode: %T", v), nil // Tangani tipe yang tidak terduga
	// }

	return ximpayID, nil
}

func DoMT(ximpayID, codePin string) error {
	config, _ := config.GetGatewayConfig("smartfren_triyakom")
	arrayOptions := config.Options["production"].(map[string]interface{})
	secretKey := arrayOptions["seckey"].(string)

	joinedString := fmt.Sprintf("%s%s%s", ximpayID, codePin, secretKey)
	lowerCaseString := strings.ToLower(joinedString)
	ximpayToken := fmt.Sprintf("%x", md5.Sum([]byte(lowerCaseString)))

	requestBody := DoMTRequest{
		XimpayID:    ximpayID,
		CodePin:     codePin,
		XimpayToken: ximpayToken,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error marshalling request body: %v", err)
	}

	requestURL := arrayOptions["confirmUrl"].(string)
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}

	var response DoMTResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	if len(response.XimpayTransaction) == 0 {
		return fmt.Errorf("unexpected empty response")
	}

	responseCode := response.XimpayTransaction[0].ResponseCode
	if responseCode != 1 {
		log.Printf("error request DoMT with code: %d", responseCode)
		return fmt.Errorf("error request DoMT with code: %d", responseCode)
	}

	return nil
}
