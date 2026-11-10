package app

import (
	"errors"
	"fmt"
	"image"

	"github.com/bodrovis-learning/kruk-shot/internal/capture"
)

// captureWithRecovery captures an image and recreates the session
// if desktop duplication was invalidated by a display mode change.
func captureWithRecovery(
	display capture.Display,
	session **capture.Session,
) (*image.RGBA, error) {
	for attempt := 0; attempt < 2; attempt++ {
		if *session == nil {
			newSession, err := capture.NewSession(display)
			if err != nil {
				return nil, fmt.Errorf("recreate capture session: %w", err)
			}

			*session = newSession
		}

		img, err := (*session).Capture()

		if !errors.Is(err, capture.ErrSessionLost) {
			return img, err
		}

		fmt.Printf("DXGI session lost for %s, recreating...\n", display.Name)

		(*session).Close()
		*session = nil
	}

	return nil, fmt.Errorf("capture session lost again after recovery")
}
