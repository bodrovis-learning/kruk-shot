package capture

import (
	"image"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var iidIDXGIOutput1 = windows.GUID{
	Data1: 0x00cddea8,
	Data2: 0x939b,
	Data3: 0x4b83,
	Data4: [8]byte{
		0xa3, 0x40, 0xa6, 0x85,
		0x22, 0x66, 0x66, 0xcc,
	},
}

// Session maintains desktop duplication resources for one display.
type Session struct {
	mu sync.Mutex

	device      unsafe.Pointer
	context     unsafe.Pointer
	duplication unsafe.Pointer

	lastFrame *image.RGBA
	closed    bool
}

// NewSession initializes desktop duplication for the specified display.
func NewSession(display Display) (*Session, error) {
	session := &Session{}
	initialized := false

	// Release partially initialized resources on failure.
	defer func() {
		if !initialized {
			session.Close()
		}
	}()

	var factory unsafe.Pointer

	hr, _, _ := createDXGIFactory1.Call(
		uintptr(unsafe.Pointer(&iidIDXGIFactory1)),
		uintptr(unsafe.Pointer(&factory)),
	)
	defer releaseCOM(factory)

	if err := checkHRESULT("CreateDXGIFactory1", hr); err != nil {
		return nil, err
	}

	// Find the adapter that owns this display.
	var adapter unsafe.Pointer

	hr, _, _ = syscall.SyscallN(
		comMethod(factory, 12), // EnumAdapters1
		uintptr(factory),
		uintptr(display.AdapterIndex),
		uintptr(unsafe.Pointer(&adapter)),
	)
	defer releaseCOM(adapter)

	if err := checkHRESULT("EnumAdapters1", hr); err != nil {
		return nil, err
	}

	// Find the output on that adapter.
	var output unsafe.Pointer

	hr, _, _ = syscall.SyscallN(
		comMethod(adapter, 7), // EnumOutputs
		uintptr(adapter),
		uintptr(display.OutputIndex),
		uintptr(unsafe.Pointer(&output)),
	)
	defer releaseCOM(output)

	if err := checkHRESULT("EnumOutputs", hr); err != nil {
		return nil, err
	}

	// Request IDXGIOutput1.
	var output1 unsafe.Pointer

	hr, _, _ = syscall.SyscallN(
		comMethod(output, 0), // QueryInterface
		uintptr(output),
		uintptr(unsafe.Pointer(&iidIDXGIOutput1)),
		uintptr(unsafe.Pointer(&output1)),
	)
	defer releaseCOM(output1)

	if err := checkHRESULT("QueryInterface IDXGIOutput1", hr); err != nil {
		return nil, err
	}

	// Create a D3D11 device on the display's adapter.
	hr, _, _ = d3d11CreateDevice.Call(
		uintptr(adapter),
		0, // D3D_DRIVER_TYPE_UNKNOWN
		0,
		0,
		0,
		0,
		d3d11SDKVersion,
		uintptr(unsafe.Pointer(&session.device)),
		0, // Feature level is not needed.
		uintptr(unsafe.Pointer(&session.context)),
	)

	if err := checkHRESULT("D3D11CreateDevice", hr); err != nil {
		return nil, err
	}

	// Create the desktop duplication interface.
	hr, _, _ = syscall.SyscallN(
		comMethod(output1, 22), // DuplicateOutput
		uintptr(output1),
		uintptr(session.device),
		uintptr(unsafe.Pointer(&session.duplication)),
	)

	if err := checkHRESULT("DuplicateOutput", hr); err != nil {
		return nil, err
	}

	initialized = true
	return session, nil
}

// Close releases all resources owned by the session.
// It is safe to call Close more than once.
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	s.lastFrame = nil

	releaseCOM(s.duplication)
	releaseCOM(s.context)
	releaseCOM(s.device)

	s.duplication = nil
	s.context = nil
	s.device = nil
}
