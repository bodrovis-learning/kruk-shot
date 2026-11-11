package capture

import (
	"errors"
	"image"
	"sync"
	"syscall"
	"unsafe"
)

// Session owns the D3D11 and Desktop Duplication resources
// associated with a single display.
type Session struct {
	mu sync.Mutex

	device      unsafe.Pointer
	context     unsafe.Pointer
	duplication unsafe.Pointer

	lastFrame *image.RGBA
	closed    bool
}

// NewSession initializes Desktop Duplication for a display.
//
// Temporary DXGI interfaces are released after initialization.
// The resulting session owns the D3D11 device, immediate context,
// and duplication interface until Close is called.
func NewSession(display Display) (*Session, error) {
	factory, err := newDXGIFactory()
	if err != nil {
		return nil, err
	}
	defer releaseCOM(factory)

	adapter, err := newDXGIAdapter(factory, display.AdapterIndex)
	if err != nil {
		return nil, err
	}
	defer releaseCOM(adapter)

	output, err := newDXGIOutput(adapter, display.OutputIndex)
	if err != nil {
		return nil, err
	}
	defer releaseCOM(output)

	output1, err := newDXGIOutput1(output)
	if err != nil {
		return nil, err
	}
	defer releaseCOM(output1)

	session := &Session{}
	initialized := false

	// Clean up resources if initialization fails halfway through.
	defer func() {
		if !initialized {
			session.Close()
		}
	}()

	if err := session.createDevice(adapter); err != nil {
		return nil, err
	}

	if err := session.createDuplication(output1); err != nil {
		return nil, err
	}

	initialized = true
	return session, nil
}

// createDevice creates a D3D11 device and its immediate context.
//
// The device must use the same adapter as the display output.
// With an explicit adapter, D3D_DRIVER_TYPE_UNKNOWN is required.
func (s *Session) createDevice(adapter unsafe.Pointer) error {
	hr, _, _ := d3d11CreateDevice.Call(
		uintptr(adapter),
		d3d11DriverTypeUnknown,
		0, // Software rasterizer module.
		0, // Device creation flags.
		0, // Use default feature levels.
		0, // Number of explicitly supplied feature levels.
		d3d11SDKVersion,
		uintptr(unsafe.Pointer(&s.device)),
		0, // Feature level output is not needed.
		uintptr(unsafe.Pointer(&s.context)),
	)

	if err := checkHRESULT("D3D11CreateDevice", hr); err != nil {
		return err
	}

	if s.device == nil || s.context == nil {
		return errors.New("D3D11CreateDevice returned incomplete resources")
	}

	return nil
}

// createDuplication initializes Desktop Duplication using
// the D3D11 device associated with the display's adapter.
func (s *Session) createDuplication(output1 unsafe.Pointer) error {
	hr, _, _ := syscall.SyscallN(
		comMethod(output1, dxgiOutput1Duplicate),
		uintptr(output1),
		uintptr(s.device),
		uintptr(unsafe.Pointer(&s.duplication)),
	)

	if err := checkHRESULT("DuplicateOutput", hr); err != nil {
		return err
	}

	if s.duplication == nil {
		return errors.New("DuplicateOutput returned a null interface")
	}

	return nil
}

// Close releases all resources owned by the session.
// Calling Close more than once is safe.
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	s.lastFrame = nil

	// Release in reverse order of acquisition.
	releaseCOM(s.duplication)
	releaseCOM(s.context)
	releaseCOM(s.device)

	s.duplication = nil
	s.context = nil
	s.device = nil
}
