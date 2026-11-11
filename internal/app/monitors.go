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

	previous := m.existingSessions()
	next := make([]monitor, 0, len(displays))

	for _, display := range displays {
		next = append(next, monitor{
			display: display,
			session: previous[display],
		})

		delete(previous, display)
	}

	// Release removed or reconfigured displays before
	// creating new duplication interfaces.
	closeRemoved(previous)

	m.monitors = next
	m.initializeMissing()

	return nil
}

// existingSessions indexes the current sessions by display configuration.
func (m *monitorManager) existingSessions() map[capture.Display]*capture.Session {
	sessions := make(map[capture.Display]*capture.Session, len(m.monitors))

	for _, mon := range m.monitors {
		sessions[mon.display] = mon.session
	}

	return sessions
}

// closeRemoved releases sessions that are no longer reusable.
func closeRemoved(sessions map[capture.Display]*capture.Session) {
	for display, session := range sessions {
		if session != nil {
			session.Close()
		}

		fmt.Printf(
			"Display removed or reconfigured: %s\n",
			display.Name,
		)
	}
}

// initializeMissing creates sessions for new or unavailable displays.
func (m *monitorManager) initializeMissing() {
	for i := range m.monitors {
		mon := &m.monitors[i]

		if mon.session != nil {
			continue
		}

		if err := mon.ensureSession(); err != nil {
			fmt.Printf("Display unavailable: %v\n", err)
			continue
		}

		fmt.Printf(
			"Ready: %s (%dx%d at %d,%d)\n",
			mon.display.Name,
			mon.display.Width,
			mon.display.Height,
			mon.display.X,
			mon.display.Y,
		)
	}
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
