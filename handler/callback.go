package handler

import (
	"app/config"
	"app/database"
	"app/dto/model"
	"app/helper"
	"app/lib"
	"app/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MidtransCallbackRequest struct {
	StatusCode        *string `json:"status_code"`
	TransactionID     *string `json:"transaction_id"`
	GrossAmount       *string `json:"gross_amount"`
	Currency          *string `json:"currency"`
	OrderID           *string `json:"order_id"`
	PaymentType       *string `json:"payment_type"`
	SignatureKey      *string `json:"signature_key"`
	TransactionStatus *string `json:"transaction_status"`
	FraudStatus       *string `json:"fraud_status"`
	StatusMessage     *string `json:"status_message"`
	MerchantID        *string `json:"merchant_id"`
	TransactionTime   *string `json:"transaction_time"`
	ExpiryTime        *string `json:"expiry_time"`
}

type XlCallbackRequest struct {
	UserIdentifier  *string `json:"userIdentifier"`
	TransactionId   *string `json:"transactionId"`
	RefferenceId    *string `json:"refferenceId"`
	PartnerId       *string `json:"partnerId"`
	Item            *string `json:"item"`
	ItemDescription *string `json:"itemDescription"`
	BalanceType     *string `json:"balanceType"`
	Amount          *string `json:"amount"`
	Currency        *string `json:"currency"`
	ResultCode      *string `json:"resultCode"`
	ResultDesc      *string `json:"resultDesc"`
}

type CallbackDanaPayload struct {
	Request struct {
		Head struct {
			Function string `json:"function"`
			ClientID string `json:"clientId"`
			Version  string `json:"version"`
			ReqTime  string `json:"reqTime"`
			ReqMsgID string `json:"reqMsgId"`
		} `json:"head"`
		Body struct {
			AcquirementID     string `json:"acquirementId"`
			OrderAmount       Amount `json:"orderAmount"`
			MerchantID        string `json:"merchantId"`
			MerchantTransId   string `json:"merchantTransId"`
			FinishedTime      string `json:"finishedTime"`
			CreatedTime       string `json:"createdTime"`
			AcquirementStatus string `json:"acquirementStatus"`
			PaymentView       struct {
				PayOptionInfos []struct {
					TransAmount struct {
						Currency string `json:"currency"`
						Value    string `json:"value"`
					} `json:"transAmount"`
					PayAmount struct {
						Currency string `json:"currency"`
						Value    string `json:"value"`
					} `json:"payAmount"`
					PayMethod    string `json:"payMethod"`
					ChargeAmount struct {
						Currency string `json:"currency"`
						Value    string `json:"value"`
					} `json:"chargeAmount"`
					ExtendInfo              string `json:"extendInfo"`
					PayOptionBillExtendInfo string `json:"payOptionBillExtendInfo"`
				} `json:"payOptionInfos"`
				CashierRequestID     string `json:"cashierRequestId"`
				PaidTime             string `json:"paidTime"`
				PayRequestExtendInfo string `json:"payRequestExtendInfo"`
				ExtendInfo           string `json:"extendInfo"`
			} `json:"paymentView"`
			ExtendInfo string `json:"extendInfo"`
		} `json:"body"`
	} `json:"request"`
	Signature string `json:"signature"`
}

type DanaFaspayPaymentNotification struct {
	Request           string `json:"request"`
	TrxID             string `json:"trx_id"`
	MerchantID        string `json:"merchant_id"`
	Merchant          string `json:"merchant"`
	BillNo            string `json:"bill_no"`
	PaymentReff       string `json:"payment_reff"`
	PaymentDate       string `json:"payment_date"`
	PaymentStatusCode string `json:"payment_status_code"`
	PaymentStatusDesc string `json:"payment_status_desc"`
	BillTotal         string `json:"bill_total"`
	PaymentTotal      string `json:"payment_total"`
	PaymentChannelUID string `json:"payment_channel_uid"`
	PaymentChannel    string `json:"payment_channel"`
	Signature         string `json:"signature"`
}

type DanaFaspayCallbackResponse struct {
	Response     string `json:"response"`
	TrxID        string `json:"trx_id"`
	MerchantID   string `json:"merchant_id"`
	Merchant     string `json:"merchant"`
	BillNo       string `json:"bill_no"`
	ResponseCode string `json:"response_code"`
	ResponseDesc string `json:"response_desc"`
	ResponseDate string `json:"response_date"`
}

