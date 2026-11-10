package hotkey

import "golang.org/x/sys/windows"

const (
	hotkeyID = 1

	modControl  = 0x0002
	modShift    = 0x0004
	modNoRepeat = 0x4000

	vkF9 = 0x78

	wmHotkey = 0x0312
	wmQuit   = 0x0012
	wmUser   = 0x0400
)

type point struct {
	X int32
	Y int32
}

type message struct {
	HWND     uintptr
	Message  uint32
	WParam   uintptr
	LParam   uintptr
	Time     uint32
	Point    point
	LPrivate uint32
}

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
