package config

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "config.json")
	cfg := Default()
	cfg.Salt = "salt"
	cfg.PasswordVerifier = "verifier"
	cfg.EncryptedDEK = "dek"
	cfg.AutoLockMinutes = 9

	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if got.Salt != cfg.Salt || got.PasswordVerifier != cfg.PasswordVerifier || got.EncryptedDEK != cfg.EncryptedDEK {
		t.Fatal("loaded config did not preserve crypto fields")
	}
	if got.AutoLockMinutes != 9 {
		t.Fatalf("got auto lock %d want 9", got.AutoLockMinutes)
	}
}
