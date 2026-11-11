package output

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"
)

// PNGWriter saves a batch of screenshots with a shared timestamp.
type PNGWriter struct {
	dir       string
	timestamp string
}

// NewPNGWriter creates a writer for the screens directory next to the executable.
func NewPNGWriter() (*PNGWriter, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("get executable path: %w", err)
	}

	dir := filepath.Join(filepath.Dir(exePath), "screens")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create screenshots directory: %w", err)
	}

	return &PNGWriter{
		dir:       dir,
		timestamp: time.Now().Format("20060102-150405.000000000"),
	}, nil
}

// Save writes a screenshot without overwriting an existing file.
func (w *PNGWriter) Save(displayIndex int, img image.Image) (string, error) {
	filename := fmt.Sprintf(
		"kruk-%s-display-%d.png",
		w.timestamp,
		displayIndex,
	)

	path := filepath.Join(w.dir, filename)

	// Never overwrite an existing screenshot.
	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0o644,
	)
	if err != nil {
		return "", fmt.Errorf("create screenshot: %w", err)
	}

	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("encode PNG: %w", err)
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("close screenshot: %w", err)
	}

	return path, nil
}
