package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

func strtr(input string) string {
	replacer := strings.NewReplacer("+", "-", "/", "_")
	return replacer.Replace(input)
}

func GenerateBodySign(bodyJson string, appSecret string) (string, error) {
	// bodyJSON, err := json.Marshal(body)
	// if err != nil {
	// 	return "", err
	// }

	// h := hmac.New(sha256.New, []byte(appSecret))
	// h.Write(bodyJSON)
	// signature := h.Sum(nil)

	// b64Signature := base64.StdEncoding.EncodeToString(signature)

	// bodysign := strings.ReplaceAll(b64Signature, "+", "-")
	// bodysign = strings.ReplaceAll(bodysign, "/", "_")

	// return bodysign, nil

	h := hmac.New(sha256.New, []byte(appSecret))

	// Write the data (bodyJson) to the HMAC
	h.Write([]byte(bodyJson))

	// Get the HMAC result
	signature := h.Sum(nil)

	// Encode the HMAC result to Base64
	base64Encoded := base64.StdEncoding.EncodeToString(signature)

	// Replace '+' with '-' and '/' with '_'
	bodysign := strtr(base64Encoded)

	return bodysign, nil
}
