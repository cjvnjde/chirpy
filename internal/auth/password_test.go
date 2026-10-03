package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	originalPassword := "strong_password"
	psw, err := HashPassword(originalPassword)
	if err != nil {
		t.Errorf("Hash failed: %v", err)
	}
	if psw == originalPassword {
		t.Error("Hash should not be the same as the password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	originalPassword := "strong_password"
	psw, err := HashPassword(originalPassword)
	if err != nil {
		t.Errorf("Hash failed: %v", err)
	}

	isValid, err := CheckPasswordHash(originalPassword, psw)
	if err != nil {
		t.Errorf("Hash check failed: %v", err)
	}
	if !isValid {
		t.Error("Password's hash doesn't match the password")
	}
}
