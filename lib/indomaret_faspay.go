package lib

import (
	"app/helper"
	"app/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func RequestIndomaretFaspay(transactionId, itemName, price, redirectUrl, customerName, UserMDN, PaymentChannel string) (*FaspayVaResponse, string, error) {

	location, _ := time.LoadLocation("Asia/Jakarta")
	tomorrow := time.Now().In(location).AddDate(0, 0, 1)
	itemDesc := fmt.Sprintf("item %s", price)

	signature, err := helper.GenerateFaspaySign("bot34184", "AGpzaek@", transactionId)
	if err != nil {
		log.Println("Error generate faspay dana sign")
	}

	now := time.Now().In(location)

	expiredDate := tomorrow.Format("2006-01-02 15:04:05")

	requestData := RequestVaFaspay{
		Request:        itemDesc,
		MerchantID:     "34184",
		Merchant:       "Redigame",
		BillNo:         transactionId,
		BillReff:       fmt.Sprintf("REF-%s", transactionId),
		BillDate:       now.Format("2006-01-02 15:04:05"),
		BillExpired:    expiredDate,
		BillDesc:       fmt.Sprintf("Payment Online Dana %s", itemDesc),
		BillCurrency:   "IDR",
		BillGross:      price,
		BillMiscFee:    "0",
		BillTotal:      price,
		CustNo:         UserMDN,
		CustName:       customerName,
		PaymentChannel: PaymentChannel,
		PayType:        "01",
		Msisdn:         UserMDN,
		Email:          "redision@gmail.com",
		Terminal:       "10",
		BillingCountry: "ID",
		ReceiverName:   customerName,
		ShippingState:  "Indonesia",
		Item: ItemRequestFaspay{
			ID:         transactionId,
			Product:    fmt.Sprintf("Invoice #%s", transactionId),
			Amount:     price,
			MerchantID: "99999",
		},
		Signature: signature,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		log.Println("Error marshaling request")
	}

	req, err := http.NewRequest("POST", "https://web.faspay.co.id/cvr/300011/10", bytes.NewReader(requestBody))
	if err != nil {
		log.Println("Error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request")
		return nil, "", fmt.Errorf("error charging dana: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Println("Error reading response")
	}

	helper.FaspayLogger.LogAPICall(
		"https://web.faspay.co.id/cvr/300011/10",
		"POST",
		time.Since(now),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id": transactionId,
			"request_body":   requestBody,
		},
		map[string]interface{}{
			"body": body,
		},
	)

	requestDate := &now
	err = repository.UpdateTransactionTimestamps(context.Background(), transactionId, requestDate, nil, nil)
	if err != nil {
		log.Printf("Error updating request timestamp for transaction %s: %s", transactionId, err)
	}

	var danaResponse FaspayVaResponse
	err = json.Unmarshal(body, &danaResponse)
	if err != nil {
		log.Println("Error decoding response")
	}

	return &danaResponse, expiredDate, nil
}
