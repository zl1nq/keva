package app

import (
	"context"
	"encoding/base64"
	"errors"
	"path/filepath"
	"sync"
	"time"

	"keva/internal/clipboard"
	"keva/internal/config"
	vaultcrypto "keva/internal/crypto"
	"keva/internal/database"
	"keva/internal/logger"
	"keva/internal/security"
	"keva/internal/vault"
	"keva/internal/windows"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	ErrAlreadyInitialized = errors.New("vault is already initialized")
	ErrNotInitialized     = errors.New("vault is not initialized")
	ErrLocked             = errors.New("vault is locked")
)

// App is the Wails-facing application facade. It owns orchestration only; crypto,
// storage, and vault behavior live in internal packages.
type App struct {
	ctx             context.Context
	paths           config.Paths
	log             *logger.Logger
	db              *database.DB
	vault           *vault.Vault
	clip            *clipboard.Clipboard
	hotkey          *windows.Hotkey
	tray            *windows.Tray
	mu              sync.RWMutex
	locked          bool
	dek             []byte
	autoLockMinutes int
	lastActivity    time.Time
	autoLockStop    chan struct{}
	autoLockDone    chan struct{}
}

func New() *App {
	return &App{
		locked:          true,
		log:             logger.New(),
		clip:            clipboard.New(30 * time.Second),
		autoLockMinutes: DefaultSettings().AutoLockMinutes,
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	paths, err := config.ResolvePortablePaths()
	if err != nil {
		a.log.Error("resolve portable paths failed", err)
		return
	}
	if err := config.EnsurePortableDirs(paths); err != nil {
		a.log.Error("ensure portable dirs failed", err)
		return
	}

	db, err := database.Open(paths.Database)
	if err != nil {
		a.log.Error("open database failed", err)
		return
	}

	a.mu.Lock()
	a.paths = paths
	a.db = db
	a.vault = vault.New(db)
	a.locked = true
	a.autoLockMinutes = a.loadAutoLockMinutes()
	a.lastActivity = time.Now()
	a.mu.Unlock()

	a.startAutoLockLoop()
	if hotkey, err := windows.RegisterQuickSearchHotkey(a.handleQuickSearchHotkey); err != nil {
		a.log.Error("register quick search hotkey failed", err)
	} else {
		a.hotkey = hotkey
	}
	a.tray = windows.StartTray(windows.TrayCallbacks{
		IconPaths: trayIconPaths(paths),
		Open:      a.OpenWindow,
		Lock: func() {
			_ = a.Lock()
		},
		Quit: a.Quit,
	})
}

func (a *App) IsInitialized() bool {
	return config.Exists(a.paths.ConfigFile)
}

func (a *App) GetLockState() LockState {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return LockState{
		Initialized:     config.Exists(a.paths.ConfigFile),
		Locked:          a.locked,
		AutoLockMinutes: a.autoLockMinutes,
	}
}

func (a *App) InitializeVault(input InitializeVaultInput) error {
	if input.MasterPassword == "" {
		return errors.New("master password is required")
	}
	if a.IsInitialized() {
		return ErrAlreadyInitialized
	}

	cfg := config.Default()
	salt, err := vaultcrypto.RandomBytes(16)
	if err != nil {
		return err
	}

	password := []byte(input.MasterPassword)
	defer security.ZeroBytes(password)

	kek, err := vaultcrypto.DeriveKey(password, salt, argon2Params(cfg.Argon2))
	if err != nil {
		return err
	}
	defer security.ZeroBytes(kek)

	dek, err := vaultcrypto.GenerateDEK()
	if err != nil {
		return err
	}

	encryptedDEK, err := vaultcrypto.Encrypt(kek, dek)
	if err != nil {
		security.ZeroBytes(dek)
		return err
	}

	verifier, err := vaultcrypto.CreatePasswordVerifier(kek)
	if err != nil {
		security.ZeroBytes(dek)
		return err
	}

	cfg.Salt = base64.StdEncoding.EncodeToString(salt)
	cfg.EncryptedDEK = base64.StdEncoding.EncodeToString(encryptedDEK)
	cfg.PasswordVerifier = base64.StdEncoding.EncodeToString(verifier)

	if err := config.Save(a.paths.ConfigFile, cfg); err != nil {
		security.ZeroBytes(dek)
		return err
	}

	a.replaceDEK(dek)
	a.RecordActivity()
	return nil
}

func (a *App) Unlock(input UnlockInput) error {
	if input.MasterPassword == "" {
		return errors.New("master password is required")
	}
	cfg, err := config.Load(a.paths.ConfigFile)
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			return ErrNotInitialized
		}
		return err
	}

	salt, err := base64.StdEncoding.DecodeString(cfg.Salt)
	if err != nil {
		return err
	}
	verifier, err := base64.StdEncoding.DecodeString(cfg.PasswordVerifier)
	if err != nil {
		return err
	}
	encryptedDEK, err := base64.StdEncoding.DecodeString(cfg.EncryptedDEK)
	if err != nil {
		return err
	}

	password := []byte(input.MasterPassword)
	defer security.ZeroBytes(password)

	kek, err := vaultcrypto.DeriveKey(password, salt, argon2Params(cfg.Argon2))
	if err != nil {
		return err
	}
	defer security.ZeroBytes(kek)

	if err := vaultcrypto.VerifyPassword(kek, verifier); err != nil {
		return err
	}

	dek, err := vaultcrypto.Decrypt(kek, encryptedDEK)
	if err != nil {
		return errors.New("authentication failed")
	}

	a.replaceDEK(dek)
	a.RecordActivity()
	return nil
}