type Amount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}

type DanaCallbackResponse struct {
	Response  DanaCallbackResponseBody `json:"response"`
	Signature string                   `json:"signature"`
}

type DanaCallbackResponseBody struct {
	Head DanaCallbackResponseHead        `json:"head"`
	Body DanaCallbackResponseBodyContent `json:"body"`
}

type DanaCallbackResponseHead struct {
	Version  string `json:"version"`
	Function string `json:"function"`
	ClientID string `json:"clientId"`
	RespTime string `json:"respTime"`
	ReqMsgId string `json:"reqMsgId"`
}

type DanaCallbackResponseBodyContent struct {
	ResultInfo DanaCallbackResultInfo `json:"resultInfo"`
}

type DanaCallbackResultInfo struct {
	ResultStatus string `json:"resultStatus"`
	ResultCodeId string `json:"resultCodeId"`
	ResultCode   string `json:"resultCode"`
	ResultMsg    string `json:"resultMsg"`
}

type DigiphCallbackPayload struct {
	ID             string  `json:"id"`
	ReferenceID    string  `json:"referenceId"`
	Status         string  `json:"status"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	PaidAt         string  `json:"paidAt"`
	PaymentMethod  string  `json:"paymentMethod"`
	PaymentChannel string  `json:"paymentChannel"`
	Description    string  `json:"description"`
}

// Tambahkan: struct parsing chargeDetails[].card
type BinInformations struct {
	Type        string `json:"type"`
	IssuingBank string `json:"issuingBank"`
	Brand       string `json:"brand"`
	Country     string `json:"country"`
}
type AuthenticationResult struct {
	ThreeDsVersion string `json:"threeDsVersion"`
	ThreeDsResult  string `json:"threeDsResult"`
	ThreeDsMethod  string `json:"threeDsMethod"`
	EciCode        string `json:"eciCode"`
}
type AuthorizedAmount struct {
	Value    int    `json:"value"`
	Currency string `json:"currency"`
}
type AuthorizationResult struct {
	AcquirerReferenceNumber  string           `json:"acquirerReferenceNumber"`
	RetrievalReferenceNumber string           `json:"retrievalReferenceNumber"`
	Stan                     string           `json:"stan"`
	AvsResult                string           `json:"avsResult"`
	CvvResult                string           `json:"cvvResult"`
	AuthorizedAmount         AuthorizedAmount `json:"authorizedAmount"`
	IssuerAuthorizationCode  string           `json:"issuerAuthorizationCode"`
}

type CardDetail struct {
	First6               string               `json:"first6"`
	First8               string               `json:"first8"`
	Last4                string               `json:"last4"`
	ExpMonth             string               `json:"expMonth"`
	ExpYear              string               `json:"expYear"`
	BinInformations      BinInformations      `json:"binInformations"`
	AuthenticationResult AuthenticationResult `json:"authenticationResult"`
	AuthorizationResult  AuthorizationResult  `json:"authorizationResult"`
}

type ChargeDetail struct {
	ID                              string `json:"id"`
	PaymentSessionId                string `json:"paymentSessionId"`
	PaymentSessionClientReferenceId string `json:"paymentSessionClientReferenceId"`
	Amount                          struct {
		Value    int    `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	StatementDescriptor string      `json:"statementDescriptor"`
	Status              string      `json:"status"`
	FailureCode         string      `json:"failureCode"`
	FailureMessage      string      `json:"failureMessage"`
	Recommendation      string      `json:"recommendation"`
	CreatedAt           string      `json:"createdAt"`
	UpdatedAt           string      `json:"updatedAt"`
	PaidAt              string      `json:"paidAt"`
	Card                *CardDetail `json:"card"`
}

type HarsyaCallbackFull struct {
	Event string `json:"event"`
	Data  struct {
		ID                string `json:"id"`
		ClientReferenceID string `json:"clientReferenceId"`
		Status            string `json:"status"`
		PaymentMethod     struct {
			Type string `json:"type"`
		} `json:"paymentMethod"`
		StatementDescriptor string         `json:"statementDescriptor"`
		Amount              AmountPivot    `json:"amount"`
		ChargeDetails       []ChargeDetail `json:"chargeDetails"`
	} `json:"data"`
}

