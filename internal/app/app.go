package app

import (
	"fmt"

	"github.com/bodrovis-learning/kruk-shot/internal/buildinfo"
	"github.com/bodrovis-learning/kruk-shot/internal/hotkey"
)

func Run() error {
	manager := &monitorManager{}
	defer manager.close()

	if err := manager.refresh(); err != nil {
		return err
	}

	fmt.Printf("Kruk Shot %s running\n", buildinfo.Version)
	fmt.Println("Ctrl+Shift+F9: screenshot")
	fmt.Println("Ctrl+C: exit")

	return hotkey.Run(func() {
		// Detect connected, disconnected and reconfigured displays.
		if err := manager.refresh(); err != nil {
			fmt.Printf("Monitor refresh failed: %v\n", err)
			return
		}

		if err := manager.captureAll(); err != nil {
			fmt.Printf("Capture failed: %v\n", err)
		}
	})
}
