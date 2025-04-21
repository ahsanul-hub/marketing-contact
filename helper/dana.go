package helper

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

func GenerateDanaSign(data string) (string, error) {

	privateKey := []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC78abeCHePJNdAB8pFdpyOxdj2FWpYLPAjmKZl89qNHAyuvJzRwbPC694mmAys6koY/wXTSqhs+ov7OEN+Fj+N8s0dWuUsrIFF9e1TbZQP09O3T3k+C+0tDW+3RVrrsCAbiophz67YZkWk2FuhwCGH2OGqiMWhBv0zgg669ZNVi7zwbPLNUso1a7trqg37cVYe8w4P8g7wf1GnxqYFtVOafcP73zGF1u1y87UOz9QO73wSnfxi6mDE+B251z/U+31szuMzgHdp/sJjtaDJ3l9iRV7WfLQBvGecQNhZTjWNkWAHt3T9gg2M8PrO9Fm2yfIm1PFbsYhbmTU0es8JrHntAgMBAAECggEAYyzuiCXhqVigeXpi42rmzHRcu+arGmKESdRook4e2u2dR6vh+NIFYOuEa8s6jRiJB02zrj6sR+2iZmvXObbVzLr+P+pSGtPg16Ehni+pvPxjsUyvxu0WN/rqI8TmaI6lMsNVqK2mLy0wvP8qw10WlI/+7TWFTCbbAA42ZbPnDnFpV9b5gIcqOsQAZEGoDRI+PQwK5uhq/+ZCaN6pB9X+5xoib6Lfzb3udQjVkcnJ/d/WkGSM//SWfOirxCDBILA4pT7fWBskyJvg99JWS3tRAByFC3r0UzHgca8DyeDO4WWtuhIm8GKvX9GeFk88smMCnOuaP/ETsrFl5iH68uRwxQKBgQDvo1Pc6QE/7sM0HyyEeBI/iJ0RerLudVvgm50TUYXCiDgcECJ2Yua4Y2xfaWHELmwH/dYbk/TbMx5TRHaViPhe3UumeGYye0TP7Y6Z/X3m7apvrErsfV8P37chOrvDVzooFDgtFWp+I8Ho33b5jhXlNSiRYoi6Ni+batml9VNFEwKBgQDIxsFiw05QgApnJ8JIHyWmbDa0Qtjiblo5neT/SYQPf2RMWLEqQgnAcxrLfxrJJhc2C/LiN/OWIBphQYT1K3QX9EGvE8BJGDR3/V0o+DPQsJz/lWHjuCpqlDxGQxG7zH7yXqQ/sXq0ElaIi8ajT0eiQo8D8Q0fZcEQZqqqJmwk/wKBgF0Q5UTqCN43cAASC0v3Bb8+4yEisdMCKQh15u7VvkjqdkAP1BJ+HnSFyFTVrG5wSOxhnIFhWLq4g5J7CELSywKslvCz2ZzJWtQVwkfztq20p3hvRTnLBtw3WfvBv6IBgkiGcbqwkocig/BYuO/6Sm6V0oeD6O3IlXyaZqSZPhmZAoGAEZA5eI9HOYmJ324985st6voKawhx+pTWtbWXQ7HFqKlnN7qGfQDb44buMCEFUdVQMH0pGRr15wsV464cmGndtP68BDnBF2PTqy9xx9S2i6n3gfAqaQZCR6KCB090rK396OvYiG3ZIwl7omQ/0ydrR8l0w06B7F41Xl7szQehbDcCgYEApKCF1LDdrxgYS5agRRyes2nRsZOk3aQk3kHFli3NI/6PrEs5yeXprd/ls1UnE1SnI6BR+v+2cD97KBTTqAH6196/Yg2xNW2vwEXKPTHxLz/j226HXkv1p0uWdiyC3KfWprRRrCaIITbDklnkrGqudqDUKLQyeOHipA4bGLGygak=
-----END PRIVATE KEY-----`)

	block, _ := pem.Decode(privateKey)
	if block == nil {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaPriv, ok := priv.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("not an RSA private key")
	}

	hashed := sha256.Sum256([]byte(data))

	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPriv, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("error signing data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}
