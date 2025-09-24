package lib

import (
	"app/helper"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type RequestVaFaspay struct {
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

type FaspayVaResponse struct {
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

func RequestChargingVaFaspay(transactionId, itemName, price, redirectUrl, customerName, UserMDN, PaymentChannel string) (*FaspayVaResponse, string, error) {

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

	var FaspayResponse FaspayVaResponse
	err = json.Unmarshal(body, &FaspayResponse)
	if err != nil {
		log.Println("Error decoding response")
	}

	return &FaspayResponse, expiredDate, nil
}