type AmountPivot struct {
	Currency string `json:"currency"`
	Value    int    `json:"value"`
}

func CallbackTriyakom(c *fiber.Ctx) error {
	// ximpayId := c.Query("ximpayid")
	ximpayStatus := c.Query("ximpaystatus")
	cbParam := c.Query("cbparam")
	ipChannel := c.IP()

	// ximpaytoken := c.Query("ximpaytoken")
	failcode := c.Query("failcode")
	transactionId := cbParam[2:]

	// Capture all query parameters
	req := make(map[string]string)
	for key, value := range c.Queries() {
		req[key] = value
	}

	// log.Println("cbParam", cbParam)
	// log.Println("ximpayStatus", ximpayStatus)
	now := time.Now()

	receiveCallbackDate := &now

	helper.TriyakomLogger.LogCallback(transactionId, true,
		map[string]interface{}{
			"transaction_id":   transactionId,
			"ip":               ipChannel,
			"request_callback": req,
		},
	)

	switch ximpayStatus {
	case "1":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionId, 1003, nil, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionId, err)
		}

	case "2":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionId, 1005, "", "Insufficient balance"); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transactionId, err)
		}
	case "3":
		var failReason string

		switch failcode {
		case "301":
			failReason = "User not input phone number"
		case "302":
			failReason = "User not send sms"
		case "303":
			failReason = "User not input PIN"
		case "304":
			failReason = "Failed send PIN to user"
		case "305":
			failReason = "Over limit balance"
		case "306":
			failReason = "Failed Charge"
		case "307":
			failReason = "Failed send to operator"
		default:
			failReason = "Failed / Undeliverable"

		}
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionId, 1005, "", failReason); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transactionId, err)
		}

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid Payment",
		})

	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Success",
	})
}

func MoTelkomsel(c *fiber.Ctx) error {
	msisdn := c.Query("msisdn")
	sms := c.Query("sms")
	trxId := c.Query("trx_id")

	// Pastikan sms memiliki format yang diharapkan
	parts := strings.Split(sms, " ")
	if len(parts) != 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid SMS format",
		})
	}

	keyword := strings.ToUpper(parts[0])
	otp, err := strconv.Atoi(parts[1])
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid OTP format",
		})
	}

	// log.Println("Parsed keyword:", keyword, "OTP:", otp)

	beautifyMsisdn := helper.BeautifyIDNumber(msisdn, true)

	transaction, err := repository.GetTransactionMoTelkomsel(context.Background(), beautifyMsisdn, keyword, otp)
	if err != nil {
		log.Println("Error get transactions tsel:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal server error",
		})
	}

	if transaction == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Transaction not found",
		})
	}

	// log.Println("transaction", transaction)

	denom := fmt.Sprintf("%d", transaction.Amount)
	res, err := lib.RequestMtTsel(transaction.UserMDN, trxId, denom)
	if err != nil {
		log.Println("Mt request failed:", err)
		return fmt.Errorf("mt request failed: %v", err)
	}

	now := time.Now()

	receiveCallbackDate := &now

	// log.Println("Mt request status for id ", transaction.ID, "is", res.Status)

	switch res.Status {
	case "1":
		if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1003, &trxId, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transaction.ID, err)
		}
	case "3:3:21":
		if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, &trxId, nil, "Not enough credit", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
		}
	case "5:997":
		if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, &trxId, nil, "Invalid trx_id", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
		}
	case "3:101":
		if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, &trxId, nil, "Charging timeout", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
		}
	case "2":
		if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, &trxId, nil, "Authentication Failed", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
		}
	case "4:4:2":
		if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, 1005, &trxId, nil, "The provided “tid” by CP is not allowed", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
		}
	}

	if database.RedisClient != nil {
		ctx := context.Background()
		// Dua kemungkinan format key yang digunakan di codebase:
		// 1) repository.GetTransactionMoTelkomsel: "tx:%s:%s:%d"
		// 2) lib.RequestMoTsel: "tsel:tx:%s:%s:%d"
		altMsisdn := helper.BeautifyIDNumber(msisdn, false)
		keys := []string{
			fmt.Sprintf("tsel:tx:%s:%s:%d", beautifyMsisdn, keyword, otp),
		}
		if altMsisdn != beautifyMsisdn {
			keys = append(keys,
				fmt.Sprintf("tsel:tx:%s:%s:%d", altMsisdn, keyword, otp),
			)
		}
		for _, k := range keys {
			if err := database.RedisClient.Del(ctx, k).Err(); err != nil {
				log.Printf("failed to delete redis key %s: %v", k, err)
			}
		}
	}

	return c.JSON(fiber.Map{
		"message": "MO request received",
	})
}

