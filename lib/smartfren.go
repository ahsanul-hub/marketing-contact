package lib

import (
	"app/config"
	"app/dto/model"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func SendPaymentSmartfren(data model.InputPaymentRequest, appKey string, bodySign string) (error, map[string]interface{}) {
	// configGateway, err := config
	config, _ := config.GetGatewayConfig("smartfren")
	amount := data.Amount

	// if !ok {
	// 	return fmt.Errorf("keyword for amount %d not found in config", amount), nil

	// }

	serviceNode := config.Options["serviceNode"].(string)
	// msisdn := data.UserMDNP
	msgCoding := config.Options["msgCoding"]
	sender := config.Options["sender"]
	smscId := config.Options["smscId"]
	bearerId := config.Options["bearerId"]

	smsCode := generateSMSCode()
	amountStr := strconv.FormatFloat(float64(amount), 'f', 0, 32)
	keyword := config.Denom[amountStr]
	hexMsg := keyword["keyword"] + " " + smsCode

	query := url.Values{}
	query.Add("serviceNode", serviceNode)
	// query.Add("msisdn", msisdn)
	query.Add("keyword", keyword["keyword"])
	query.Add("msgCoding", msgCoding.(string))
	query.Add("sender", sender.(string))
	query.Add("hexMsg", hexMsg)
	query.Add("smscId", smscId.(string))
	query.Add("bearerId", bearerId.(string))

	gateway := config.Options["ip"].(string) + ":" + config.Options["port"].(string) + "/moReq"

	requestURL := gateway + "?" + query.Encode()

	// Encode request data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %v", err), nil
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err), nil
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("appkey", appKey)
	req.Header.Set("bodysign", bodySign)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response status: %s", resp.Status), nil
	}

	responseData := map[string]interface{}{
		"requestURL": requestURL,
		"sms_code":   smsCode,
		"keyword":    keyword,
	}

	return nil, responseData

}

func SmartfrenTriyakomFlexible(data model.InputPaymentRequest) (error, map[string]interface{}) {
	config, _ := config.GetGatewayConfig("smartfren_triyakom_flex2")
	arrayOptions := config.Options["production"].(map[string]interface{})
	currentTime := time.Now()
	// arrayDenoms := config.Denom

	partnerID := "REDIS" //arrayOptions["partnerid"].(string)
	// itemID := arrayDenoms[strconv.FormatFloat(data.Amount, 'f', 0, 64)]
	cbParam := "6749a928109k5273128b7576" //data.TransactionID
	date := currentTime.Format("1/2/2006")
	secretKey := "DE9D7033E2584FCBBC479FFD654sF44C7" //arrayOptions["seckey"].(string)
	amount := data.Amount

	// Generate token sesuai dengan spesifikasi
	// Gabungkan parameter menjadi satu string
	// Ubah menjadi huruf kecil
	// lowerCaseString := "REDIS500012345611/25/2024DE9D7033E2584FCBBC479FFD654F44C7"
	// Enkripsi dengan MD5

	joinedString := fmt.Sprintf("%s%.0f%s%s%s", partnerID, amount, cbParam, date, secretKey)
	lowerCaseString := strings.ToLower(joinedString)
	token := fmt.Sprintf("%x", md5.Sum([]byte(lowerCaseString)))
	// log.Println("lowerCaseString: ", lowerCaseString)
	log.Println("token: ", token)

	arrBody := map[string]interface{}{
		"partnerid":  partnerID,
		"amount_exc": amount,
		"item_name":  "test item",
		"item_desc":  "item test desc",
		"cbparam":    cbParam,
		"token":      token,
		"op":         "SF",
		"msisdn":     data.UserMDN,
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

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response status: %s", resp.Status), nil
	}

	// Handle the response
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

	log.Println("responseCodeInterface", responseCodeInterface)
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
