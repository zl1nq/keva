package crypto

import (
	"crypto/rand"
	"io"
)

func RandomBytes(size int) ([]byte, error) {
	out := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, out); err != nil {
		return nil, err
	}
	return out, nil
}

func GenerateDEK() ([]byte, error) {
	return RandomBytes(KeySize)
}
