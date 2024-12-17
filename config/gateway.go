package config

import (
	"fmt"
)

// GatewayConfig represents the configuration structure for a payment gateway.
type GatewayConfig struct {
	DirView string                       `json:"dir_view"`
	Driver  string                       `json:"driver"`
	Denom   map[string]map[string]string `json:"denom"`
	Options map[string]interface{}       `json:"options"`
	Lang    string                       `json:"lang"`
	MT      map[string]interface{}       `json:"mt"`
}

// GetGatewayConfig retrieves the configuration for a specified gateway.
func GetGatewayConfig(gatewayName string) (GatewayConfig, error) {
	config := map[string]GatewayConfig{
		"telkomsel_airtime_ussd": {
			DirView: "telkomsel_airtime_ussd",
			Driver:  "TelkomselAirtime_Ussd",
			Options: map[string]interface{}{
				"secretKey":    "72Zwth2Dd75yuYzRhgKhGcsdf",
				"appKey":       "7d51a9a750575a294df94a78bde79628",
				"merchantCode": "ID-0031",
				"baseUrl":      "http://3.1.41.116/api/v1/create",
				"serverIp":     []string{"202.53.250.116", "54.169.195.130", "3.1.41.116"},
				"dir":          "telkomsel_airtime_ussd",
			},
			Lang: "id",
		},
		"telkomsel_airtime_sms": {
			DirView: "telkomsel_airtime_sms",
			Driver:  "TelkomselAirtime_Sms",
			Options: map[string]interface{}{
				"secretKey":    "72Zwth2Dd75yuYzRhgKhGcsdf",
				"appKey":       "7d51a9a750575a294df94a78bde79628",
				"merchantCode": "ID-0031",
				"baseUrl":      "http://3.1.41.116/api/v1/create",
				"serverIp":     []string{"202.53.250.116", "54.169.195.130", "3.1.41.116"},
				"dir":          "telkomsel_airtime_sms",
			},
			Lang: "id",
		},
		"xl_twt": {
			DirView: "xl_twt",
			Driver:  "XlTwt_Xl",
			Options: map[string]interface{}{
				"development": map[string]interface{}{
					"clientid":      "3S7QIae30ToXBghLAdoQY8V8rWnlYqiA",
					"clientsecret":  "2bgqX6U4UUCjbsCGCLAB0kyl5x04WQGA",
					"partnerid":     "RDSN",
					"tokenurl":      "https://staging.redigame.co.id/xlproxy/pushtoken",
					"inquiryurl":    "https://staging.redigame.co.id/xlproxy/inquiry",
					"chargingurl":   "https://staging.redigame.co.id/xlproxy/pushcharging",
					"checkurl":      "https://staging.redigame.co.id/xlproxy/check_status",
					"mturl":         "https://staging.redigame.co.id/xlproxy/push_mt",
					"sdp_partnerid": "008003",
					"sdp_password":  "Rdcb!1234",
					"sdp_programid": "0080032000031742",
					"sdp_sdc":       "99899",
					"sdp_sid":       "9989900",
					"sdp_mturl":     "http://10.44.7.5:80/webapp-partnerhub-xl-idn-partner-sms/push-mt",
				},
			},
			Lang: "id",
		},
		"smartfren": {
			DirView: "smartfren",
			Driver:  "Smartfren",
			Denom: map[string]map[string]string{
				"5000":   {"keyword": "RED5K"},
				"10000":  {"keyword": "RED10K"},
				"25000":  {"keyword": "RED25K"},
				"50000":  {"keyword": "RED50K"},
				"100000": {"keyword": "RED100K"},
			},
			Options: map[string]interface{}{
				"ip":          "http://10.14.42.148",
				"port":        "8090",
				"serviceNode": "PTRTI",
				"msgCoding":   "1",
				"sender":      "99899",
				"smscId":      "SMPP",
				"bearerId":    "5102",
			},
			MT: map[string]interface{}{
				"ip":             "http://10.14.42.164",
				"port":           "9078",
				"user":           "cgi",
				"pass":           "cgi123",
				"from":           "99899",
				"ProxyParameter": "mtsms",
				"PartnerID":      "5",
			},
			Lang: "id",
		},
		"smartfren_triyakom_flex2": {
			DirView: "smartfren_triyakom_flex2",
			Driver:  "Triyakom_SmartfrenFlexible2",
			Options: map[string]interface{}{
				"development": map[string]interface{}{
					"partnerid":   "REDIS",
					"partnername": "Redision",
					"seckey":      "DE9D7033E2584FCBBC479FFD654F44C7",
					"requestUrl":  "https://secure.ximpay.com/api/dev10SDPflex/Gopayment.aspx",
					"confirmUrl":  "https://secure.ximpay.com/api/dev10SDPflex/Gopin.aspx",
					"dir":         "smartfren_triyakom_flex2",
				},
				"production": map[string]interface{}{
					"partnerid":   "REDIS",
					"partnername": "Redision",
					"seckey":      "DE9D7033E2584FCBBC479FFD654F44C7",
					"requestUrl":  "https://secure.ximpay.com/api/10SDPflex/Gopayment.aspx",
					"confirmUrl":  "https://secure.ximpay.com/api/10SDPflex/Gopin.aspx",
					"dir":         "smartfren_triyakom_flex2",
				},
			},
			Lang: "id",
		},
		"indosat_triyakom": {
			DirView: "indosat_triyakom",
			Driver:  "Triyakom_Indosat",
			Options: map[string]interface{}{
				"development": map[string]interface{}{
					"partnerid":   "REDIS",
					"partnername": "Redision",
					"seckey":      "DE9D7033E2584FCBBC479FFD654F44C7",
					"requestUrl":  "https://secure.ximpay.com/api/dev07/Gopayment.aspx",
					"dir":         "indosat_triyakom",
				},
				"production": map[string]interface{}{
					"partnerid":   "REDIS",
					"partnername": "Redision",
					"seckey":      "DE9D7033E2584FCBBC479FFD654F44C7",
					"requestUrl":  "https://secure.ximpay.com/api/07/Gopayment.aspx",
					"dir":         "indosat_triyakom",
				},
			},
			Lang: "id",
		},
		// [
		// 	'dir_view'=>'indosat_triyakom',
		// 	'driver'=>'Triyakom_Indosat',
		// 	'options'=>[
		// 		'development'=>[
		// 			'partnerid'=>'REDIS',
		// 			'partnername'=>'Redision',
		// 			// 'seckey'=>'B42D3BC1D3AF4CE694209CC1FA69C7C6',
		// 			'seckey'=>'DE9D7033E2584FCBBC479FFD654F44C7',
		// 			'requestUrl'=>'https://secure.ximpay.com/api/dev07/Gopayment.aspx',
		// 			'dir' => 'indosat_triyakom',

		// 		],
		// 		'production'=>[
		// 			'partnerid'=>'REDIS',
		// 			'partnername'=>'Redision',
		// 			// 'seckey'=>'B42D3BC1D3AF4CE694209CC1FA69C7C6',
		// 			'seckey'=>'DE9D7033E2584FCBBC479FFD654F44C7',
		// 			'requestUrl'=>'https://secure.ximpay.com/api/07/Gopayment.aspx',
		// 			'dir' => 'indosat_triyakom',

		// 		],
		// Tambahkan gateway lainnya di sini...
	}

	if config, exists := config[gatewayName]; exists {
		return config, nil
	}
	return GatewayConfig{}, fmt.Errorf("gateway %s not found", gatewayName)
}

// func main() {
// 	// Contoh penggunaan fungsi GetGatewayConfig
// 	gatewayName := "telkomsel_airtime_ussd"
// 	config, err := GetGatewayConfig(gatewayName)
// 	if err != nil {
// 		fmt.Println("Error:", err)
// 		return
// 	}
// 	fmt.Printf("Configuration for %s: %+v\n", gatewayName, config)
// }
