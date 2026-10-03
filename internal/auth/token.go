package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetBearerToken(headers http.Header) (string, error) {
	authToken := headers.Get("Authorization")
	if authToken == "" {
		return "", errors.New("Authorization token doesn't exist")
	}
	tokenString := strings.Replace(authToken, "Bearer ", "", 1)

	return tokenString, nil
}
