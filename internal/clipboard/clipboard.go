package clipboard

import (
	"context"
	"crypto/sha256"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Clipboard struct {
	clearAfter time.Duration
	sequence   atomic.Uint64
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

	checksum := sha256.Sum256([]byte(password))
	sequence := c.sequence.Add(1)

	time.AfterFunc(clearAfter, func() {
		if !c.isLatest(sequence) {
			return
		}
		current, err := runtime.ClipboardGetText(ctx)
		if err != nil || sha256.Sum256([]byte(current)) != checksum || !c.isLatest(sequence) {
			return
		}
		_ = runtime.ClipboardSetText(ctx, "")
	})
	return nil
}

func (c *Clipboard) isLatest(sequence uint64) bool {
	return c.sequence.Load() == sequence
}
