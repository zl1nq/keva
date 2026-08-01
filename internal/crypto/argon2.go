package crypto

import (
	"errors"

	"golang.org/x/crypto/argon2"
)

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
}

func DeriveKey(password, salt []byte, params Argon2Params) ([]byte, error) {
	if len(password) == 0 {
		return nil, errors.New("password is required")
	}
	if len(salt) == 0 {
		return nil, errors.New("salt is required")
	}
	if params.Memory == 0 || params.Iterations == 0 || params.Parallelism == 0 || params.KeyLength == 0 {
		return nil, errors.New("argon2 parameters are incomplete")
	}

	return argon2.IDKey(password, salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength), nil
}
