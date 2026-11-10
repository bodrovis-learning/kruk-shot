package app

import (
	"errors"
	"fmt"

	"github.com/bodrovis-learning/kruk-shot/internal/output"
)

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
		mon := &m.monitors[i]

		img, err := captureWithRecovery(
			mon.display,
			&mon.session,
		)
		if err != nil {
			failures = append(
				failures,
				fmt.Errorf("capture %s: %w", mon.display.Name, err),
			)
			continue
		}

		path, err := writer.Save(i, img)
		if err != nil {
			failures = append(
				failures,
				fmt.Errorf("save %s: %w", mon.display.Name, err),
			)
			continue
		}

		fmt.Printf(
			"Saved %s (%dx%d)\n",
			path,
			img.Bounds().Dx(),
			img.Bounds().Dy(),
		)
	}

	return errors.Join(failures...)
}
