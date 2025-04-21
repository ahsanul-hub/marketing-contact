package lib

import (
	"app/config"
	"app/helper"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	"time"
)

type CreateOrderPayload struct {
	Request   RequestData `json:"request"`
	Signature string      `json:"signature"`
}

type RequestData struct {
	Head HeadData `json:"head"`
	Body BodyData `json:"body"`
}

type HeadData struct {
	Version      string `json:"version"`
	Function     string `json:"function"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	ReqTime      string `json:"reqTime"`
	ReqMsgId     string `json:"reqMsgId"`
	Reserve      string `json:"reserve"`
}

type BodyData struct {
	EnvInfo     EnvInfo   `json:"envInfo"`
	Order       OrderData `json:"order"`
	ProductCode string    `json:"productCode"`
	MCC         string    `json:"mcc"`
	MerchantID  string    `json:"merchantId"`
	// ExtendInfo        string      `json:"extendInfo"`
	PaymentPreference PaymentPref `json:"paymentPreference"`
	NotificationUrls  []NotifyURL `json:"notificationUrls"`
}

type EnvInfo struct {
	TerminalType string `json:"terminalType"`
	// OsType            string `json:"osType"`
	// ExtendInfo        string `json:"extendInfo"`
	// OrderOsType       string `json:"orderOsType"`
	// SdkVersion        string `json:"sdkVersion"`
	OrderTerminalType string `json:"orderTerminalType"`
	SourcePlatform    string `json:"sourcePlatform"`
	// ClientIp          string `json:"clientIp"`
	// ClientKey         string `json:"clientKey"`
}

type OrderData struct {
	ExpiryTime        string        `json:"expiryTime"`
	MerchantTransType string        `json:"merchantTransType"`
	OrderTitle        string        `json:"orderTitle"`
	MerchantTransId   string        `json:"merchantTransId"`
	OrderMemo         string        `json:"orderMemo"`
	CreatedTime       string        `json:"createdTime"`
	OrderAmount       Amount        `json:"orderAmount"`
	Goods             []GoodsDetail `json:"goods"`
}

type Amount struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

type GoodsDetail struct {
	Description string `json:"description"`
	Price       Amount `json:"price"`
}

type PaymentPref struct {
	SupportDeepLinkCheckoutUrl bool `json:"supportDeepLinkCheckoutUrl"`
}

type NotifyURL struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type DanaResponse struct {
	Response struct {
		Head struct {
			Function string `json:"function"`
			ClientID string `json:"clientId"`
			Version  string `json:"version"`
			RespTime string `json:"respTime"`
			ReqMsgId string `json:"reqMsgId"`
		} `json:"head"`
		Body struct {
			MerchantTransId string `json:"merchantTransId,omitempty"`
			AcquirementId   string `json:"acquirementId,omitempty"`
			CheckoutUrl     string `json:"checkoutUrl,omitempty"`
			ResultInfo      struct {
				ResultStatus string `json:"resultStatus"`
				ResultCodeId string `json:"resultCodeId"`
				ResultMsg    string `json:"resultMsg"`
				ResultCode   string `json:"resultCode"`
			} `json:"resultInfo"`
		} `json:"body"`
	} `json:"response"`
	Signature string `json:"signature"`
}

func RequestChargingDana(transactionId, itemName, price string) (string, error) {

	loc := time.FixedZone("IST", 5*60*60+30*60) // GMT+5:30
	reqTime := time.Now().In(loc).Format("2006-01-02T15:04:05-07:00")
	location, _ := time.LoadLocation("Asia/Jakarta")
	tomorrow := time.Now().In(location).AddDate(0, 0, 1)
	formattedTomorrow := tomorrow.Format("2006-01-02T15:04:05-07:00")

	notifyUrl := fmt.Sprintf("%s/callback/dana", config.Config("APIURL", ""))
	returnUrl := fmt.Sprintf("%s/return/dana", config.Config("APIURL", ""))
	clientId := "2023060711065517686870"
	clientSecret := "dd4592b541c0c1e2530c044efdf1eb412d94ea6071e9ccead1cfbf1616269d17"
	merchantId := "216620060007007966853"
	// log.Println("price", price)

	requestData := RequestData{
		Head: HeadData{
			Version:      "2.0",
			Function:     "dana.acquiring.order.createOrder",
			ClientID:     clientId,
			ClientSecret: clientSecret,
			ReqTime:      reqTime,
			ReqMsgId:     transactionId,
			Reserve:      "{}",
		},
		Body: BodyData{
			EnvInfo: EnvInfo{
				TerminalType:      "SYSTEM",
				OrderTerminalType: "SYSTEM",
				SourcePlatform:    "IPG",
			},
			Order: OrderData{
				ExpiryTime:        formattedTomorrow,
				MerchantTransType: itemName,
				OrderTitle:        itemName,
				MerchantTransId:   transactionId,
				OrderMemo:         itemName,
				CreatedTime:       reqTime,
				OrderAmount: Amount{
					Value:    price,
					Currency: "IDR",
				},
				Goods: []GoodsDetail{
					{
						Description: itemName,
						Price: Amount{
							Value:    price,
							Currency: "IDR",
						},
					},
				},
			},
			ProductCode: "51051000100000000001",
			MCC:         "123",
			MerchantID:  merchantId,
			PaymentPreference: PaymentPref{
				SupportDeepLinkCheckoutUrl: true,
			},
			NotificationUrls: []NotifyURL{
				{
					Type: "PAY_RETURN",
					URL:  returnUrl,
				},
				{
					Type: "NOTIFICATION",
					URL:  notifyUrl,
				},
			},
		},
	}

	minifiedData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("error marshalling requestData for sign: %v", err)
	}

	signature, err := helper.GenerateDanaSign(string(minifiedData))
	if err != nil {
		return "", fmt.Errorf("error generating signature: %v", err)
	}

	chargeRequest := CreateOrderPayload{
		Request:   requestData,
		Signature: signature,
	}

	requestBody, err := json.Marshal(chargeRequest)
	if err != nil {
		// return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
		// 	"success": false,
		// 	"message": fmt.Sprintf("Error marshalling request body: %v", err),
		// })
		log.Println("Error marshaling request")
	}

	req, err := http.NewRequest("POST", "https://api.saas.dana.id/dana/acquiring/order/createOrder.htm", bytes.NewReader(requestBody))
	if err != nil {
		log.Println("Error creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	log.Println("Error sending request")
	if err != nil {
		return "", fmt.Errorf("error charging dana: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		log.Println("Error reading response")
	}

	// log.Println("res", string(body))

	var danaResponse DanaResponse
	err = json.Unmarshal(body, &danaResponse)
	if err != nil {
		log.Println("Error decoding response")
	}

	checkoutUrl := danaResponse.Response.Body.CheckoutUrl

	return checkoutUrl, nil
}
