package encryption

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestCipherRoundTrip(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32)))
	cipher, err := New(key)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := cipher.Encrypt("telegram-secret")
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "telegram-secret" {
		t.Fatal("ciphertext must not equal plaintext")
	}
	decrypted, err := cipher.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "telegram-secret" {
		t.Fatalf("unexpected plaintext %q", decrypted)
	}
}
