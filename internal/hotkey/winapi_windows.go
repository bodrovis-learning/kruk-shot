package hotkey

import "golang.org/x/sys/windows"

const (
	// Application-defined hotkey identifier.
	// Used to distinguish our hotkey from other registered hotkeys.
	hotkeyID = 1

	// RegisterHotKey modifier flags (winuser.h).
	modControl  = 0x0002 // MOD_CONTROL: Ctrl must be pressed.
	modShift    = 0x0004 // MOD_SHIFT: Shift must be pressed.
	modNoRepeat = 0x4000 // MOD_NOREPEAT: suppress keyboard auto-repeat.

	// Windows virtual-key code (winuser.h).
	vkF9 = 0x78 // VK_F9: F9 key.

	// Windows message identifiers (winuser.h).
	wmHotkey = 0x0312 // WM_HOTKEY: a registered hotkey was pressed.
	wmQuit   = 0x0012 // WM_QUIT: terminate the message loop.
	wmUser   = 0x0400 // WM_USER: base value for application-defined messages.
)

// point mirrors the Win32 POINT structure.
// Coordinates are expressed in screen pixels.
type point struct {
	X int32
	Y int32
}

// message mirrors the Win32 MSG structure.
//
// Field order and native types must match the Windows ABI.
// GetMessageW and PeekMessageW write directly into this structure.
type message struct {
	HWND     uintptr // HWND: target window handle (zero for thread messages).
	Message  uint32  // UINT: message identifier, such as WM_HOTKEY.
	WParam   uintptr // WPARAM: additional message data (hotkey ID for WM_HOTKEY).
	LParam   uintptr // LPARAM: additional message data (modifiers and key code).
	Time     uint32  // DWORD: time when the message was posted.
	Point    point   // POINT: cursor position when the message was posted.
	LPrivate uint32  // DWORD: reserved for internal system use.
}

// Windows system procedures loaded at runtime.
//
// user32.dll handles hotkey registration and message queues.
// kernel32.dll provides the current OS thread identifier.
var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	registerHotKey     = user32.NewProc("RegisterHotKey")
	unregisterHotKey   = user32.NewProc("UnregisterHotKey")
	getMessageW        = user32.NewProc("GetMessageW")
	peekMessageW       = user32.NewProc("PeekMessageW")
	postThreadMessageW = user32.NewProc("PostThreadMessageW")

	getCurrentThreadID = windows.NewLazySystemDLL("kernel32.dll").
				NewProc("GetCurrentThreadId")
)
