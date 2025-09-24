package lib

import (
	"app/dto/model"
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

type RequestDanaFaspay struct {
	Request         string            `json:"request"`
	MerchantID      string            `json:"merchant_id"`
	Merchant        string            `json:"merchant"`
	BillNo          string            `json:"bill_no"`
	BillReff        string            `json:"bill_reff,omitempty"`
	BillDate        string            `json:"bill_date"`
	BillExpired     string            `json:"bill_expired"`
	BillDesc        string            `json:"bill_desc"`
	BillCurrency    string            `json:"bill_currency"`
	BillGross       string            `json:"bill_gross,omitempty"`
	BillMiscFee     string            `json:"bill_miscfee,omitempty"`
	BillTotal       string            `json:"bill_total"`
	CustNo          string            `json:"cust_no"`
	CustName        string            `json:"cust_name"`
	PaymentChannel  string            `json:"payment_channel"`
	PayType         string            `json:"pay_type"`
	Msisdn          string            `json:"msisdn"`
	Email           string            `json:"email"`
	Terminal        string            `json:"terminal"`
	BillingAddress  string            `json:"billing_address,omitempty"`
	BillingCity     string            `json:"billing_address_city,omitempty"`
	BillingRegion   string            `json:"billing_address_region,omitempty"`
	BillingState    string            `json:"billing_address_state,omitempty"`
	BillingPoscode  string            `json:"billing_address_poscode,omitempty"`
	BillingCountry  string            `json:"billing_address_country_code,omitempty"`
	ReceiverName    string            `json:"receiver_name_for_shipping,omitempty"`
	ShippingAddress string            `json:"shipping_address,omitempty"`
	ShippingCity    string            `json:"shipping_address_city,omitempty"`
	ShippingRegion  string            `json:"shipping_address_region,omitempty"`
	ShippingState   string            `json:"shipping_address_state,omitempty"`
	ShippingPoscode string            `json:"shipping_address_poscode,omitempty"`
	Item            ItemRequestFaspay `json:"item"`
	Reserve1        string            `json:"reserve1,omitempty"`
	Reserve2        string            `json:"reserve2,omitempty"`
	Signature       string            `json:"signature"`
}

type ItemRequestFaspay struct {
	ID          string `json:"id"`
	Product     string `json:"product,omitempty"`
	Qty         string `json:"qty,omitempty"`
	Amount      string `json:"amount"`
	PaymentPlan string `json:"payment_plan,omitempty"`
	MerchantID  string `json:"merchant_id"`
	Tenor       string `json:"tenor,omitempty"`
}

type FaspayDanaResponse struct {
	Response     string           `json:"response"`
	TrxID        string           `json:"trx_id"`
	MerchantID   string           `json:"merchant_id"`
	Merchant     string           `json:"merchant"`
	BillNo       string           `json:"bill_no"`
	ExternalID   string           `json:"external_id"`
	BillItems    []FaspayBillItem `json:"bill_items"`
	ResponseCode string           `json:"response_code"`
	ResponseDesc string           `json:"response_desc"`
	RedirectURL  string           `json:"redirect_url"`
}

type FaspayBillItem struct {
	ID          string `json:"id"`
	Product     string `json:"product"`
	Qty         string `json:"qty"`
	Amount      string `json:"amount"`
	PaymentPlan string `json:"payment_plan"`
	MerchantID  string `json:"merchant_id"`
	Tenor       string `json:"tenor"`
}

type RequestDanaCheckStatus struct {
	Request    string `json:"request"`
	TrxID      string `json:"trx_id"`
	MerchantID string `json:"merchant_id"`
	BillNo     string `json:"bill_no"`
	Signature  string `json:"signature"`
}

type DanaFaspayQueryResponse struct {
	Response          string `json:"response"`
	TrxID             string `json:"trx_id"`
	MerchantID        string `json:"merchant_id"`
	Merchant          string `json:"merchant"`
	BillNo            string `json:"bill_no"`
	PaymentReff       string `json:"payment_reff"`
	PaymentDate       string `json:"payment_date"`
	PaymentStatusCode string `json:"payment_status_code"`
	PaymentStatusDesc string `json:"payment_status_desc"`
	ResponseCode      string `json:"response_code"`
	ResponseDesc      string `json:"response_desc"`
}

func RequestChargingDanaFaspay(transactionId, itemName, price, redirectUrl, customerName, UserMDN string) (*FaspayDanaResponse, error) {

	location, _ := time.LoadLocation("Asia/Jakarta")
	tomorrow := time.Now().In(location).AddDate(0, 0, 1)
	itemDesc := fmt.Sprintf("item %s", price)

	signature, err := helper.GenerateFaspaySign("bot34184", "AGpzaek@", transactionId)
	if err != nil {
		log.Println("Error generate faspay dana sign")
	}

	now := time.Now().In(location)
	requestData := RequestDanaFaspay{
		Request:        itemDesc,
		MerchantID:     "34184",
		Merchant:       "Redigame",
		BillNo:         transactionId,
		BillReff:       fmt.Sprintf("REF-%s", transactionId),
		BillDate:       now.Format("2006-01-02 15:04:05"),
		BillExpired:    tomorrow.Format("2006-01-02 15:04:05"),
		BillDesc:       fmt.Sprintf("Payment Online Dana %s", itemDesc),
		BillCurrency:   "IDR",
		BillGross:      price,
		BillMiscFee:    "0",
		BillTotal:      price,
		CustNo:         UserMDN,
		CustName:       customerName,
		PaymentChannel: "819",
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
		return nil, fmt.Errorf("error charging dana: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Println("Error reading response")
	}

	helper.DanaLogger.LogAPICall(
		"https://web.faspay.co.id/cvr/300011/10",
		"POST",
		time.Since(now),
		resp.StatusCode,
		map[string]interface{}{
			"transaction_id": transactionId,
			"request_body":   requestBody,
		},
		map[string]interface{}{
			"body": string(body),
		},
	)

	var danaResponse FaspayDanaResponse
	err = json.Unmarshal(body, &danaResponse)
	if err != nil {
		log.Println("Error decoding response")
	}

	return &danaResponse, nil
}

func CheckOrderDanaFaspay(transactionId, referenceID string) (*DanaFaspayQueryResponse, error) {
	if referenceID == "" {
		return nil, fmt.Errorf("referenceID kosong, tidak bisa cek status Dana Faspay")
	}

	signature, err := helper.GenerateFaspaySign("bot34184", "AGpzaek@", transactionId)
	if err != nil {
		log.Println("Error generate faspay dana sign:", err)
		return nil, err
	}

	requestData := RequestDanaCheckStatus{
		Request:    "Inquiry Payment Status",
		TrxID:      referenceID,
		MerchantID: "34184",
		BillNo:     transactionId,
		Signature:  signature,
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		log.Println("Error marshaling request:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://web.faspay.co.id/cvr/100004/10", bytes.NewReader(requestBody))
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return nil, fmt.Errorf("error check status dana: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response:", err)
		return nil, err
	}

	var resCheckStatus DanaFaspayQueryResponse
	err = json.Unmarshal(body, &resCheckStatus)
	if err != nil {
		log.Println("Error decoding response:", err)
		return nil, err
	}

	return &resCheckStatus, nil
}

func CheckTransactionStatusDanaFaspay(transaction model.Transactions) {

	res, err := CheckOrderDanaFaspay(transaction.ID, transaction.ReferenceID)
	if err != nil {
		// log.Printf("error check order dana faspay: %s", err.Error())
		return
	}
	now := time.Now()

	receiveCallbackDate := &now

	if res.PaymentStatusCode == "2" {

		if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1003, nil, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transaction.ID, err)
		}
	} else {
		switch res.PaymentStatusCode {
		case "4":
			if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, nil, nil, "Payment Reserval ", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", transaction.ID, err)
			}
		case "5":
			if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, nil, nil, "No bills found ", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", transaction.ID, err)
			}
		case "8":
			if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, nil, nil, "Payment Cancelled", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", transaction.ID, err)
			}
		case "9":
			if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, nil, nil, "Unknown", receiveCallbackDate); err != nil {
				log.Printf("Error updating transaction status for %s: %s", transaction.ID, err)
			}
		default:
			createdAt := transaction.CreatedAt
			timeLimit := time.Now().Add(-10 * time.Minute)

			expired := createdAt.Before(timeLimit)
			if expired {
				if err := repository.UpdateTransactionStatusExpired(context.Background(), transaction.ID, 1005, "", ""); err != nil {
					log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
				}
			}
		}

	}

}

// func worker(jobs <-chan model.Transactions) {
// 	for tx := range jobs {
// 		CheckTransactionStatusDanaFaspay(tx)
// 	}
// }

// func ProcessPendingDanaFaspayTransactions() {
// 	const workerCount = 10
// 	jobs := make(chan model.Transactions, 100)

// 	// Worker
// 	for i := 0; i < workerCount; i++ {
// 		go worker(jobs)
// 	}

// 	// Producer
// 	go func() {
// 		for {
// 			transactions, err := repository.GetPendingTransactions(context.Background(), "dana")
// 			if err != nil {
// 				log.Printf("Error retrieving pending transactions: %s", err)
// 				time.Sleep(1 * time.Minute)
// 				continue
// 			}

// 			for _, tx := range transactions {
// 				jobs <- tx
// 			}

// 			time.Sleep(5 * time.Second)
// 		}
// 	}()
// }
