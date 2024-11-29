package utils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var RSAPrivateKey *rsa.PrivateKey
var ECDSAPrivateKey *ecdsa.PrivateKey
var EdDSAPrivateKey ed25519.PrivateKey

func GetCorrectSigningMethod(signingMethod string) string {
	signingMethods := []string{"HS256", "HS384", "HS512", "RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "ES256", "ES384", "ES512", "EdDSA"}
	for _, method := range signingMethods {
		if strings.EqualFold(signingMethod, method) {
			return method
		}
	}
	return ""
}

func GenerateToken(id string, expiration time.Duration, signingMethod string, signingKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.GetSigningMethod(signingMethod), jwt.MapClaims{
		"user_id": id,
		"exp":     time.Now().Add(expiration).Unix(),
	})

	t, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return t, nil
}

func LoadPrivateKey(signingMethod string, signingKey string) error {
	var err error

	switch signingMethod {
	case "RS256", "RS384", "RS512", "PS256", "PS384", "PS512":
		RSAPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(signingKey))
	case "ES256", "ES384", "ES512":
		ECDSAPrivateKey, err = jwt.ParseECPrivateKeyFromPEM([]byte(signingKey))
	case "EdDSA":
		var key crypto.PrivateKey
		key, err = jwt.ParseEdPrivateKeyFromPEM([]byte(signingKey))
		EdDSAPrivateKey = key.(ed25519.PrivateKey)
	}
	return err
}
