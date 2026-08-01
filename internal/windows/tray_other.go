//go:build !windows

package windows

type Tray struct{}

func StartTray(callbacks TrayCallbacks) *Tray {
	return &Tray{}
}

func (t *Tray) Close() {}
