package crypto

import "testing"

func TestPasswordHashingAndVerification(t *testing.T) {
	rawPassword := "SecureP@ssw0rd!123"

	hash, err := HashPassword(rawPassword)
	if err != nil {
		t.Fatalf("expected no error hashing password, got %v", err)
	}

	if hash == rawPassword {
		t.Fatalf("hash should not match raw password")
	}

	if !CheckPassword(rawPassword, hash) {
		t.Errorf("expected CheckPassword to return true for correct password")
	}

	if CheckPassword("WrongPassword", hash) {
		t.Errorf("expected CheckPassword to return false for wrong password")
	}
}
