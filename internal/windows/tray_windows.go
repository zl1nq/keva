//go:build windows

package windows

import (
	"sync"

	"github.com/getlantern/systray"
)

type Tray struct {
	ready chan struct{}
	once  sync.Once
}

func StartTray(callbacks TrayCallbacks) *Tray {
	tray := &Tray{ready: make(chan struct{})}
	go systray.Run(func() {
		systray.SetTitle("KEVA")
		systray.SetTooltip("KEVA local password vault")

		open := systray.AddMenuItem("Open KEVA", "Show KEVA")
		lock := systray.AddMenuItem("Lock", "Lock the vault")
		quit := systray.AddMenuItem("Exit", "Exit KEVA")

		close(tray.ready)

		go func() {
			for {
				select {
				case <-open.ClickedCh:
					if callbacks.Open != nil {
						callbacks.Open()
					}
				case <-lock.ClickedCh:
					if callbacks.Lock != nil {
						callbacks.Lock()
					}
				case <-quit.ClickedCh:
					if callbacks.Quit != nil {
						callbacks.Quit()
					}
					systray.Quit()
					return
				}
			}
		}()
	}, func() {})
	return tray
}

func (t *Tray) Close() {
	t.once.Do(func() {
		systray.Quit()
	})
}
