package capture

import (
	"fmt"

	"golang.org/x/sys/windows"
)

var setProcessDPIAwarenessContext = windows.NewLazySystemDLL("user32.dll").
	NewProc("SetProcessDpiAwarenessContext")

func EnableDPIAwareness() error {
	// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4
	const perMonitorAwareV2 = ^uintptr(3)

	result, _, err := setProcessDPIAwarenessContext.Call(perMonitorAwareV2)
	if result == 0 {
		return fmt.Errorf("SetProcessDpiAwarenessContext: %w", err)
	}

	return nil
}
