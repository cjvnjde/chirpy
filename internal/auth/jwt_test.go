package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userUUID := uuid.New()
	tokenSecret := "token_secret"
	duration, err := time.ParseDuration("5s")
	if err != nil {
		t.Errorf("Duration has failed: %v", err)
	}

	jwtStr, err := MakeJWT(userUUID, tokenSecret, duration)
	if err != nil {
		t.Errorf("MakeJWT has failed: %v", err)
	}
	if len(jwtStr) < 20 {
		t.Error("JWT string is too short")
	}
}

func TestValidateJWT(t *testing.T) {
	userUUID := uuid.New()
	tokenSecret := "token_secret"
	duration, err := time.ParseDuration("5s")
	if err != nil {
		t.Errorf("Duration has failed: %v", err)
	}

	jwtStr, err := MakeJWT(userUUID, tokenSecret, duration)
	d, err := ValidateJWT(jwtStr, tokenSecret)
	if err != nil {
		t.Errorf("Validation has failed: %v", err)
	}

	if userUUID.String() != d.String() {
		t.Error("User UUID form jwt does not match the original one")
	}
}
