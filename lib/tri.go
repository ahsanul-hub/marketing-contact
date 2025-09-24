package lib

import (
	"app/config"
	"app/helper"
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

func RequestChargingTriTriyakom(msisdn, itemName, transactionId, amount string) (string, error) {
	config, _ := config.GetGatewayConfig("tri")
	arrayOptions := config.Options["production"].(map[string]interface{})
	keyword := config.Denom[amount]

	partnerID := arrayOptions["partnerid"].(string)
	cbParam := fmt.Sprintf("rt%s", transactionId)
	itemId := keyword["keyword"]

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return "", fmt.Errorf("failed to load timezone: %v", err)
	}
	currentTime := time.Now().In(loc)

	date := currentTime.Format("1/2/2006")
	secretKey := arrayOptions["seckey"].(string)

	joinedString := fmt.Sprintf("%s%s%s%s%s", partnerID, itemId, cbParam, date, secretKey)
	lowerCaseString := strings.ToLower(joinedString)
	token := fmt.Sprintf("%x", md5.Sum([]byte(lowerCaseString)))

	arrBody := map[string]interface{}{
		"partnerid":   strings.ToLower(partnerID),
		"charge_type": "HTI_GENERAL",
		"itemid":      itemId,
		"item_desc":   itemName,
		"cbparam":     cbParam,
		"token":       token,
		"op":          "HTI",
		"msisdn":      msisdn,
	}

	jsonBody, err := json.Marshal(arrBody)
	if err != nil {
		return "", fmt.Errorf("error marshalling body: %v", err)
	}

	// Prepare the request
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

	helper.TriLogger.LogAPICall(
		requestURL,
		"POST",
		time.Since(currentTime),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id": transactionId,
			"request_body":   jsonBody,
		},
		map[string]interface{}{
			"body": string(body),
		},
	)

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

	return ximpayID, nil

}
