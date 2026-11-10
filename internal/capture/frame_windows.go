package capture

import (
	"errors"
	"image"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	dxgiErrorWaitTimeout = 0x887A0027
	dxgiErrorAccessLost  = 0x887A0026

	initialFrameAttempts = 20
	cachedFrameChecks    = 3
)

var ErrSessionLost = errors.New("desktop duplication session lost")

var iidID3D11Texture2D = windows.GUID{
	Data1: 0x6f15aaf2,
	Data2: 0xd208,
	Data3: 0x4e89,
	Data4: [8]byte{
		0x9a, 0xb4, 0x48, 0x95,
		0x35, 0xd3, 0x4f, 0x9c,
	},
}

type pointerPosition struct {
	X       int32
	Y       int32
	Visible int32
}

type frameInfo struct {
	LastPresentTime           int64
	LastMouseUpdateTime       int64
	AccumulatedFrames         uint32
	RectsCoalesced            int32
	ProtectedContentMaskedOut int32
	PointerPosition           pointerPosition
	TotalMetadataBufferSize   uint32
	PointerShapeBufferSize    uint32
}

// Capture returns a new desktop image or the most recently cached image.
// The returned image must be treated as read-only.
func (s *Session) Capture() (*image.RGBA, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("capture session is closed")
	}

	attempts := initialFrameAttempts
	timeoutMS := uintptr(500)

	if s.lastFrame != nil {
		attempts = cachedFrameChecks
		timeoutMS = 0
	}

	for attempt := 0; attempt < attempts; attempt++ {
		var info frameInfo
		var resource uintptr

		hr, _, _ := syscall.SyscallN(
			comMethod(s.duplication, 8), // AcquireNextFrame
			s.duplication,
			timeoutMS,
			uintptr(unsafe.Pointer(&info)),
			uintptr(unsafe.Pointer(&resource)),
		)

		switch uint32(hr) {
		case dxgiErrorWaitTimeout:
			if s.lastFrame != nil {
				return s.lastFrame, nil
			}
			continue

		case dxgiErrorAccessLost:
			// Never reuse a frame from an invalid duplication session.
			s.lastFrame = nil
			releaseCOM(resource)
			return nil, ErrSessionLost
		}

		if err := checkHRESULT("AcquireNextFrame", hr); err != nil {
			releaseCOM(resource)
			return nil, err
		}

		var img *image.RGBA
		var captureErr error

		// A zero timestamp means that only the mouse pointer changed.
		if info.LastPresentTime != 0 {
			if resource == 0 {
				captureErr = errors.New("desktop frame has no resource")
			} else {
				img, captureErr = s.readFrame(resource)
			}
		}

		releaseCOM(resource)

		// Every successful AcquireNextFrame needs a ReleaseFrame.
		hr, _, _ = syscall.SyscallN(
			comMethod(s.duplication, 14), // ReleaseFrame
			s.duplication,
		)

		if uint32(hr) == dxgiErrorAccessLost {
			s.lastFrame = nil
			return nil, ErrSessionLost
		}

		if err := checkHRESULT("ReleaseFrame", hr); err != nil {
			return nil, err
		}

		if captureErr != nil {
			return nil, captureErr
		}

		if img != nil {
			// Replace the cache only after a successful readback
			// and ReleaseFrame.
			s.lastFrame = img
			return img, nil
		}

		// Cursor-only event. Try another pending frame rather
		// than immediately returning the previous screenshot.
	}

	if s.lastFrame != nil {
		return s.lastFrame, nil
	}

	return nil, errors.New("no desktop image update received")
}

// readFrame converts an acquired DXGI resource into an RGBA image.
func (s *Session) readFrame(resource uintptr) (*image.RGBA, error) {
	var texture uintptr

	hr, _, _ := syscall.SyscallN(
		comMethod(resource, 0), // QueryInterface
		resource,
		uintptr(unsafe.Pointer(&iidID3D11Texture2D)),
		uintptr(unsafe.Pointer(&texture)),
	)
	defer releaseCOM(texture)

	if err := checkHRESULT("QueryInterface ID3D11Texture2D", hr); err != nil {
		return nil, err
	}

	var desc texture2DDesc

	syscall.SyscallN(
		comMethod(texture, 10), // GetDesc
		texture,
		uintptr(unsafe.Pointer(&desc)),
	)

	return readbackTexture(s.device, s.context, texture, desc)
}
