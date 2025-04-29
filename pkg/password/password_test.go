package password

import "testing"

func TestHashAndSaltPassword(t *testing.T) {
	password := "mySecret123"

	hashed, err := HashAndSaltPassword(password)
	if err != nil {
		t.Fatalf("Hashing failed: %v", err)
	}

	if hashed == "" {
		t.Fatal("Expected non-empty hash")
	}

	if !VerifyPassword(password, hashed) {
		t.Fatal("Password should match the hash")
	}

	if VerifyPassword("wrongPassword", hashed) {
		t.Fatal("Wrong password should not match the hash")
	}
}
