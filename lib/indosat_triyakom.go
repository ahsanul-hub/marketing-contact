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
	"strings"
	"time"
)

func IndosatTriyakom(data model.InputPaymentRequest) (error, map[string]interface{}) {
	config, _ := config.GetGatewayConfig("indosat_triyakom")
	arrayOptions := config.Options["production"].(map[string]interface{})
	currentTime := time.Now()
	// arrayDenoms := config.Denom

	partnerID := "REDIS" //arrayOptions["partnerid"].(string)
	// itemID := arrayDenoms[strconv.FormatFloat(data.Amount, 'f', 0, 64)]
	cbParam := "6749a928109k5273128b7576" //data.TransactionID
	date := currentTime.Format("1/2/2006")
	secretKey := "DE9D7033E2584FCBBC479FFD654F44C7" //arrayOptions["seckey"].(string)
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
		"op":         "ISAT",
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