func MidtransCallback(c *fiber.Ctx) error {
	var req MidtransCallbackRequest
	ipClient := c.IP()

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if req.OrderID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing required fields",
		})
	}

	transactionID := *req.OrderID

	transaction, err := repository.GetTransactionByID(context.Background(), transactionID)
	if err != nil || transaction == nil {
		return nil
	}

	statusCallback := true

	strAmount := fmt.Sprintf("%d.00", transaction.Amount)

	message := "success"

	if req.GrossAmount == nil || strAmount != *req.GrossAmount {
		statusCallback = false
		message = "amount doesn't match"
	}

	helper.QrisLogger.LogCallback(transactionID, statusCallback,
		map[string]interface{}{
			"transaction_id":   transactionID,
			"ip":               ipClient,
			"message":          message,
			"request_callback": req,
		},
	)

	now := time.Now()

	receiveCallbackDate := &now

	switch *req.TransactionStatus {
	case "settlement", "capture":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1003, nil, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", *req.TransactionID, err)
		}
	case "expire":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Transaction expired"); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", *req.TransactionID, err)
		}
	case "cancel", "deny", "failure":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Transaction failed"); err != nil {
			log.Printf("Error updating transaction status for %s to failed: %s", *req.TransactionID, err)
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid transaction status",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Callback processed successfully",
	})
}

func XLCallback(c *fiber.Ctx) error {
	var req XlCallbackRequest
	ipClient := c.IP()

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if req.TransactionId == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing required fields",
		})
	}

	transactionID := *req.TransactionId

	transaction, err := repository.GetTransactionByID(context.Background(), transactionID)
	if err != nil || transaction == nil {
		return nil
	}

	statusCallback := true

	strAmount := fmt.Sprintf("%d", transaction.Amount)

	message := "success"

	if req.Amount == nil || strAmount != *req.Amount {
		statusCallback = false
		message = "amount doesn't match"
	}

	helper.XLLogger.LogCallback(transactionID, statusCallback,
		map[string]interface{}{
			"transaction_id":   transactionID,
			"ip":               ipClient,
			"message":          message,
			"request_callback": req,
		},
	)

	// Forward payload ke endpoint eksternal (non-blocking)
	go func(payload []byte) {
		forwardURL := "https://sandbox-payment.redision.com/api/notify/xl"
		reqFwd, err := http.NewRequest(http.MethodPost, forwardURL, bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("XLCallback forward build request error: %v", err)
			return
		}
		reqFwd.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(reqFwd)
		if err != nil {
			log.Printf("XLCallback forward send error: %v", err)
			return
		}
		defer resp.Body.Close()
	}(c.Body())

	now := time.Now()

	receiveCallbackDate := &now

	switch *req.ResultCode {
	case "00":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1003, nil, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", *req.TransactionId, err)
		}
	case "11", "12":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Insufficient Balance"); err != nil {
			log.Printf("Error updating transaction status for %s to expired: %s", *req.TransactionId, err)
		}
	case "20", "21", "22":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Invalid MSISDN or not found"); err != nil {
			log.Printf("Error updating transaction status for %s to failed: %s", *req.TransactionId, err)
		}
	case "30", "31", "39", "40", "41", "45", "46", "47", "48", "99", "49", "50":
		if err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Transaction Failed"); err != nil {
			log.Printf("Error updating transaction status for %s to failed: %s", *req.TransactionId, err)
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid transaction status",
		})
	}

	// Set response headers
	c.Set("Content-Type", "application/json")

	return c.Status(200).JSON(fiber.Map{
		"status":  "success",
		"message": "Callback processed successfully",
	})
}

