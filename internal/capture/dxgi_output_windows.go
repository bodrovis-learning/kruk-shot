package capture

import (
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// COM vtable indices from DXGI interfaces.
	dxgiFactoryEnumAdapters1 = 12 // IDXGIFactory1::EnumAdapters1
	dxgiAdapterEnumOutputs   = 7  // IDXGIAdapter::EnumOutputs
	dxgiOutput1Duplicate     = 22 // IDXGIOutput1::DuplicateOutput

	// HRESULT returned when DXGI enumeration reaches its end.
	dxgiErrorNotFound = 0x887A0002 // DXGI_ERROR_NOT_FOUND

	dxgiOutputGetDesc = 7 // IDXGIOutput::GetDesc
)

// CreateDXGIFactory1 requests an IDXGIFactory1 interface.
var (
	createDXGIFactory1 = windows.NewLazySystemDLL("dxgi.dll").
				NewProc("CreateDXGIFactory1")

	iidIDXGIFactory1 = windows.GUID{
		Data1: 0x770aae78,
		Data2: 0xf26f,
		Data3: 0x4dba,
		Data4: [8]byte{
			0xa8, 0x29, 0x25, 0x3c,
			0x83, 0xd1, 0xb3, 0x87,
		},
	}

	// IID of IDXGIOutput1, defined in dxgi1_2.h.
	iidIDXGIOutput1 = windows.GUID{
		Data1: 0x00cddea8,
		Data2: 0x939b,
		Data3: 0x4b83,
		Data4: [8]byte{
			0xa3, 0x40, 0xa6, 0x85,
			0x22, 0x66, 0x66, 0xcc,
		},
	}

	errDXGINotFound = errors.New("DXGI enumeration item not found")
)

// newDXGIFactory creates an IDXGIFactory1 interface.
func newDXGIFactory() (unsafe.Pointer, error) {
	var factory unsafe.Pointer

	hr, _, _ := createDXGIFactory1.Call(
		uintptr(unsafe.Pointer(&iidIDXGIFactory1)),
		uintptr(unsafe.Pointer(&factory)),
	)

	if err := checkHRESULT("CreateDXGIFactory1", hr); err != nil {
		releaseCOM(factory)
		return nil, err
	}

	if factory == nil {
		return nil, errors.New("CreateDXGIFactory1 returned a null interface")
	}

	return factory, nil
}

// newDXGIAdapter retrieves an adapter from a DXGI factory.
func newDXGIAdapter(factory unsafe.Pointer, index uint32) (unsafe.Pointer, error) {
	var adapter unsafe.Pointer

	hr, _, _ := syscall.SyscallN(
		comMethod(factory, dxgiFactoryEnumAdapters1),
		uintptr(factory),
		uintptr(index),
		uintptr(unsafe.Pointer(&adapter)),
	)

	if uint32(hr) == dxgiErrorNotFound {
		releaseCOM(adapter)
		return nil, errDXGINotFound
	}

	if err := checkHRESULT("EnumAdapters1", hr); err != nil {
		releaseCOM(adapter)
		return nil, err
	}

	if adapter == nil {
		return nil, errors.New("EnumAdapters1 returned a null adapter")
	}

	return adapter, nil
}

// newDXGIOutput retrieves a display output from an adapter.
func newDXGIOutput(adapter unsafe.Pointer, index uint32) (unsafe.Pointer, error) {
	var output unsafe.Pointer

	hr, _, _ := syscall.SyscallN(
		comMethod(adapter, dxgiAdapterEnumOutputs),
		uintptr(adapter),
		uintptr(index),
		uintptr(unsafe.Pointer(&output)),
	)

	if uint32(hr) == dxgiErrorNotFound {
		releaseCOM(output)
		return nil, errDXGINotFound
	}

	if err := checkHRESULT("EnumOutputs", hr); err != nil {
		releaseCOM(output)
		return nil, err
	}

	if output == nil {
		return nil, errors.New("EnumOutputs returned a null output")
	}

	return output, nil
}

// newDXGIOutput1 queries the IDXGIOutput1 interface required
// for Desktop Duplication.
func newDXGIOutput1(output unsafe.Pointer) (unsafe.Pointer, error) {
	var output1 unsafe.Pointer

	hr, _, _ := syscall.SyscallN(
		comMethod(output, iunknownQueryInterface),
		uintptr(output),
		uintptr(unsafe.Pointer(&iidIDXGIOutput1)),
		uintptr(unsafe.Pointer(&output1)),
	)

	if err := checkHRESULT("QueryInterface IDXGIOutput1", hr); err != nil {
		releaseCOM(output1)
		return nil, err
	}

	if output1 == nil {
		return nil, errors.New("QueryInterface returned a null IDXGIOutput1")
	}

	return output1, nil
}
