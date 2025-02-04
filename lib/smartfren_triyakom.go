package lib

import (
	"app/config"
	"bytes"
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
	Price     int    `json:"price"`
}

type XimpayTransaction struct {
	ResponseCode int          `json:"responsecode"`
	XimpayID     string       `json:"ximpayid"`
	XimpayItem   []XimpayItem `json:"ximpayitem"`
}

type XimpayResponse struct {
	XimpayTransaction []XimpayTransaction `json:"ximpaytransaction"`
}

func RequestChargingSfTriyakom(msisdn, itemName, transactionId string, amount uint) (error, map[string]interface{}) {
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
		return fmt.Errorf("error marshalling body: %v", err), nil
	}

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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err), nil
	}

	// log.Println("resp:", string(body))

	defer resp.Body.Close()

	var response XimpayResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("error decoding response: %v", err), nil
	}

	responseCode := response.XimpayTransaction[0].ResponseCode

	if responseCode != 1 {
		log.Printf("error request charging with code: %d", responseCode)
		return fmt.Errorf("error request charging with code: %d", responseCode), nil
	}
	// switch responseCode {
	// case :
	// 	responseCode = v // Jika sudah string, langsung gunakan
	// case float64:
	// 	responseCode = fmt.Sprintf("%.0f", v) // Konversi float64 ke string
	// default:
	// 	return fmt.Errorf("unexpected type for responsecode: %T", v), nil // Tangani tipe yang tidak terduga
	// }

	return nil, map[string]interface{}{
		"success": true,
		"response": map[string]interface{}{
			"responsecode": responseCode,
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
