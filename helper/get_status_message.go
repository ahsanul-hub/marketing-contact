package helper

var StatusMessages = map[string]string{
	"0000":  "Successful",
	"E0001": "Invalid appkey or appid",
	"E0002": "Invalid secret",
	"E0003": "Invalid body signature",
	"E0004": "Invalid Transaction ID",
	"E0005": "This payment method is under maintenance",
	"E0006": "Invalid APP_ID / APP_ID is not active",
	"E0007": "This payment method is not available for this merchant",
	"E0008": "This denom is not supported for this payment method",
	"E0009": "Phone number is needed for this payment method",
	"E0010": "Please wait for 5 minutes before doing another topup",
	"E0011": "Invalid order or this order has been completed",
	"E0012": "Invalid SMS code or SMS code already expired. Please create a new purchase",
	"E0013": "Some field(s) missing",
	"E0014": "This denom is not available",
	"E0015": "Blocked MSISDN!",
	"E0016": "Invalid MSISDN!",
	"E0017": "This MDN has reached the maximum topup!",
	"E0018": "Payment not supported",
	"E0019": "Invalid Data!",
	"E0020": "Payment amount does not meet the minimum requirement",
	"E0021": "Amount exceeds limit!",
	"E0022": "You have no access to make this request!",
	"E0023": "Duplicate merchant_transaction_id",
	"E0099": "System is too busy, please try again later.",
	"E0000": "Unknown error",
	"E4001": "Database Error",
	"1001":  "created",
	"1002":  "waiting_for_payment",
	"1003":  "waiting_for_dr_notification",
	"1000":  "payment_completed",
	"1005":  "failed",
	"999":   "unknown_error",
}

func GetStatusMessage(code string) string {
	if message, exists := StatusMessages[code]; exists {
		return message
	}
	return "Unknown error code"
}
