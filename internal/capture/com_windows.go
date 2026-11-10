package capture

import (
	"fmt"
	"syscall"
	"unsafe"
)

func comMethod(object uintptr, index uintptr) uintptr {
	vtable := *(*uintptr)(unsafe.Pointer(object))

	return *(*uintptr)(
		unsafe.Pointer(vtable + index*unsafe.Sizeof(uintptr(0))),
	)
}

func releaseCOM(object uintptr) {
	if object == 0 {
		return
	}

	syscall.SyscallN(comMethod(object, 2), object)
}

func checkHRESULT(operation string, hr uintptr) error {
	if uint32(hr)&0x80000000 != 0 {
		return fmt.Errorf("%s: HRESULT 0x%08X", operation, uint32(hr))
	}

	return nil
}
