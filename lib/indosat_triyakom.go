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

func RequestChargingIsatTriyakom(msisdn, itemName, transactionId string, chargingPrice uint) (string, error) {
	config, _ := config.GetGatewayConfig("indosat_triyakom")
	arrayOptions := config.Options["production"].(map[string]interface{})
	currentTime := time.Now()

	partnerID := arrayOptions["partnerid"].(string)
	cbParam := fmt.Sprintf("r%s", transactionId)
	date := currentTime.Format("1/2/2006")
	secretKey := arrayOptions["seckey"].(string)
	amount := chargingPrice

	joinedString := fmt.Sprintf("%s%d%s%s%s", partnerID, amount, cbParam, date, secretKey)
	lowerCaseString := strings.ToLower(joinedString)
	token := fmt.Sprintf("%x", md5.Sum([]byte(lowerCaseString)))
	log.Println("lowerCaseString: ", lowerCaseString)
	log.Println("token: ", token)

	arrBody := map[string]interface{}{
		"partnerid":   strings.ToLower(partnerID),
		"amount":      amount,
		"item_name":   itemName,
		"charge_type": "ISAT_GENERAL",
		"item_desc":   itemName,
		"cbparam":     cbParam,
		"token":       token,
		"op":          "ISAT",
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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response status: %s", resp.Status)
	}

	// var response map[string]interface{}
	// if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
	// 	return fmt.Errorf("error decoding response: %v", err), nil
	// }

	// responseCodeInterface := response["ximpaytransaction"]
	// ximpayID := response.XimpayTransaction[0].XimpayID

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
	// if responseCode == "1" {
	// 	// Handle success
	// 	return nil, map[string]interface{}{
	// 		"success": true,
	// 		"response": map[string]interface{}{
	// 			"responsecode": responseCode,
	// 			"code":         "00",
	// 			"message":      "success",
	// 			// "transaction_id": data.,
	// 		},
	// 		"phone_number": data.UserMDN,
	// 	}
	// } else {
	// 	// Handle failure
	// 	return fmt.Errorf("transaction failed with response code: %s", responseCodeInterface), nil
	// }
}
