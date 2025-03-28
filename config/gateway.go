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
		"xl_twt": {
			DirView: "xl_twt",
			Driver:  "XlTwt_Xl",
			Options: map[string]interface{}{
				"development": map[string]interface{}{
					"clientid":      "3S7QIae30ToXBghLAdoQY8V8rWnlYqiA",
					"clientsecret":  "2bgqX6U4UUCjbsCGCLAB0kyl5x04WQGA",
					"partnerid":     "RDSN",
					"tokenurl":      "https://sdp.xlaxiata.co.id/dcb-nongoogle/oauth2/token",
					"inquiryurl":    "https://sdp.xlaxiata.co.id/dcb-nongoogle/accounts/inquiry",
					"chargingurl":   "https://sdp.xlaxiata.co.id/dcb-nongoogle/ocs/charge",
					"checkurl":      "https://sdp.xlaxiata.co.id/dcb-nongoogle/ocs/checktrans",
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
		"smartfren_triyakom": {
			DirView: "smartfren_triyakom",
			Driver:  "Triyakom_SmartfrenFlexible",
			Options: map[string]interface{}{
				"development": map[string]interface{}{
					"partnerid":   "REDIS",
					"partnername": "Redision",
					"seckey":      "DE9D7033E2584FCBBC479FFD654F44C7",
					"requestUrl":  "https://secure.ximpay.com/api/dev10SDPflex/Gopayment.aspx",
					"confirmUrl":  "https://secure.ximpay.com/api/dev10SDPflex/Gopin.aspx",
					"dir":         "smartfren_triyakom_flex",
				},
				"production": map[string]interface{}{
					"partnerid":   "REDIS",
					"partnername": "Redision",
					"seckey":      "DE9D7033E2584FCBBC479FFD654F44C7",
					"requestUrl":  "https://secure.ximpay.com/api/10SDPflex/Gopayment.aspx",
					"confirmUrl":  "https://secure.ximpay.com/api/10SDPflex/Gopin.aspx",
					"dir":         "smartfren_triyakom_flex",
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
					"requestUrl":  "https://secure.ximpay.com/api/07flex/Gopayment.aspx",
					"dir":         "indosat_triyakom",
				},
			},
			Lang: "id",
		},
		"tri": {
			DirView: "three_triyakom",
			Driver:  "Triyakom_Tri",
			Denom: map[string]map[string]string{
				"1000":   {"keyword": "RED00003"},
				"2000":   {"keyword": "RED00054"},
				"3000":   {"keyword": "RED00004"},
				"5000":   {"keyword": "RED00005"},
				"10000":  {"keyword": "RED00006"},
				"20000":  {"keyword": "RED00008"},
				"25000":  {"keyword": "RED00009"},
				"30000":  {"keyword": "RED00056"},
				"50000":  {"keyword": "RED00011"},
				"60000":  {"keyword": "RED00057"},
				"100000": {"keyword": "RED00002"},
				"200000": {"keyword": "RED00016"},
				"250000": {"keyword": "RED00017"},
				"500000": {"keyword": "RED00019"},
			},
			Options: map[string]interface{}{
				"development": map[string]interface{}{
					"partnerid":   "REDIS",
					"partnername": "Redision",
					"seckey":      "DE9D7033E2584FCBBC479FFD654F44C7",
					"requestUrl":  "https://secure.ximpay.com/api/dev03/Gopayment.aspx",
					"dir":         "tri_triyakom",
				},
				"production": map[string]interface{}{
					"partnerid":   "REDIS",
					"partnername": "Redision",
					"seckey":      "DE9D7033E2584FCBBC479FFD654F44C7",
					"requestUrl":  "https://secure.ximpay.com/api/03/Gopayment.aspx",
					"dir":         "tri_triyakom",
				},
			},
			Lang: "id",
		},
		"telkomsel_airtime_sms": {
			DirView: "telkomsel",
			Driver:  "Telkomsel",
			Denom: map[string]map[string]string{
				"350": {"keyword": "REDS350", "sid": "GAMGENRRGREDS350_IOD", "denom": "350", "price": "555", "tid": "141"},
				// "1100":   {"keyword": "REDS770", "sid": "GAMGENRRGREDS770_IOD", "denom": "770", "price": "1221", "tid": ""},
				// "1300":   {"keyword": "REDS910", "sid": "GAMGENRRGREDS910_IOD", "denom": "910", "price": "1443", "tid": ""},
				// "1500":   {"keyword": "REDS1050", "sid": "GAMGENRRGREDS1050_IOD", "denom": "1050", "price": "1665", "tid": ""},
				// "2200":   {"keyword": "REDS1540", "sid": "GAMGENRRGREDS1540_IOD", "denom": "1540", "price": "2442", "tid": ""},
				// "4000":   {"keyword": "REDS2800", "sid": "GAMGENRRGREDS2800_IOD", "denom": "2800", "price": "3108", "tid": ""},
				// "5045":   {"keyword": "REDS3532", "sid": "GAMGENRRGREDS3532_IOD", "denom": "3532", "price": "3921", "tid": ""},
				// "5300":   {"keyword": "REDS3710", "sid": "GAMGENRRGREDS3710_IOD", "denom": "3710", "price": "4119", "tid": ""},
				// "5500":   {"keyword": "REDS3850", "sid": "GAMGENRRGREDS3850_IOD", "denom": "3850", "price": "4274", "tid": ""},
				// "6000":   {"keyword": "REDS4200", "sid": "GAMGENRRGREDS4200_IOD", "denom": "4200", "price": "4662", "tid": ""},
				// "7700":   {"keyword": "REDS5390", "sid": "GAMGENRRGREDS5390_IOD", "denom": "5390", "price": "5983", "tid": ""},
				// "10091":  {"keyword": "REDS7064", "sid": "GAMGENRRGREDS7064_IOD", "denom": "7064", "price": "7831", "tid": ""},
				// "10700":  {"keyword": "REDS7490", "sid": "GAMGENRRGREDS7490_IOD", "denom": "7490", "price": "8314", "tid": ""},
				// "11000":  {"keyword": "REDS048", "sid": "GAMGENRRGREDS048_IOD", "denom": "11000", "price": "12210", "tid": ""},
				// "11450":  {"keyword": "REDS8015", "sid": "GAMGENRRGREDS8015_IOD", "denom": "11450", "price": "12710", "tid": ""},
				// "12500":  {"keyword": "REDS8750", "sid": "GAMGENRRGREDS8750_IOD", "denom": "12500", "price": "13875", "tid": ""},
				// "13000":  {"keyword": "REDS052", "sid": "GAMGENRRGREDS052_IOD", "denom": "13000", "price": "14430", "tid": ""},
				// "16000":  {"keyword": "REDS11200", "sid": "GAMGENRRGREDS11200_IOD", "denom": "16000", "price": "17760", "tid": ""},
				// "16500":  {"keyword": "REDS11550", "sid": "GAMGENRRGREDS11550_IOD", "denom": "16500", "price": "18315", "tid": ""},
				// "20182":  {"keyword": "REDS14127", "sid": "GAMGENRRGREDS14127_IOD", "denom": "20182", "price": "22402", "tid": ""},
				// "22000":  {"keyword": "REDS076", "sid": "GAMGENRRGREDS076_IOD", "denom": "22000", "price": "24420", "tid": ""},
				// "22900":  {"keyword": "REDS16030", "sid": "GAMGENRRGREDS16030_IOD", "denom": "22900", "price": "25419", "tid": ""},
				// "27500":  {"keyword": "REDS19250", "sid": "GAMGENRRGREDS19250_IOD", "denom": "27500", "price": "30575", "tid": ""},
				// "30273":  {"keyword": "REDS21191", "sid": "GAMGENRRGREDS21191_IOD", "denom": "30273", "price": "33503", "tid": ""},
				// "32900":  {"keyword": "REDS23030", "sid": "GAMGENRRGREDS23030_IOD", "denom": "32900", "price": "36579", "tid": ""},
				// "35000":  {"keyword": "REDS24500", "sid": "GAMGENRRGREDS24500_IOD", "denom": "35000", "price": "38850", "tid": ""},
				// "38500":  {"keyword": "REDS26950", "sid": "GAMGENRRGREDS26950_IOD", "denom": "38500", "price": "42735", "tid": ""},
				// "44000":  {"keyword": "REDS30800", "sid": "GAMGENRRGREDS30800_IOD", "denom": "44000", "price": "48840", "tid": ""},
				// "49000":  {"keyword": "REDS34300", "sid": "GAMGENRRGREDS34300_IOD", "denom": "49000", "price": "54390", "tid": ""},
				// "55000":  {"keyword": "REDS152", "sid": "GAMGENRRGREDS152_IOD", "denom": "55000", "price": "61050", "tid": ""},
				// "57000":  {"keyword": "REDS39900", "sid": "GAMGENRRGREDS39900_IOD", "denom": "57000", "price": "63270", "tid": ""},
				// "60545":  {"keyword": "REDS42382", "sid": "GAMGENRRGREDS42382_IOD", "denom": "60545", "price": "67205", "tid": ""},
				// "65000":  {"keyword": "REDS45500", "sid": "GAMGENRRGREDS45500_IOD", "denom": "65000", "price": "72150", "tid": ""},
				// "65500":  {"keyword": "REDS45850", "sid": "GAMGENRRGREDS45850_IOD", "denom": "65500", "price": "72705", "tid": ""},
				// "65800":  {"keyword": "REDS46060", "sid": "GAMGENRRGREDS46060_IOD", "denom": "65800", "price": "73038", "tid": ""},
				// "66000":  {"keyword": "REDS46200", "sid": "GAMGENRRGREDS46200_IOD", "denom": "66000", "price": "73260", "tid": ""},
				// "75000":  {"keyword": "REDS52500", "sid": "GAMGENRRGREDS52500_IOD", "denom": "75000", "price": "83250", "tid": ""},
				// "77000":  {"keyword": "REDS6101", "sid": "GAMGENRRGREDS6101_IOD", "denom": "77000", "price": "85470", "tid": ""},
				// "79000":  {"keyword": "REDS55300", "sid": "GAMGENRRGREDS55300_IOD", "denom": "79000", "price": "87690", "tid": ""},
				// "80000":  {"keyword": "REDS185", "sid": "GAMGENRRGREDS185_IOD", "denom": "80000", "price": "88800", "tid": ""},
				// "81000":  {"keyword": "REDS56700", "sid": "GAMGENRRGREDS56700_IOD", "denom": "81000", "price": "89910", "tid": ""},
				// "82500":  {"keyword": "REDS57750", "sid": "GAMGENRRGREDS57750_IOD", "denom": "82500", "price": "91575", "tid": ""},
				// "88000":  {"keyword": "REDS61600", "sid": "GAMGENRRGREDS61600_IOD", "denom": "88000", "price": "97680", "tid": ""},
				// "90000":  {"keyword": "REDS197", "sid": "GAMGENRRGREDS197_IOD", "denom": "90000", "price": "99900", "tid": ""},
				// "99000":  {"keyword": "REDS69300", "sid": "GAMGENRRGREDS69300_IOD", "denom": "99000", "price": "109890", "tid": ""},
				// "100909": {"keyword": "REDS1452", "sid": "GAMGENRRGREDS1452_IOD", "denom": "100909", "price": "111009", "tid": ""},
				// "110000": {"keyword": "REDS3767", "sid": "GAMGENRRGREDS3767_IOD", "denom": "110000", "price": "122100", "tid": ""},
				// "113000": {"keyword": "REDS79100", "sid": "GAMGENRRGREDS79100_IOD", "denom": "113000", "price": "125430", "tid": ""},
				// "114300": {"keyword": "REDS80010", "sid": "GAMGENRRGREDS80010_IOD", "denom": "114300", "price": "126873", "tid": ""},
				// "129000": {"keyword": "REDS90300", "sid": "GAMGENRRGREDS90300_IOD", "denom": "129000", "price": "143190", "tid": ""},
				// "131200": {"keyword": "REDS91840", "sid": "GAMGENRRGREDS91840_IOD", "denom": "131200", "price": "145632", "tid": ""},
				// "149000": {"keyword": "REDS104300", "sid": "GAMGENRRGREDS104300_IOD", "denom": "149000", "price": "165390", "tid": ""},
				// "150000": {"keyword": "REDS478", "sid": "GAMGENRRGREDS478_IOD", "denom": "150000", "price": "166500", "tid": ""},
				// "151364": {"keyword": "REDS3190", "sid": "GAMGENRRGREDS3190_IOD", "denom": "151364", "price": "168014", "tid": ""},
				// "159000": {"keyword": "REDS3770", "sid": "GAMGENRRGREDS3770_IOD", "denom": "159000", "price": "176490", "tid": ""},
				// "165000": {"keyword": "REDS3282", "sid": "GAMGENRRGREDS3282_IOD", "denom": "165000", "price": "183150", "tid": ""},
				// "199000": {"keyword": "REDS139300", "sid": "GAMGENRRGREDS139300_IOD", "denom": "199000", "price": "220890", "tid": ""},
				// "220000": {"keyword": "REDS3027", "sid": "GAMGENRRGREDS3027_IOD", "denom": "220000", "price": "244200", "tid": ""},
				// "249000": {"keyword": "REDS174300", "sid": "GAMGENRRGREDS174300_IOD", "denom": "249000", "price": "276390", "tid": ""},
				// "255000": {"keyword": "REDS178500", "sid": "GAMGENRRGREDS178500_IOD", "denom": "255000", "price": "283050", "tid": ""},
				// "259000": {"keyword": "REDS181300", "sid": "GAMGENRRGREDS181300_IOD", "denom": "259000", "price": "287490", "tid": ""},
				// "262400": {"keyword": "REDS183680", "sid": "GAMGENRRGREDS183680_IOD", "denom": "262400", "price": "291264", "tid": ""},
				// "275000": {"keyword": "REDS722", "sid": "GAMGENRRGREDS722_IOD", "denom": "275000", "price": "305250", "tid": ""},
				// "300000": {"keyword": "REDS210K", "sid": "GAMGENRRGREDS210K_IOD", "denom": "300000", "price": "333000", "tid": ""},
				// "302727": {"keyword": "REDS3189", "sid": "GAMGENRRGREDS3189_IOD", "denom": "302727", "price": "335026", "tid": ""},
				// "309000": {"keyword": "REDS216300", "sid": "GAMGENRRGREDS216300_IOD", "denom": "309000", "price": "342990", "tid": ""},
				// "328300": {"keyword": "REDS229810", "sid": "GAMGENRRGREDS229810_IOD", "denom": "328300", "price": "364413", "tid": ""},
				// "330000": {"keyword": "REDS231K", "sid": "GAMGENRRGREDS231K_IOD", "denom": "330000", "price": "366300", "tid": ""},
				// "350000": {"keyword": "REDS724", "sid": "GAMGENRRGREDS724_IOD", "denom": "350000", "price": "388500", "tid": ""},
				// "359000": {"keyword": "REDS251300", "sid": "GAMGENRRGREDS251300_IOD", "denom": "359000", "price": "398490", "tid": ""},
				// "399000": {"keyword": "REDS279300", "sid": "GAMGENRRGREDS279300_IOD", "denom": "399000", "price": "442890", "tid": ""},
				// "400000": {"keyword": "REDS726", "sid": "GAMGENRRGREDS726_IOD", "denom": "400000", "price": "444000", "tid": ""},
				// "410000": {"keyword": "REDS287K", "sid": "GAMGENRRGREDS287K_IOD", "denom": "410000", "price": "455100", "tid": ""},
				// "450000": {"keyword": "REDS315K", "sid": "GAMGENRRGREDS315K_IOD", "denom": "450000", "price": "499500", "tid": ""},
				// "469000": {"keyword": "REDS520590", "sid": "GAMGENRRGREDS520590_IOD", "denom": "469000", "price": "520590", "tid": ""},
				// "479000": {"keyword": "REDS335300", "sid": "GAMGENRRGREDS335300_IOD", "denom": "479000", "price": "531690", "tid": ""},
				// "489000": {"keyword": "REDS342300", "sid": "GAMGENRRGREDS342300_IOD", "denom": "489000", "price": "542790", "tid": ""},
				// "550000": {"keyword": "REDS3033", "sid": "GAMGENRRGREDS3033_IOD", "denom": "550000", "price": "610500", "tid": ""},
				"3000":   {"keyword": "REDS3K", "sid": "GAMGENRRPRREDS3K_IOD", "denom": "3000", "price": "3330", "tid": "96"},
				"5000":   {"keyword": "REDS5K", "sid": "GAMGENRPGREDISION5K_IOD", "denom": "5000", "price": "5550", "tid": "142"},
				"15000":  {"keyword": "REDS15K", "sid": "GAMGENRRPRREDS15K_IOD", "denom": "15000", "price": "16650", "tid": "58"},
				"10000":  {"keyword": "REDS10K", "sid": "GAMGENRPGREDISION10K_IOD", "denom": "10000", "price": "11100", "tid": "43"},
				"25000":  {"keyword": "REDS25K", "sid": "GAMGENRPGREDISION25K_IOD", "denom": "25000", "price": "27750", "tid": "86"},
				"20000":  {"keyword": "REDS20K", "sid": "GAMGENRRPRREDS20K_IOD", "denom": "20000", "price": "22200", "tid": "72"},
				"30000":  {"keyword": "REDS30K", "sid": "GAMGENRRPRREDS30K_IOD", "denom": "30000", "price": "33300", "tid": "98"},
				"40000":  {"keyword": "REDS40K", "sid": "GAMGENRPGREDISION40K_IOD", "denom": "40000", "price": "44400", "tid": "120"},
				"50000":  {"keyword": "REDS50K", "sid": "GAMGENRPGREDISION50K_IOD", "denom": "50000", "price": "55500", "tid": "145"},
				"60000":  {"keyword": "REDS60K", "sid": "GAMGENRRPRREDS60K_IOD", "denom": "60000", "price": "66600", "tid": "159"},
				"70000":  {"keyword": "REDS70K", "sid": "GAMGENRPGREDISION70K_IOD", "denom": "70000", "price": "77700", "tid": "171"},
				"100000": {"keyword": "REDS100K", "sid": "GAMGENRPGREDISION100K_IOD", "denom": "100000", "price": "111000", "tid": "45"},
				"125000": {"keyword": "REDS125K", "sid": "GAMGENRPGREDISION125K_IOD", "denom": "125000", "price": "138750", "tid": "719"},
				"250000": {"keyword": "REDS250K", "sid": "GAMGENRPGREDISION250K_IOD", "denom": "250000", "price": "277500", "tid": "461"},
				"200000": {"keyword": "REDS200K", "sid": "GAMGENRRPRREDS200K_IOD", "denom": "200000", "price": "277500", "tid": "479"},
				"325000": {"keyword": "REDS325K", "sid": "GAMGENRRPRREDS325K_IOD", "denom": "325000", "price": "360750", "tid": "723"},
				"500000": {"keyword": "REDS500K", "sid": "GAMGENRPGREDISION500K_IOD", "denom": "500000", "price": "555000", "tid": "462"},
			},
			Options: map[string]interface{}{
				"moUrl":  "https://api.digitalcore.telkomsel.com/scrt/cp/smsbulk/submit.jsp",
				"mtUrl":  "https://api.digitalcore.telkomsel.com/scrt/cp/submitSM.jsp",
				"secret": "5KCA35pNcy",
				"apikey": "9yt7a9uets3mqbwpz4dwxtsf",
			},
			MT:   map[string]interface{}{},
			Lang: "id",
		},
		"va_bca_direct": {
			DirView: "va_bca_direct",
			Driver:  "Bca_Va",
			Options: map[string]interface{}{
				"production": map[string]interface{}{
					"api_key":    "283256b0-6e1a-4b14-8ca4-0e50c46e1659",
					"api_secret": "b4edb249-6748-474b-9e3a-c6d87bf5e19a",
					"tokenUrl":   "https://sandbox.bca.co.id:443/api/oauth/token",
					"prefix":     "11131",
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
