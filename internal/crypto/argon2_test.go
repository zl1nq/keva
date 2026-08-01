package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKeyStableForSameInputs(t *testing.T) {
	params := Argon2Params{
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 1,
		KeyLength:   32,
	}

	first, err := DeriveKey([]byte("master"), []byte("salt"), params)
	if err != nil {
		t.Fatal(err)
	}
	second, err := DeriveKey([]byte("master"), []byte("salt"), params)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(first, second) {
		t.Fatal("expected stable derived key")
	}
	if len(first) != KeySize {
		t.Fatalf("got key length %d want %d", len(first), KeySize)
	}
}
