package capture

import (
	"errors"
	"image"
	"syscall"
	"unsafe"
)

// readFrame converts an acquired DXGI resource into an RGBA image.
//
// The resource is owned by acquireFrame and must remain valid
// until readbackTexture finishes copying its contents.
func (s *Session) readFrame(resource unsafe.Pointer) (*image.RGBA, error) {
	var texture unsafe.Pointer

	// IDXGIResource::QueryInterface(ID3D11Texture2D).
	hr, _, _ := syscall.SyscallN(
		comMethod(resource, iunknownQueryInterface),
		uintptr(resource),
		uintptr(unsafe.Pointer(&iidID3D11Texture2D)),
		uintptr(unsafe.Pointer(&texture)),
	)
	defer releaseCOM(texture)

	if err := checkHRESULT("QueryInterface ID3D11Texture2D", hr); err != nil {
		return nil, err
	}

	if texture == nil {
		return nil, errors.New("QueryInterface returned a null texture")
	}

	var desc texture2DDesc

	// ID3D11Texture2D::GetDesc returns void.
	_, _, _ = syscall.SyscallN(
		comMethod(texture, d3d11Texture2DGetDesc),
		uintptr(texture),
		uintptr(unsafe.Pointer(&desc)),
	)

	return readbackTexture(s.device, s.context, texture, desc)
}
