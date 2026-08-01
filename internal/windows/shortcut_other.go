//go:build !windows

package windows

type Hotkey struct{}

func RegisterQuickSearchHotkey(onPressed func()) (*Hotkey, error) {
	return &Hotkey{}, nil
}

func (h *Hotkey) Close() error {
	return nil
}
