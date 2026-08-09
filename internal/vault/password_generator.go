package vault

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
)

type PasswordOptions struct {
	Length           int  `json:"length"`
	IncludeUppercase bool `json:"include_uppercase"`
	IncludeLowercase bool `json:"include_lowercase"`
	IncludeNumbers   bool `json:"include_numbers"`
	IncludeSymbols   bool `json:"include_symbols"`
}

func DefaultPasswordOptions() PasswordOptions {
	return PasswordOptions{
		Length:           20,
		IncludeUppercase: true,
		IncludeLowercase: true,
		IncludeNumbers:   true,
		IncludeSymbols:   true,
	}
}

func GeneratePassword(options PasswordOptions) (string, error) {
	if options.Length <= 0 {
		options.Length = DefaultPasswordOptions().Length
	}
	if options.Length < 8 || options.Length > 128 {
		return "", errors.New("password length must be between 8 and 128")
	}

	charsets := selectedCharsets(options)
	if len(charsets) == 0 {
		return "", errors.New("at least one character set is required")
	}

	alphabet := strings.Join(charsets, "")
	out := make([]byte, options.Length)
	for i := 0; i < options.Length; i++ {
		index, err := randomIndex(len(alphabet))
		if err != nil {
			return "", err
		}
		out[i] = alphabet[index]
	}

	for i, charset := range charsets {
		index, err := randomIndex(len(charset))
		if err != nil {
			return "", err
		}
		out[i] = charset[index]
	}

	for i := len(out) - 1; i > 0; i-- {
		j, err := randomIndex(i + 1)
		if err != nil {
			return "", err
		}
		out[i], out[j] = out[j], out[i]
	}

	return string(out), nil
}

func randomIndex(length int) (int, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
	if err != nil {
		return 0, err
	}
	return int(index.Int64()), nil
}

func selectedCharsets(options PasswordOptions) []string {
	charsets := make([]string, 0, 4)
	if options.IncludeUppercase {
		charsets = append(charsets, "ABCDEFGHJKLMNPQRSTUVWXYZ")
	}
	if options.IncludeLowercase {
		charsets = append(charsets, "abcdefghijkmnopqrstuvwxyz")
	}
	if options.IncludeNumbers {
		charsets = append(charsets, "23456789")
	}
	if options.IncludeSymbols {
		charsets = append(charsets, "!@#$%^&*()-_=+[]{};:,.?")
	}
	return charsets
}
