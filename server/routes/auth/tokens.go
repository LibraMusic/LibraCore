package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	rsaPrivateKey   *rsa.PrivateKey
	ecdsaPrivateKey *ecdsa.PrivateKey
	eddsaPrivateKey ed25519.PrivateKey
)

type TokenClaims struct {
	jwt.RegisteredClaims

	UserID string `json:"user_id"`
}

func GenerateToken(id string, expiration time.Duration, signingMethod, signingKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.GetSigningMethod(signingMethod), &TokenClaims{
		UserID:    id,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	})

	t, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return t, nil
}

func LoadPrivateKey(signingMethod, signingKey string) error {
	var err error

	switch signingMethod {
	case "RS256", "RS384", "RS512", "PS256", "PS384", "PS512":
		rsaPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(signingKey))
	case "ES256", "ES384", "ES512":
		ecdsaPrivateKey, err = jwt.ParseECPrivateKeyFromPEM([]byte(signingKey))
	case "EdDSA":
		var key crypto.PrivateKey
		key, err = jwt.ParseEdPrivateKeyFromPEM([]byte(signingKey))
		eddsaPrivateKey = key.(ed25519.PrivateKey)
	}
	return err
}

func SigningKey(signingMethod, configSigningKey string) any {
	var key any
	switch signingMethod {
	case "HS256", "HS384", "HS512":
		key = []byte(configSigningKey)
	case "RS256", "RS384", "RS512", "PS256", "PS384", "PS512":
		key = rsaPrivateKey.Public()
	case "ES256", "ES384", "ES512":
		key = ecdsaPrivateKey.Public()
	case "EdDSA":
		key = eddsaPrivateKey.Public()
	}
	return key
}

func CorrectSigningMethod(signingMethod string) string {
	signingMethods := []string{
		"HS256",
		"HS384",
		"HS512",
		"RS256",
		"RS384",
		"RS512",
		"PS256",
		"PS384",
		"PS512",
		"ES256",
		"ES384",
		"ES512",
		"EdDSA",
	}
	for _, method := range signingMethods {
		if strings.EqualFold(signingMethod, method) {
			return method
		}
	}
	return ""
}