func (a *App) Lock() error {
	return a.lock("vault:locked")
}

func (a *App) lock(eventName string) error {
	a.mu.Lock()
	security.ZeroBytes(a.dek)
	a.dek = nil
	a.locked = true
	a.mu.Unlock()

	if a.ctx != nil && eventName != "" {
		runtime.EventsEmit(a.ctx, eventName)
	}
	return nil
}

func (a *App) CreateAccount(input AccountInput) (AccountSummary, error) {
	a.RecordActivity()
	dek, err := a.sessionDEK()
	if err != nil {
		return AccountSummary{}, err
	}
	defer security.ZeroBytes(dek)

	summary, err := a.vault.CreateAccount(a.context(), dek, toVaultInput(input))
	if err != nil {
		return AccountSummary{}, err
	}
	return fromVaultSummary(summary), nil
}

func (a *App) UpdateAccount(id string, input AccountInput) (AccountSummary, error) {
	a.RecordActivity()
	dek, err := a.sessionDEK()
	if err != nil {
		return AccountSummary{}, err
	}
	defer security.ZeroBytes(dek)

	summary, err := a.vault.UpdateAccount(a.context(), dek, id, toVaultInput(input))
	if err != nil {
		return AccountSummary{}, err
	}
	return fromVaultSummary(summary), nil
}

func (a *App) DeleteAccount(id string) error {
	a.RecordActivity()
	dek, err := a.sessionDEK()
	if err != nil {
		return err
	}
	defer security.ZeroBytes(dek)

	return a.vault.DeleteAccount(a.context(), id)
}

func (a *App) ListAccounts() ([]AccountSummary, error) {
	a.RecordActivity()
	dek, err := a.sessionDEK()
	if err != nil {
		return nil, err
	}
	defer security.ZeroBytes(dek)

	summaries, err := a.vault.ListAccounts(a.context(), dek)
	if err != nil {
		return nil, err
	}
	return fromVaultSummaries(summaries), nil
}