func DanaCallback(c *fiber.Ctx) error {
	// log.Println("Raw Request Body:\n", string(body))
	loc := time.FixedZone("IST", 5*60*60+30*60)
	ipChannel := c.IP()

	var req CallbackDanaPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	resTime := time.Now().In(loc).Format("2006-01-02T15:04:05-07:00")

	transactionID := req.Request.Body.MerchantTransId

	transaction, err := repository.GetTransactionByID(context.Background(), transactionID)
	if err != nil || transaction == nil {
		return nil
	}

	// minifiedData, err := json.Marshal(req.Request)
	// if err != nil {
	// 	return fmt.Errorf("error marshalling requestData for sign: %v", err)
	// }

	// expectedSignature, err := helper.GenerateDanaSign(string(minifiedData))
	// if err != nil {
	// 	return fmt.Errorf("error generating signature: %v", err)
	// }
	// log.Println("expectedSignature: ", expectedSignature)
	// log.Println("signature: ", req.Signature)

	// if req.Signature != expectedSignature {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
	// 		"success": false,
	// 		"message": "Invalid bodysign",
	// 	})
	// }

	// reqJSON, _ := json.MarshalIndent(req, "", "  ")
	// log.Println("Parsed Request JSON:\n", string(reqJSON))

	helper.DanaLogger.LogCallback(transactionID, true,
		map[string]interface{}{
			"transaction_id":   transactionID,
			"ip":               ipChannel,
			"request_callback": req,
		},
	)

	status := req.Request.Body.AcquirementStatus
	referenceId := req.Request.Body.AcquirementID
	now := time.Now()

	receiveCallbackDate := &now

	switch status {
	case "SUCCESS":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1003, &referenceId, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "CLOSED":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, &referenceId, nil, "order is closed", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "CANCELLED":
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, &referenceId, nil, "order is cancelled", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	}

	var resp DanaCallbackResponse

	respBody := DanaCallbackResponseBody{
		Head: DanaCallbackResponseHead{
			Version:  req.Request.Head.Version,
			Function: req.Request.Head.Function,
			ClientID: req.Request.Head.ClientID,
			RespTime: resTime,
			ReqMsgId: req.Request.Head.ReqMsgID,
		},
		Body: DanaCallbackResponseBodyContent{
			ResultInfo: DanaCallbackResultInfo{
				ResultStatus: "S",
				ResultCodeId: "00000000",
				ResultCode:   "SUCCESS",
				ResultMsg:    "success",
			},
		},
	}

	minifiedDataResp, err := json.Marshal(respBody)
	if err != nil {
		return fmt.Errorf("error marshalling requestData for sign: %v", err)
	}

	respSignature, err := helper.GenerateDanaSign(string(minifiedDataResp))
	if err != nil {
		return fmt.Errorf("error generating signature: %v", err)
	}

	resp = DanaCallbackResponse{
		Response:  respBody,
		Signature: respSignature,
	}

	return c.JSON(resp)
}

func DanaFaspayCallback(c *fiber.Ctx) error {
	body := c.Body()
	ipChannel := c.IP()

	var req DanaFaspayPaymentNotification
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	transactionID := req.BillNo

	transaction, err := repository.GetTransactionByID(context.Background(), transactionID)
	if err != nil || transaction == nil {
		return nil
	}

	status := req.PaymentStatusCode
	now := time.Now()

	helper.FaspayLogger.LogCallback(transactionID, true,
		map[string]interface{}{
			"transaction_id":   transactionID,
			"ip":               ipChannel,
			"request_callback": req,
		},
	)

	receiveCallbackDate := &now

	switch status {
	case "2":
		log.Println("Success Request Body:\n", string(body))
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1003, nil, nil, "", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "0":
		log.Println("CLOSED Request Body:\n", string(body))
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "Unprocessed", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "3":
		log.Println("Failed Request Body:\n", string(body))
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "Payment Failed", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "4":
		log.Println("Reversal Request Body:\n", string(body))
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "Payment Reversal", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "5":
		log.Println("No Bills Found Request Body:\n", string(body))
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "No bills found", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "7":
		log.Println("Payment Expired Request Body:\n", string(body))
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "Payment Expired", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "8":
		log.Println("Payment Cancelled Request Body:\n", string(body))
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "Payment Cancelled", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	case "9":
		log.Println("Unknown Request Body:\n", string(body))
		if err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "Unknown", receiveCallbackDate); err != nil {
			log.Printf("Error updating transaction status for %s: %s", transactionID, err)
		}
	}

	resp := DanaFaspayCallbackResponse{
		Response:     req.Request,
		TrxID:        req.TrxID,
		MerchantID:   req.MerchantID,
		Merchant:     req.Merchant,
		BillNo:       req.BillNo,
		ResponseCode: "00",
		ResponseDesc: "Success",
		ResponseDate: now.Format("2006-01-02 15:04:05"),
	}

	return c.JSON(resp)
}

