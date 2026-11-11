package app

import (
	"errors"
	"fmt"
	"image"

	"github.com/bodrovis-learning/kruk-shot/internal/capture"
)

const maxCaptureAttempts = 2

// capture captures a display image, recreating its session if necessary.
func (m *monitor) capture() (*image.RGBA, error) {
	for attempt := 0; attempt < maxCaptureAttempts; attempt++ {
		if err := m.ensureSession(); err != nil {
			return nil, err
		}

		img, err := m.session.Capture()
		if !errors.Is(err, capture.ErrSessionLost) {
			return img, err
		}

		m.session.Close()
		m.session = nil

		if attempt+1 < maxCaptureAttempts {
			fmt.Printf(
				"DXGI session lost for %s, recreating...\n",
				m.display.Name,
			)
		}
	}

	return nil, fmt.Errorf(
		"DXGI session lost for %s after %d attempts: %w",
		m.display.Name,
		maxCaptureAttempts,
		capture.ErrSessionLost,
	)
}

// ensureSession creates a capture session if one does not exist.
func (m *monitor) ensureSession() error {
	if m.session != nil {
		return nil
	}

	session, err := capture.NewSession(m.display)
	if err != nil {
		return fmt.Errorf(
			"initialize capture session for %s: %w",
			m.display.Name,
			err,
		)
	}

	m.session = session
	return nil
}