func (a *App) SearchAccounts(keyword string) ([]AccountSummary, error) {
	a.RecordActivity()
	dek, err := a.sessionDEK()
	if err != nil {
		return nil, err
	}
	defer security.ZeroBytes(dek)

	summaries, err := a.vault.SearchAccounts(a.context(), dek, keyword)
	if err != nil {
		return nil, err
	}
	return fromVaultSummaries(summaries), nil
}

func (a *App) GetAccount(id string) (AccountDetail, error) {
	a.RecordActivity()
	dek, err := a.sessionDEK()
	if err != nil {
		return AccountDetail{}, err
	}
	defer security.ZeroBytes(dek)

	detail, err := a.vault.GetAccount(a.context(), dek, id)
	if err != nil {
		return AccountDetail{}, err
	}
	return fromVaultDetail(detail), nil
}

func (a *App) GeneratePassword(options PasswordOptions) (string, error) {
	a.RecordActivity()
	return vault.GeneratePassword(vault.PasswordOptions{
		Length:           options.Length,
		IncludeUppercase: options.IncludeUppercase,
		IncludeLowercase: options.IncludeLowercase,
		IncludeNumbers:   options.IncludeNumbers,
		IncludeSymbols:   options.IncludeSymbols,
	})
}

func (a *App) CopyPassword(id string) error {
	a.RecordActivity()
	dek, err := a.sessionDEK()
	if err != nil {
		return err
	}
	defer security.ZeroBytes(dek)

	detail, err := a.vault.GetAccount(a.context(), dek, id)
	if err != nil {
		return err
	}
	return a.clip.CopyPassword(a.context(), detail.Password)
}

func (a *App) GetSettings() (Settings, error) {
	cfg, err := config.Load(a.paths.ConfigFile)
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			return DefaultSettings(), nil
		}
		return Settings{}, err
	}

	return Settings{
		AutoLockMinutes: cfg.AutoLockMinutes,
	}, nil
}

func (a *App) UpdateSettings(settings Settings) error {
	if settings.AutoLockMinutes <= 0 {
		return errors.New("auto lock minutes must be greater than zero")
	}

	cfg, err := config.Load(a.paths.ConfigFile)
	if err != nil {
		if errors.Is(err, config.ErrConfigNotFound) {
			return ErrNotInitialized
		}
		return err
	}
	cfg.AutoLockMinutes = settings.AutoLockMinutes
	if err := config.Save(a.paths.ConfigFile, cfg); err != nil {
		return err
	}

	a.mu.Lock()
	a.autoLockMinutes = settings.AutoLockMinutes
	a.lastActivity = time.Now()
	a.mu.Unlock()
	return nil
}

func (a *App) RecordActivity() {
	a.mu.Lock()
	a.lastActivity = time.Now()
	a.mu.Unlock()
}

func (a *App) OpenWindow() {
	if a.ctx == nil {
		return
	}
	runtime.WindowShow(a.ctx)
}

func (a *App) HideWindow() {
	if a.ctx == nil {
		return
	}
	runtime.WindowHide(a.ctx)
}

func (a *App) Quit() {
	if a.ctx == nil {
		return
	}
	runtime.Quit(a.ctx)
}

func (a *App) GetRuntimeStatus() RuntimeStatus {
	return RuntimeStatus{
		QuickSearchShortcut: "Ctrl+Shift+K",
		TrayAvailable:       true,
	}
}

func (a *App) BeforeClose(ctx context.Context) bool {
	_ = a.lock("vault:locked")
	runtime.WindowHide(ctx)
	return true
}

func (a *App) Shutdown(ctx context.Context) {
	_ = a.lock("")
	a.stopAutoLockLoop()
	if a.hotkey != nil {
		_ = a.hotkey.Close()
	}
	if a.tray != nil {
		a.tray.Close()
	}
	if a.db != nil {
		_ = a.db.Close()
	}
}

func (a *App) replaceDEK(dek []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()

	security.ZeroBytes(a.dek)
	a.dek = dek
	a.locked = false
	a.lastActivity = time.Now()
}

