package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func GenerateBodySign(bodyJson string, appSecret string) (string, error) {

	h := hmac.New(sha256.New, []byte(appSecret))

	// Write the data (bodyJson) to the HMAC
	h.Write([]byte(bodyJson))

	// Get the HMAC result
	signature := h.Sum(nil)

	// Encode the HMAC result to Base64
	base64Encoded := base64.StdEncoding.EncodeToString(signature)

	bodysign := strings.NewReplacer("+", "-", "/", "_").Replace(base64Encoded)

	return bodysign, nil
}
