package app

type InitializeVaultInput struct {
	MasterPassword string `json:"master_password"`
}

type UnlockInput struct {
	MasterPassword string `json:"master_password"`
}

type LockState struct {
	Initialized     bool `json:"initialized"`
	Locked          bool `json:"locked"`
	AutoLockMinutes int  `json:"auto_lock_minutes"`
}

type Settings struct {
	AutoLockMinutes int `json:"auto_lock_minutes"`
}

type AccountInput struct {
	Title    string `json:"title"`
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Note     string `json:"note"`
}

type AccountSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	URL       string `json:"url"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type AccountDetail struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	URL       string `json:"url"`
	Note      string `json:"note"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type PasswordOptions struct {
	Length           int  `json:"length"`
	IncludeUppercase bool `json:"include_uppercase"`
	IncludeLowercase bool `json:"include_lowercase"`
	IncludeNumbers   bool `json:"include_numbers"`
	IncludeSymbols   bool `json:"include_symbols"`
}

type RuntimeStatus struct {
	QuickSearchShortcut string `json:"quick_search_shortcut"`
	TrayAvailable       bool   `json:"tray_available"`
}

func DefaultSettings() Settings {
	return Settings{
		AutoLockMinutes: 5,
	}
}
