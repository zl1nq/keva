package app

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"keva/internal/config"
	"keva/internal/database"
	"keva/internal/vault"
)

func TestInitializeLockUnlock(t *testing.T) {
	testApp := New()
	paths := testPaths(t)
	if err := config.EnsurePortableDirs(paths); err != nil {
		t.Fatal(err)
	}
	testApp.paths = paths

	if testApp.IsInitialized() {
		t.Fatal("new temp vault should not be initialized")
	}

	if err := testApp.InitializeVault(InitializeVaultInput{MasterPassword: "correct horse battery staple"}); err != nil {
		t.Fatal(err)
	}

	if !testApp.IsInitialized() {
		t.Fatal("expected config file after initialization")
	}
	if state := testApp.GetLockState(); state.Locked {
		t.Fatal("expected vault to be unlocked after initialization")
	}

	cfg, err := config.Load(paths.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Salt == "" || cfg.PasswordVerifier == "" || cfg.EncryptedDEK == "" {
		t.Fatal("config is missing crypto fields")
	}

	if err := testApp.Lock(); err != nil {
		t.Fatal(err)
	}
	if state := testApp.GetLockState(); !state.Locked {
		t.Fatal("expected vault to be locked")
	}

	if err := testApp.Unlock(UnlockInput{MasterPassword: "wrong password"}); err == nil {
		t.Fatal("expected wrong password to fail")
	}
	if state := testApp.GetLockState(); !state.Locked {
		t.Fatal("wrong password must not unlock vault")
	}

	if err := testApp.Unlock(UnlockInput{MasterPassword: "correct horse battery staple"}); err != nil {
		t.Fatal(err)
	}
	if state := testApp.GetLockState(); state.Locked {
		t.Fatal("expected correct password to unlock vault")
	}
}

func TestUpdateSettingsRequiresInitializedVault(t *testing.T) {
	testApp := New()
	testApp.paths = testPaths(t)

	err := testApp.UpdateSettings(Settings{AutoLockMinutes: 3})
	if !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("got %v want %v", err, ErrNotInitialized)
	}
}

func TestAutoLockIfIdle(t *testing.T) {
	testApp := New()
	testApp.locked = false
	testApp.dek = []byte{1, 2, 3}
	testApp.autoLockMinutes = 1
	testApp.lastActivity = time.Now().Add(-2 * time.Minute)

	testApp.autoLockIfIdle()

	state := testApp.GetLockState()
	if !state.Locked {
		t.Fatal("expected idle vault to auto-lock")
	}
	for _, value := range testApp.dek {
		if value != 0 {
			t.Fatal("expected DEK to be cleared")
		}
	}
}

func TestAccountAccessRequiresUnlockedVault(t *testing.T) {
	testApp := New()
	paths := testPaths(t)
	if err := config.EnsurePortableDirs(paths); err != nil {
		t.Fatal(err)
	}
	db, err := database.Open(paths.Database)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	testApp.paths = paths
	testApp.db = db
	testApp.vault = vault.New(db)

	if _, err := testApp.CreateAccount(AccountInput{Title: "GitHub", Password: "secret"}); !errors.Is(err, ErrLocked) {
		t.Fatalf("got %v want %v", err, ErrLocked)
	}

	if err := testApp.InitializeVault(InitializeVaultInput{MasterPassword: "master password"}); err != nil {
		t.Fatal(err)
	}

	created, err := testApp.CreateAccount(AccountInput{
		Title:    "GitHub",
		Username: "octo",
		Password: "secret",
		URL:      "https://github.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	detail, err := testApp.GetAccount(created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Password != "secret" {
		t.Fatalf("got password %q want secret", detail.Password)
	}

	if err := testApp.Lock(); err != nil {
		t.Fatal(err)
	}
	if _, err := testApp.GetAccount(created.ID); !errors.Is(err, ErrLocked) {
		t.Fatalf("got %v want %v", err, ErrLocked)
	}
}

func testPaths(t *testing.T) config.Paths {
	t.Helper()

	root := t.TempDir()
	configDir := filepath.Join(root, "config")
	return config.Paths{
		Root:       root,
		ConfigDir:  configDir,
		ConfigFile: filepath.Join(configDir, "config.json"),
		DataDir:    filepath.Join(root, "data"),
		Database:   filepath.Join(root, "data", "keva.db"),
		LogDir:     filepath.Join(root, "logs"),
		Resources:  filepath.Join(root, "resources"),
		Libs:       filepath.Join(root, "libs"),
	}
}
