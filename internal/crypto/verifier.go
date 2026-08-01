package crypto

import (
	"bytes"
	"errors"
)

var verifierPlaintext = []byte("KEVA password verifier v1")

func CreatePasswordVerifier(kek []byte) ([]byte, error) {
	return Encrypt(kek, verifierPlaintext)
}

func VerifyPassword(kek, verifier []byte) error {
	plaintext, err := Decrypt(kek, verifier)
	if err != nil {
		return errors.New("authentication failed")
	}
	if !bytes.Equal(plaintext, verifierPlaintext) {
		return errors.New("authentication failed")
	}
	return nil
}
