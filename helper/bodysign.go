package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"
)

// func GenerateBodySign(bodyJson string, appSecret string) (string, error) {

// 	h := hmac.New(sha256.New, []byte(appSecret))

// 	// Write the data (bodyJson) to the HMAC
// 	h.Write([]byte(bodyJson))

// 	// Get the HMAC result
// 	signature := h.Sum(nil)

// 	// Encode the HMAC result to Base64
// 	base64Encoded := base64.StdEncoding.EncodeToString(signature)

// 	bodysign := strings.NewReplacer("+", "-", "/", "_").Replace(base64Encoded)

//		return bodysign, nil
//	}

// func CleanPayload(payload interface{}) map[string]interface{} {
// 	cleaned := make(map[string]interface{})
// 	v := reflect.ValueOf(payload)

// 	// Pastikan payload berupa struct
// 	if v.Kind() == reflect.Ptr {
// 		v = v.Elem()
// 	}

// 	// Iterasi field
// 	for i := 0; i < v.NumField(); i++ {
// 		field := v.Type().Field(i)
// 		value := v.Field(i).Interface()

// 		// Hanya tambahkan field jika tidak kosong atau nol
// 		if value != "" && value != 0 && !reflect.DeepEqual(value, reflect.Zero(v.Field(i).Type()).Interface()) {
// 			cleaned[field.Name] = value
// 		}
// 	}

// 	return cleaned
// }

func GenerateBodySign(payload interface{}, appSecret string) string {
	// Convert payload ke JSON (pretty format)
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		return ""
	}

	// Buat HMAC-SHA256 hash
	// log.Println("payloadJSON: ", string(payloadJSON))
	// log.Println("secret: ", appSecret)

	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write(payloadJSON)
	hash := h.Sum(nil)

	// Encode to base64 URL safe
	bodysign := base64.StdEncoding.EncodeToString(hash)
	bodysign = strings.NewReplacer("+", "-", "/", "_").Replace(bodysign)

	return bodysign
}

// Function untuk menghapus padding "=" dari Base64
