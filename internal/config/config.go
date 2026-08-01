package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var ErrConfigNotFound = errors.New("config file not found")

type Config struct {
	Salt             string       `json:"salt"`
	PasswordVerifier string       `json:"password_verifier"`
	EncryptedDEK     string       `json:"encrypted_dek"`
	CryptoVersion    int          `json:"crypto_version"`
	Argon2           Argon2Config `json:"argon2"`
	AutoLockMinutes  int          `json:"auto_lock_minutes"`
}

type Argon2Config struct {
	Memory      uint32 `json:"memory"`
	Iterations  uint32 `json:"iterations"`
	Parallelism uint8  `json:"parallelism"`
	KeyLength   uint32 `json:"key_length"`
}

func Default() Config {
	return Config{
		CryptoVersion:   1,
		AutoLockMinutes: 5,
		Argon2: Argon2Config{
			Memory:      65536,
			Iterations:  3,
			Parallelism: 4,
			KeyLength:   32,
		},
	}
}

func Exists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func Load(path string) (Config, error) {
	if !Exists(path) {
		return Config{}, ErrConfigNotFound
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	_ = os.Remove(path)
	return os.Rename(tmp, path)
}
