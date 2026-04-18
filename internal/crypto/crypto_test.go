package crypto_test

import (
	"bytes"
	"testing"

	"github.com/keyman/keyman/internal/crypto"
)

func TestNewSalt(t *testing.T) {
	s1, err := crypto.NewSalt()
	if err != nil {
		t.Fatalf("NewSalt() error: %v", err)
	}
	if len(s1) != 32 {
		t.Fatalf("expected 32-byte salt, got %d", len(s1))
	}

	s2, _ := crypto.NewSalt()
	if bytes.Equal(s1, s2) {
		t.Fatal("two salts should not be equal")
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := crypto.DeriveKey("hunter2", []byte("testsalt12345678testsalt12345678"))

	plaintext := []byte("super secret api key")
	ciphertext, err := crypto.Encrypt(key, plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error: %v", err)
	}

	got, err := crypto.Decrypt(key, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error: %v", err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Fatalf("roundtrip mismatch: got %q, want %q", got, plaintext)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	key1 := crypto.DeriveKey("correct", []byte("testsalt12345678testsalt12345678"))
	key2 := crypto.DeriveKey("incorrect", []byte("testsalt12345678testsalt12345678"))

	ct, _ := crypto.Encrypt(key1, []byte("data"))
	_, err := crypto.Decrypt(key2, ct)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestDecryptCorrupted(t *testing.T) {
	key := crypto.DeriveKey("pass", []byte("testsalt12345678testsalt12345678"))
	ct, _ := crypto.Encrypt(key, []byte("data"))

	// Flip a byte in the ciphertext body
	ct[len(ct)-1] ^= 0xFF

	_, err := crypto.Decrypt(key, ct)
	if err == nil {
		t.Fatal("expected error on corrupted ciphertext")
	}
}

func TestDecryptTooShort(t *testing.T) {
	key := crypto.DeriveKey("pass", []byte("testsalt12345678testsalt12345678"))
	_, err := crypto.Decrypt(key, []byte{0x01, 0x02})
	if err == nil {
		t.Fatal("expected error on too-short input")
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	salt := []byte("testsalt12345678testsalt12345678")
	k1 := crypto.DeriveKey("password", salt)
	k2 := crypto.DeriveKey("password", salt)
	if !bytes.Equal(k1, k2) {
		t.Fatal("same password+salt should produce the same key")
	}
}

func TestDeriveKeyDifferentSalts(t *testing.T) {
	k1 := crypto.DeriveKey("password", []byte("salt1___salt1___salt1___salt1___"))
	k2 := crypto.DeriveKey("password", []byte("salt2___salt2___salt2___salt2___"))
	if bytes.Equal(k1, k2) {
		t.Fatal("different salts should produce different keys")
	}
}
