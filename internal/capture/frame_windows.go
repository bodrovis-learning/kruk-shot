package capture

import (
	"errors"
	"image"
	"syscall"
	"unsafe"
)

// ErrSessionLost indicates that desktop duplication must be recreated.
// This can happen when Windows changes the display mode.
var ErrSessionLost = errors.New("desktop duplication session lost")

// Capture returns a new desktop image or the most recently cached image.
//
// The first capture waits for a desktop image update. Later captures
// use a non-blocking check and fall back to the cached image.
//
// The returned image must be treated as read-only.
func (s *Session) Capture() (*image.RGBA, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("capture session is closed")
	}

	attempts := initialFrameAttempts
	timeoutMS := uintptr(initialFrameTimeoutMS)

	if s.lastFrame != nil {
		attempts = cachedFrameChecks
		timeoutMS = 0
	}

	for attempt := 0; attempt < attempts; attempt++ {
		img, received, err := s.acquireFrame(timeoutMS)
		if err != nil {
			// An invalid duplication session must never retain
			// a cached image from the previous display mode.
			if errors.Is(err, ErrSessionLost) {
				s.lastFrame = nil
			}

			return nil, err
		}

		if img != nil {
			s.lastFrame = img
			return img, nil
		}

		// No new frame was available. Return the cached image.
		if !received && s.lastFrame != nil {
			return s.lastFrame, nil
		}

		// A cursor-only event occurred, or the initial capture
		// timed out. Try again within the configured limit.
	}

	if s.lastFrame != nil {
		return s.lastFrame, nil
	}

	return nil, errors.New("no desktop image update received")
}

// acquireFrame attempts to acquire and read one DXGI frame.
//
// The returned bool indicates whether AcquireNextFrame succeeded.
// A successful acquisition can still be cursor-only and return nil.
//
// Every successful acquisition must be followed by ReleaseFrame.
func (s *Session) acquireFrame(
	timeoutMS uintptr,
) (*image.RGBA, bool, error) {
	var info frameInfo
	var resource unsafe.Pointer

	hr, _, _ := syscall.SyscallN(
		comMethod(s.duplication, dxgiDuplicationAcquireNextFrame),
		uintptr(s.duplication),
		timeoutMS,
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Pointer(&resource)),
	)

	switch uint32(hr) {
	case dxgiErrorWaitTimeout:
		releaseCOM(resource)
		return nil, false, nil

	case dxgiErrorAccessLost:
		releaseCOM(resource)
		return nil, false, ErrSessionLost
	}

	if err := checkHRESULT("AcquireNextFrame", hr); err != nil {
		releaseCOM(resource)
		return nil, false, err
	}

	var img *image.RGBA
	var readErr error

	// LastPresentTime is zero if only the pointer changed.
	// We must not replace the cached desktop image with it.
	if info.LastPresentTime != 0 {
		if resource == nil {
			readErr = errors.New("desktop frame has no resource")
		} else {
			img, readErr = s.readFrame(resource)
		}
	}

	// The acquired resource is no longer needed.
	releaseCOM(resource)

	// ReleaseFrame is required after every successful acquisition,
	// including cursor-only updates and failed readbacks.
	hr, _, _ = syscall.SyscallN(
		comMethod(s.duplication, dxgiDuplicationReleaseFrame),
		uintptr(s.duplication),
	)

	if uint32(hr) == dxgiErrorAccessLost {
		return nil, true, ErrSessionLost
	}

	if err := checkHRESULT("ReleaseFrame", hr); err != nil {
		return nil, true, err
	}

	if readErr != nil {
		return nil, true, readErr
	}

	return img, true, nil
}