func (a *App) sessionDEK() ([]byte, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.locked || len(a.dek) == 0 {
		return nil, ErrLocked
	}
	if a.vault == nil {
		return nil, errors.New("vault store is unavailable")
	}

	dek := make([]byte, len(a.dek))
	copy(dek, a.dek)
	return dek, nil
}

func (a *App) context() context.Context {
	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

func (a *App) loadAutoLockMinutes() int {
	cfg, err := config.Load(a.paths.ConfigFile)
	if err != nil || cfg.AutoLockMinutes <= 0 {
		return DefaultSettings().AutoLockMinutes
	}
	return cfg.AutoLockMinutes
}

func (a *App) startAutoLockLoop() {
	a.stopAutoLockLoop()

	a.autoLockStop = make(chan struct{})
	a.autoLockDone = make(chan struct{})
	go func() {
		defer close(a.autoLockDone)
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				a.autoLockIfIdle()
			case <-a.autoLockStop:
				return
			}
		}
	}()
}

func (a *App) stopAutoLockLoop() {
	if a.autoLockStop == nil {
		return
	}
	close(a.autoLockStop)
	if a.autoLockDone != nil {
		<-a.autoLockDone
	}
	a.autoLockStop = nil
	a.autoLockDone = nil
}

func (a *App) autoLockIfIdle() {
	a.mu.RLock()
	locked := a.locked
	minutes := a.autoLockMinutes
	lastActivity := a.lastActivity
	a.mu.RUnlock()

	if locked || minutes <= 0 {
		return
	}
	if time.Since(lastActivity) >= time.Duration(minutes)*time.Minute {
		_ = a.lock("vault:auto-locked")
	}
}

func (a *App) handleQuickSearchHotkey() {
	if a.ctx == nil {
		return
	}
	runtime.WindowShow(a.ctx)
	runtime.EventsEmit(a.ctx, "shortcut:quick-search")
}

func trayIconPaths(paths config.Paths) []string {
	return []string{
		filepath.Join(paths.Resources, "icon.ico"),
		filepath.Join(paths.Resources, "assets", "appicon.png"),
		filepath.Join(paths.Root, "icon.ico"),
		filepath.Join(paths.Root, "appicon.png"),
		filepath.Clean(filepath.Join(paths.Root, "..", "..", "icon.ico")),
		filepath.Clean(filepath.Join(paths.Root, "..", "..", "appicon.png")),
	}
}

func argon2Params(cfg config.Argon2Config) vaultcrypto.Argon2Params {
	return vaultcrypto.Argon2Params{
		Memory:      cfg.Memory,
		Iterations:  cfg.Iterations,
		Parallelism: cfg.Parallelism,
		KeyLength:   cfg.KeyLength,
	}
}

func toVaultInput(input AccountInput) vault.AccountInput {
	return vault.AccountInput{
		Title:    input.Title,
		Username: input.Username,
		Password: input.Password,
		URL:      input.URL,
		Note:     input.Note,
	}
}

func fromVaultSummary(summary vault.AccountSummary) AccountSummary {
	return AccountSummary{
		ID:        summary.ID,
		Title:     summary.Title,
		Username:  summary.Username,
		URL:       summary.URL,
		CreatedAt: summary.CreatedAt,
		UpdatedAt: summary.UpdatedAt,
	}
}

func fromVaultSummaries(summaries []vault.AccountSummary) []AccountSummary {
	out := make([]AccountSummary, 0, len(summaries))
	for _, summary := range summaries {
		out = append(out, fromVaultSummary(summary))
	}
	return out
}

func fromVaultDetail(detail vault.AccountDetail) AccountDetail {
	return AccountDetail{
		ID:        detail.ID,
		Title:     detail.Title,
		Username:  detail.Username,
		Password:  detail.Password,
		URL:       detail.URL,
		Note:      detail.Note,
		CreatedAt: detail.CreatedAt,
		UpdatedAt: detail.UpdatedAt,
	}
}
