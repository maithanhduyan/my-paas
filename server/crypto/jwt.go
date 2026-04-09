package crypto

import (
	"encoding/base64"
	"encoding/json"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// JWTClaims represents the payload of a JWT token.
type JWTClaims struct {
	Sub      string `json:"sub"`               // User ID
	Username string `json:"username"`
	Role     string `json:"role"`
	Type     string `json:"type"`              // "access" | "refresh"
	Iss      string `json:"iss"`               // "mypaas"
	Iat      int64  `json:"iat"`               // Issued at
	Exp      int64  `json:"exp"`               // Expires at
}

// GenerateJWT creates a signed JWT token using HMAC-SHA256.
func GenerateJWT(userID, username, role, tokenType, secret string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		Sub:      userID,
		Username: username,
		Role:     role,
		Type:     tokenType,
		Iss:      "mypaas",
		Iat:      now.Unix(),
		Exp:      now.Add(expiry).Unix(),
	}

	header := base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64URLEncode(payload)

	signingInput := header + "." + encodedPayload
	signature := signHMAC(signingInput, secret)

	return signingInput + "." + signature, nil
}

// ValidateJWT validates a JWT token and returns the claims.
func ValidateJWT(token, secret string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := signHMAC(signingInput, secret)

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid token signature")
	}

	payload, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal claims: %w", err)
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func signHMAC(input, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(input))
	return base64URLEncode(mac.Sum(nil))
}

func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64URLDecode(s string) ([]byte, error) {
	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
