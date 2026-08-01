package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key, err := GenerateDEK()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("secret")
	sealed, err := Encrypt(key, plaintext)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Decrypt(key, sealed)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(got, plaintext) {
		t.Fatalf("decrypt mismatch: got %q want %q", got, plaintext)
	}
}

func TestEncryptUsesRandomNonce(t *testing.T) {
	key, err := GenerateDEK()
	if err != nil {
		t.Fatal(err)
	}

	first, err := Encrypt(key, []byte("same plaintext"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Encrypt(key, []byte("same plaintext"))
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Equal(first, second) {
		t.Fatal("expected different ciphertexts for repeated plaintext")
	}
}

func TestDecryptRejectsWrongKey(t *testing.T) {
	key, err := GenerateDEK()
	if err != nil {
		t.Fatal(err)
	}
	wrong, err := GenerateDEK()
	if err != nil {
		t.Fatal(err)
	}

	sealed, err := Encrypt(key, []byte("secret"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Decrypt(wrong, sealed); err == nil {
		t.Fatal("expected decrypt with wrong key to fail")
	}
}
