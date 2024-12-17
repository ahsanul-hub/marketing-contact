package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
)

func GenerateBodySign(body interface{}, appSecret string) (string, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write(bodyJSON)
	signature := h.Sum(nil)

	b64Signature := base64.StdEncoding.EncodeToString(signature)

	bodysign := strings.ReplaceAll(b64Signature, "+", "-")
	bodysign = strings.ReplaceAll(bodysign, "/", "_")

	return bodysign, nil
}
