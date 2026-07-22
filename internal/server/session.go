package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

const sessionCookieName = "vpsfm_session"

func signSessionValue(secret, username string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(username))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	user := base64.RawURLEncoding.EncodeToString([]byte(username))
	return user + "." + signature
}

func verifySessionValue(secret, value string) (string, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return "", false
	}
	decodedUser, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	expected := signSessionValue(secret, string(decodedUser))
	if !hmac.Equal([]byte(expected), []byte(value)) {
		return "", false
	}
	return string(decodedUser), true
}
