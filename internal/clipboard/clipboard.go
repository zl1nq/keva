package clipboard

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Clipboard struct {
	clearAfter time.Duration
}

func New(clearAfter time.Duration) *Clipboard {
	return &Clipboard{clearAfter: clearAfter}
}

func (c *Clipboard) CopyPassword(ctx context.Context, password string) error {
	if err := runtime.ClipboardSetText(ctx, password); err != nil {
		return err
	}

	clearAfter := c.clearAfter
	if clearAfter <= 0 {
		clearAfter = 30 * time.Second
	}

	time.AfterFunc(clearAfter, func() {
		_ = runtime.ClipboardSetText(ctx, "")
	})
	return nil
}
