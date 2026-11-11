package app

import (
	"errors"
	"fmt"

	"github.com/bodrovis-learning/kruk-shot/internal/output"
)

// captureAll captures screenshots from all active displays.
func (m *monitorManager) captureAll() error {
	if len(m.monitors) == 0 {
		return errors.New("no active displays")
	}

	writer, err := output.NewPNGWriter()
	if err != nil {
		return fmt.Errorf("initialize PNG writer: %w", err)
	}

	var failures []error

	for i := range m.monitors {
		if err := m.monitors[i].saveScreenshot(writer, i); err != nil {
			failures = append(failures, err)
		}
	}

	return errors.Join(failures...)
}

// saveScreenshot captures and saves an image from a single display.
func (m *monitor) saveScreenshot(writer *output.PNGWriter, index int) error {
	img, err := m.capture()
	if err != nil {
		return fmt.Errorf("capture %s: %w", m.display.Name, err)
	}

	path, err := writer.Save(index, img)
	if err != nil {
		return fmt.Errorf("save %s: %w", m.display.Name, err)
	}

	fmt.Printf(
		"Saved %s (%dx%d)\n",
		path,
		img.Bounds().Dx(),
		img.Bounds().Dy(),
	)

	return nil
}
