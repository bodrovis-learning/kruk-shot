package capture

import (
	"fmt"
	"syscall"
	"unsafe"
)

func comMethod(object unsafe.Pointer, index uintptr) uintptr {
	vtable := *(*unsafe.Pointer)(object)

	return *(*uintptr)(
		unsafe.Add(vtable, index*unsafe.Sizeof(uintptr(0))),
	)
}

func releaseCOM(object unsafe.Pointer) {
	if object == nil {
		return
	}

	// IUnknown::Release returns a reference count, not HRESULT.
	_, _, _ = syscall.SyscallN(comMethod(object, 2), uintptr(object))
}

func checkHRESULT(operation string, hr uintptr) error {
	if uint32(hr)&0x80000000 != 0 {
		return fmt.Errorf("%s: HRESULT 0x%08X", operation, uint32(hr))
	}

	return nil
}