func CallbackHarsya(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	ipChannel := c.IP()

	if apiKey == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Missing X-API-Key header",
		})
	}

	var req HarsyaCallbackFull
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	if strings.ToUpper(req.Event) == "PAYMENT.TEST" {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Test callback received",
		})
	}

	transactionID := req.Data.ClientReferenceID

	if strings.Count(transactionID, "-") == 5 {
		if lastIdx := strings.LastIndex(transactionID, "-"); lastIdx != -1 {
			suffix := transactionID[lastIdx+1:]
			if _, err := strconv.Atoi(suffix); err == nil {
				transactionID = transactionID[:lastIdx]
			}
		}
	}

	helper.HarsyaLogger.LogCallback(transactionID, true,
		map[string]interface{}{
			"transaction_id":   transactionID,
			"ip":               ipChannel,
			"request_callback": req,
		},
	)

	transaction, err := repository.GetTransactionByID(context.Background(), transactionID)
	if err != nil || transaction == nil {
		log.Printf("Transaction not found: %s", transactionID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Transaction not found",
		})
	}

	now := time.Now()
	receiveCallbackDate := &now

	switch req.Data.Status {
	case "PROCESSING":
		err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1001, nil, nil, "Processing payment", receiveCallbackDate)
		if err != nil {
			log.Printf("Error updating transaction %s to PROCESSING: %s", transactionID, err)
		}
	case "PAID":
		err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1003, nil, nil, "Payment completed", receiveCallbackDate)
		if err != nil {
			log.Printf("Error updating transaction %s to PAID: %s", transactionID, err)
		}

		// Hapus cache setelah pembayaran berhasil untuk mencegah retry
		if database.RedisClient != nil {
			ctx := context.Background()

			// Dapatkan token dari reverse mapping
			reverseKey := fmt.Sprintf("cc_token_map:%s", transactionID)
			token, err := database.RedisClient.Get(ctx, reverseKey).Result()
			if err == nil && token != "" {
				// Hapus cache utama
				cacheKey := fmt.Sprintf("cc_payment:%s", token)
				if err := database.RedisClient.Del(ctx, cacheKey).Err(); err != nil {
					log.Printf("Error deleting cache for token %s: %s", token, err)
				}

				// Hapus reverse mapping
				if err := database.RedisClient.Del(ctx, reverseKey).Err(); err != nil {
					log.Printf("Error deleting reverse mapping for transaction %s: %s", transactionID, err)
				}

				// Juga coba hapus dari in-memory cache jika ada
				TransactionCache.Delete(token)
			} else {
				log.Printf("No token mapping found for transaction %s", transactionID)
			}
		}
	case "FAILED":
		err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "Transaction failed", receiveCallbackDate)
		if err != nil {
			log.Printf("Error updating transaction %s to FAILED: %s", transactionID, err)
		}
	case "EXPIRED":
		err := repository.UpdateTransactionStatusExpired(context.Background(), transactionID, 1005, "", "Transaction expired")
		if err != nil {
			log.Printf("Error updating transaction %s to EXPIRED: %s", transactionID, err)
		}
	case "CANCELLED":
		err := repository.UpdateTransactionStatus(context.Background(), transactionID, 1005, nil, nil, "Transaction cancelled", receiveCallbackDate)
		if err != nil {
			log.Printf("Error updating transaction %s to CANCELLED: %s", transactionID, err)
		}
	}

	if len(req.Data.ChargeDetails) > 0 && strings.ToUpper(req.Data.PaymentMethod.Type) == "CARD" {

		for _, chg := range req.Data.ChargeDetails {
			if chg.Card == nil {
				continue
			}
			logCard := model.CreditCardLog{
				PaymentSessionID:                req.Data.ID,
				PaymentSessionClientReferenceID: req.Data.ClientReferenceID,
				StatementDescriptor:             chg.StatementDescriptor,
				Status:                          chg.Status,
				FailureCode:                     chg.FailureCode,
				FailureMessage:                  chg.FailureMessage,
				Recommendation:                  chg.Recommendation,
				First6:                          chg.Card.First6,
				First8:                          chg.Card.First8,
				Last4:                           chg.Card.Last4,
				ExpMonth:                        chg.Card.ExpMonth,
				ExpYear:                         chg.Card.ExpYear,
				CardType:                        chg.Card.BinInformations.Type,
				Brand:                           chg.Card.BinInformations.Brand,
				IssuingBank:                     chg.Card.BinInformations.IssuingBank,
				BinCountry:                      chg.Card.BinInformations.Country,
				ThreeDsVersion:                  chg.Card.AuthenticationResult.ThreeDsVersion,
				ThreeDsResult:                   chg.Card.AuthenticationResult.ThreeDsResult,
				ThreeDsMethod:                   chg.Card.AuthenticationResult.ThreeDsMethod,
				EciCode:                         chg.Card.AuthenticationResult.EciCode,
				AcquirerReferenceNumber:         chg.Card.AuthorizationResult.AcquirerReferenceNumber,
				RetrievalReferenceNumber:        chg.Card.AuthorizationResult.RetrievalReferenceNumber,
				Stan:                            chg.Card.AuthorizationResult.Stan,
				AvsResult:                       chg.Card.AuthorizationResult.AvsResult,
				CvvResult:                       chg.Card.AuthorizationResult.CvvResult,
				AuthorizedAmountValue:           chg.Card.AuthorizationResult.AuthorizedAmount.Value,
				AuthorizedAmountCurrency:        chg.Card.AuthorizationResult.AuthorizedAmount.Currency,
				IssuerAuthorizationCode:         chg.Card.AuthorizationResult.IssuerAuthorizationCode,
				CreatedAt:                       time.Now(), UpdatedAt: time.Now(),
			}
			_ = repository.InsertCreditCardLog(context.Background(), database.DB, logCard)
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Callback processed successfully",
	})
}

