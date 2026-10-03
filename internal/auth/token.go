package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authToken := headers.Get("Authorization")
	if authToken == "" {
		return "", errors.New("authorization token doesn't exist")
	}
	tokenString := strings.Replace(authToken, "Bearer ", "", 1)

	return tokenString, nil
}

func MakeRefreshToken() string {
	someData := make([]byte, 256)
	rand.Read(someData)

	return hex.EncodeToString(someData)
}
