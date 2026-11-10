package hotkey

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"unsafe"
)

// Run registers Ctrl+Shift+F9 and processes hotkey events.
// It blocks until Ctrl+C is received.
func Run(onPress func()) error {
	// Windows hotkeys belong to a specific OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	threadID, _, _ := getCurrentThreadID.Call()

	// Ensure that this thread has a Windows message queue.
	var msg message

	peekMessageW.Call(
		uintptr(unsafe.Pointer(&msg)),
		0,
		wmUser,
		wmUser,
		0, // PM_NOREMOVE
	)

	result, _, err := registerHotKey.Call(
		0, // No window; use the current thread.
		hotkeyID,
		modControl|modShift|modNoRepeat,
		vkF9,
	)

	if result == 0 {
		return fmt.Errorf("register Ctrl+Shift+F9: %w", err)
	}

	defer unregisterHotKey.Call(0, hotkeyID)

	// Ctrl+C must wake the Windows message loop.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	defer signal.Stop(signals)

	done := make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-signals:
			result, _, err := postThreadMessageW.Call(
				threadID,
				wmQuit,
				0,
				0,
			)

			if result == 0 {
				fmt.Fprintf(os.Stderr, "PostThreadMessageW: %v\n", err)
			}

		case <-done:
			return
		}
	}()

	for {
		var msg message

		result, _, err := getMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0,
			0,
			0,
		)

		switch int32(result) {
		case -1:
			return fmt.Errorf("GetMessageW: %w", err)

		case 0:
			// WM_QUIT
			return nil
		}

		if msg.Message == wmHotkey && msg.WParam == hotkeyID {
			onPress()
		}
	}
}
