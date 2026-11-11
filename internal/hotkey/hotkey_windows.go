package hotkey

import (
	"fmt"
	"os"
	"runtime"
)

// Run registers Ctrl+Shift+F9 and processes hotkey events.
// It blocks until Ctrl+C is received.
func Run(onPress func()) error {
	// RegisterHotKey is bound to the calling Windows thread.
	// All related operations must run on that same thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ensureMessageQueue()

	if err := registerScreenshotHotkey(); err != nil {
		return err
	}
	defer unregisterScreenshotHotkey()

	threadID, _, _ := getCurrentThreadID.Call()

	// Ctrl+C interrupts the blocking GetMessageW call.
	stopInterrupt := watchInterrupt(threadID)
	defer stopInterrupt()

	return runMessageLoop(onPress)
}

// registerScreenshotHotkey registers a thread-level global hotkey.
func registerScreenshotHotkey() error {
	result, _, err := registerHotKey.Call(
		0, // NULL HWND: associate the hotkey with this thread.
		hotkeyID,
		modControl|modShift|modNoRepeat,
		vkF9,
	)

	if result == 0 {
		return fmt.Errorf("register Ctrl+Shift+F9: %w", err)
	}

	return nil
}

// unregisterScreenshotHotkey removes the hotkey registration.
func unregisterScreenshotHotkey() {
	result, _, err := unregisterHotKey.Call(0, hotkeyID)
	if result == 0 {
		fmt.Fprintf(os.Stderr, "UnregisterHotKey failed: %v\n", err)
	}
}