func DigiphCallback(c *fiber.Ctx) error {

	verificationToken := config.Config("DIGIPH_VERIFICATION_TOKEN", "")
	token := c.Get("Verification-Token")
	if token != verificationToken {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid verification token",
		})
	}

	var payload DigiphCallbackPayload
	if err := c.BodyParser(&payload); err != nil {
		log.Println("Failed to parse callback body:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid payload format",
		})
	}

	var paidAt *time.Time
	if payload.PaidAt != "" {
		t, err := time.Parse(time.RFC3339, payload.PaidAt)
		if err == nil {
			paidAt = &t
		}
	}

	var statusCode int
	switch payload.Status {
	case "success":
		statusCode = 1000
	case "failed":
		statusCode = 1005
	case "expired":
		statusCode = 1005
	default:
		statusCode = 1001
	}

	err := repository.UpdateTransactionStatus(context.Background(), payload.ReferenceID, statusCode, &payload.ID, nil, payload.Status, paidAt)
	if err != nil {
		log.Printf("Failed to update transaction status: %s", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update transaction",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Callback received and processed",
	})
}

func ProcessUpdateTransactionPending() {

	for {
		transactions, err := repository.GetPendingTransactions(context.Background(), "telkomsel_airtime")

		if err != nil {
			log.Printf("Error retrieving pending transactions: %s", err)
		}

		for _, transaction := range transactions {
			if err != nil {
				log.Println("Error parsing CreatedAt for transaction:", transaction.ID, err)
				continue
			}

			createdAt := transaction.CreatedAt
			timeLimit := time.Now().Add(-20 * time.Minute)

			expired := createdAt.Before(timeLimit)

			if expired {
				if err := repository.UpdateTransactionStatusExpired(context.Background(), transaction.ID, 1005, "", "Transaction Expired"); err != nil {
					log.Printf("Error updating transaction status for %s to expired: %s", transaction.ID, err)
				}
			}
		}

		time.Sleep(15 * time.Minute)
	}
}
