//go:build windows

package windows

import (
	"sync"
	"syscall"
	"unsafe"
)

const (
	modControl = 0x0002
	modShift   = 0x0004
	vkK        = 0x4B
	wmHotKey   = 0x0312
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procPostThreadMsgW   = user32.NewProc("PostThreadMessageW")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetCurrentThread = kernel32.NewProc("GetCurrentThreadId")
)

type Hotkey struct {
	mu       sync.Mutex
	id       int
	threadID uint32
	done     chan struct{}
}

func RegisterQuickSearchHotkey(onPressed func()) (*Hotkey, error) {
	hotkey := &Hotkey{
		id:   1,
		done: make(chan struct{}),
	}

	ready := make(chan error, 1)
	go hotkey.loop(onPressed, ready)

	if err := <-ready; err != nil {
		return nil, err
	}
	return hotkey, nil
}

func (h *Hotkey) Close() error {
	h.mu.Lock()
	threadID := h.threadID
	h.mu.Unlock()

	if threadID != 0 {
		procPostThreadMsgW.Call(uintptr(threadID), uintptr(0x0012), 0, 0)
	}
	<-h.done
	return nil
}

func (h *Hotkey) loop(onPressed func(), ready chan<- error) {
	threadID, _, _ := procGetCurrentThread.Call()
	h.mu.Lock()
	h.threadID = uint32(threadID)
	h.mu.Unlock()

	registered, _, err := procRegisterHotKey.Call(0, uintptr(h.id), uintptr(modControl|modShift), uintptr(vkK))
	if registered == 0 {
		ready <- err
		close(h.done)
		return
	}
	defer func() {
		procUnregisterHotKey.Call(0, uintptr(h.id))
		close(h.done)
	}()

	ready <- nil

	var msg msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			return
		}
		if msg.Message == wmHotKey && onPressed != nil {
			onPressed()
		}
	}
}

type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type point struct {
	X int32
	Y int32
}
