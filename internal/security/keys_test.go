package security

import "testing"

func TestAPIKeyRoundTrip(t *testing.T) {
	t.Setenv("AI_KEY_ENCRYPTION_KEY", "test-secret-that-is-not-used-in-production")
	encrypted, err := EncryptAPIKey("secret-api-key")
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "secret-api-key" {
		t.Fatal("API key was stored as plaintext")
	}
	plain, err := DecryptAPIKey(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "secret-api-key" {
		t.Fatalf("unexpected decrypted value %q", plain)
	}
}
