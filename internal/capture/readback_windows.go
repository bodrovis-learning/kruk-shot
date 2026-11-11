package capture

import (
	"fmt"
	"image"
	"syscall"
	"unsafe"
)

// readbackTexture copies a GPU texture into CPU-accessible memory
// and converts its BGRA pixels into a Go RGBA image.
func readbackTexture(
	device unsafe.Pointer,
	context unsafe.Pointer,
	source unsafe.Pointer,
	desc texture2DDesc,
) (*image.RGBA, error) {
	if desc.Format != dxgiFormatBGRA8 {
		return nil, fmt.Errorf("unsupported texture format: %d", desc.Format)
	}

	// Staging resources support GPU copy operations and CPU access.
	// They cannot be bound to the rendering pipeline.
	stagingDesc := desc
	stagingDesc.Usage = d3d11UsageStaging
	stagingDesc.BindFlags = 0
	stagingDesc.CPUAccessFlags = d3d11CPUAccessRead
	stagingDesc.MiscFlags = 0

	var staging unsafe.Pointer

	// ID3D11Device::CreateTexture2D.
	hr, _, _ := syscall.SyscallN(
		comMethod(device, d3d11DeviceCreateTexture2D),
		uintptr(device),
		uintptr(unsafe.Pointer(&stagingDesc)),
		0, // No initial texture data.
		uintptr(unsafe.Pointer(&staging)),
	)
	defer releaseCOM(staging)

	if err := checkHRESULT("CreateTexture2D", hr); err != nil {
		return nil, err
	}

	if staging == nil {
		return nil, fmt.Errorf("CreateTexture2D returned a null texture")
	}

	// ID3D11DeviceContext::CopyResource returns void.
	// The GPU copies the source texture into our staging texture.
	_, _, _ = syscall.SyscallN(
		comMethod(context, d3d11ContextCopyResource),
		uintptr(context),
		uintptr(staging),
		uintptr(source),
	)

	var mapped mappedSubresource

	// ID3D11DeviceContext::Map returns HRESULT.
	// Subresource 0 is the first (and only) texture subresource.
	// MapFlags 0 allows waiting until the GPU finishes copying.
	hr, _, _ = syscall.SyscallN(
		comMethod(context, d3d11ContextMap),
		uintptr(context),
		uintptr(staging),
		0, // Subresource index.
		d3d11MapRead,
		0, // Map flags.
		uintptr(unsafe.Pointer(&mapped)),
	)

	if err := checkHRESULT("Map", hr); err != nil {
		return nil, err
	}

	// Unmap must happen before the staging texture is released.
	defer func() {
		// ID3D11DeviceContext::Unmap returns void.
		_, _, _ = syscall.SyscallN(
			comMethod(context, d3d11ContextUnmap),
			uintptr(context),
			uintptr(staging),
			0,
		)
	}()

	if mapped.Data == nil {
		return nil, fmt.Errorf("map returned a null data pointer")
	}

	width := int(desc.Width)
	height := int(desc.Height)
	rowPitch := int(mapped.RowPitch)

	if width <= 0 || height <= 0 || rowPitch < width*bytesPerPixel {
		return nil, fmt.Errorf("invalid texture dimensions or row pitch")
	}

	// Include only the bytes needed through the last visible row.
	// This also avoids assuming that the last row's padding is readable.
	bufferSize := uint64(mapped.RowPitch)*uint64(desc.Height-1) +
		uint64(desc.Width)*bytesPerPixel

	if bufferSize > uint64(int(^uint(0)>>1)) {
		return nil, fmt.Errorf("mapped texture buffer is too large")
	}

	// Data points to native memory owned by D3D11.
	// We must copy its contents before Unmap.
	pixels := unsafe.Slice(
		(*byte)(mapped.Data),
		int(bufferSize),
	)

	return bgraToRGBA(pixels, width, height, rowPitch), nil
}
