package app

import (
	"fmt"

	"github.com/bodrovis-learning/kruk-shot/internal/capture"
)

type monitor struct {
	display capture.Display
	session *capture.Session
}

type monitorManager struct {
	monitors []monitor
}

// refresh synchronizes capture sessions with active displays.
func (m *monitorManager) refresh() error {
	displays, err := capture.ListDisplays()
	if err != nil {
		return fmt.Errorf("list displays: %w", err)
	}

	previous := m.monitors
	reused := make([]bool, len(previous))
	next := make([]monitor, len(displays))

	// Reuse sessions whose display configuration is unchanged.
	for i, display := range displays {
		next[i].display = display

		for j := range previous {
			if reused[j] || previous[j].display != display {
				continue
			}

			reused[j] = true
			next[i].session = previous[j].session
			break
		}
	}

	// Release sessions for disconnected or reconfigured displays.
	// Do this before creating new duplication interfaces.
	for i := range previous {
		if reused[i] {
			continue
		}

		if previous[i].session != nil {
			previous[i].session.Close()
		}

		fmt.Printf(
			"Display removed or reconfigured: %s\n",
			previous[i].display.Name,
		)
	}

	m.monitors = next

	// Initialize new displays and retry previously unavailable ones.
	for i := range m.monitors {
		mon := &m.monitors[i]

		if mon.session != nil {
			continue
		}

		session, err := capture.NewSession(mon.display)
		if err != nil {
			fmt.Printf(
				"Display unavailable: %s: %v\n",
				mon.display.Name,
				err,
			)
			continue
		}

		mon.session = session

		fmt.Printf(
			"Ready: %s (%dx%d at %d,%d)\n",
			mon.display.Name,
			mon.display.Width,
			mon.display.Height,
			mon.display.X,
			mon.display.Y,
		)
	}

	return nil
}

// close releases all active capture sessions.
func (m *monitorManager) close() {
	for i := range m.monitors {
		if m.monitors[i].session != nil {
			m.monitors[i].session.Close()
			m.monitors[i].session = nil
		}
	}

	m.monitors = nil
}
