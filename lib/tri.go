package lib

import (
	"app/config"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func RequestChargingTriTriyakom(msisdn, itemName, transactionId, amount string) (error, map[string]interface{}) {
	config, _ := config.GetGatewayConfig("tri")
	arrayOptions := config.Options["production"].(map[string]interface{})
	currentTime := time.Now()
	keyword := config.Denom[amount]

	partnerID := arrayOptions["partnerid"].(string)
	cbParam := fmt.Sprintf("r%s", transactionId)
	itemId := keyword["keyword"]
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
		return fmt.Errorf("error marshalling body: %v", err), nil
	}

	// Prepare the request
	requestURL := arrayOptions["requestUrl"].(string)
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err), nil
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err), nil
	}
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return fmt.Errorf("error reading response body: %v", err), nil
	// }

	// // Log the response body as a string
	// log.Println("resp:", string(body))

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response status: %s", resp.Status), nil
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("error decoding response: %v", err), nil
	}

	responseCodeInterface := response["ximpaytransaction"]

	// var responseCode string
	// switch v := responseCodeInterface.(type) {
	// case string:
	// 	responseCode = v // Jika sudah string, langsung gunakan
	// case float64:
	// 	responseCode = fmt.Sprintf("%.0f", v) // Konversi float64 ke string
	// default:
	// 	return fmt.Errorf("unexpected type for responsecode: %T", v), nil // Tangani tipe yang tidak terduga
	// }

	return nil, map[string]interface{}{
		"success": true,
		"response": map[string]interface{}{
			"responsecode": responseCodeInterface,
			"token":        arrBody,
			"code":         "00",
			"message":      "success",
			// "transaction_id": data.,
		},
		// "phone_number": data.UserMDN,
	}
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
