package capture

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	// All COM interfaces inherit the first three IUnknown methods:
	// 0: QueryInterface
	// 1: AddRef
	// 2: Release
	iunknownRelease = 2

	// IUnknown methods occupy the first three COM vtable slots.
	iunknownQueryInterface = 0
)

// comMethod returns the address of a method in a COM interface's vtable.
//
// A COM interface pointer points to a vtable pointer. The vtable is
// an array of function pointers, with each entry occupying one
// pointer-sized slot (4 bytes on 32-bit, 8 bytes on 64-bit).
func comMethod(object unsafe.Pointer, index uintptr) uintptr {
	vtable := *(*unsafe.Pointer)(object)

	return *(*uintptr)(
		unsafe.Add(vtable, index*unsafe.Sizeof(uintptr(0))),
	)
}

// releaseCOM decrements the reference count of a COM interface.
// A reference count of zero allows the object to be destroyed.
func releaseCOM(object unsafe.Pointer) {
	if object == nil {
		return
	}

	// IUnknown::Release returns ULONG, not HRESULT.
	// The returned reference count is not needed.
	_, _, _ = syscall.SyscallN(
		comMethod(object, iunknownRelease),
		uintptr(object),
	)
}

// checkHRESULT converts a failed Windows HRESULT into a Go error.
func checkHRESULT(operation string, hr uintptr) error {
	// HRESULT is a signed 32-bit value.
	// Negative values indicate failure; zero and positive values
	// indicate success or another non-failure status.
	if int32(hr) < 0 {
		return fmt.Errorf("%s: HRESULT 0x%08X", operation, uint32(hr))
	}

	return nil
}
