package hotkey

import (
	"fmt"
	"unsafe"
)

// ensureMessageQueue forces Windows to create a message queue
// for the current thread before other threads can post to it.
func ensureMessageQueue() {
	var msg message

	// PM_NOREMOVE (0) means that matching messages are not removed.
	// Returning zero is normal when the queue contains no messages.
	_, _, _ = peekMessageW.Call(
		uintptr(unsafe.Pointer(&msg)),
		0,
		wmUser,
		wmUser,
		0, // PM_NOREMOVE
	)
}

// runMessageLoop waits for thread messages and invokes the callback
// when the registered screenshot hotkey is pressed.
func runMessageLoop(onPress func()) error {
	for {
		var msg message

		result, _, err := getMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, // Receive window and thread messages.
			0, // No minimum message filter.
			0, // No maximum message filter.
		)

		switch int32(result) {
		case -1:
			// GetMessageW failed.
			return fmt.Errorf("GetMessageW: %w", err)

		case 0:
			// WM_QUIT received. Exit the message loop.
			return nil
		}

		if msg.Message == wmHotkey && msg.WParam == hotkeyID {
			onPress()
		}
	}
}
