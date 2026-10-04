package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

type authorizationType int

const (
	Bearer authorizationType = iota
	Api
)

func getAuthorization(authType authorizationType, headers http.Header) (string, error) {
	authToken := headers.Get("Authorization")
	if authToken == "" {
		return "", errors.New("authorization token doesn't exist")
	}
	prefix := ""
	switch authType {
	case Bearer:
		prefix = "Bearer "
	case Api:
		prefix = "ApiKey "
	}

	tokenString := strings.Replace(authToken, prefix, "", 1)
	return tokenString, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	return getAuthorization(Bearer, headers)
}

func MakeRefreshToken() string {
	someData := make([]byte, 256)
	rand.Read(someData)

	return hex.EncodeToString(someData)
}

func GetAPIKey(headers http.Header) (string, error) {
	return getAuthorization(Api, headers)
}
