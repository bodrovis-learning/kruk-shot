package capture

import (
	"fmt"
	"image"
	"syscall"
	"unsafe"
)

const (
	d3d11UsageStaging  = 3
	d3d11CPUAccessRead = 0x20000
	d3d11MapRead       = 1
	dxgiFormatBGRA8    = 87
)

type mappedSubresource struct {
	Data       uintptr
	RowPitch   uint32
	DepthPitch uint32
}

type texture2DDesc struct {
	Width          uint32
	Height         uint32
	MipLevels      uint32
	ArraySize      uint32
	Format         uint32
	SampleCount    uint32
	SampleQuality  uint32
	Usage          uint32
	BindFlags      uint32
	CPUAccessFlags uint32
	MiscFlags      uint32
}

func readbackTexture(
	device uintptr,
	context uintptr,
	source uintptr,
	desc texture2DDesc,
) (*image.RGBA, error) {
	if desc.Format != dxgiFormatBGRA8 {
		return nil, fmt.Errorf("unsupported texture format: %d", desc.Format)
	}

	stagingDesc := desc
	stagingDesc.Usage = d3d11UsageStaging
	stagingDesc.BindFlags = 0
	stagingDesc.CPUAccessFlags = d3d11CPUAccessRead
	stagingDesc.MiscFlags = 0

	var staging uintptr

	hr, _, _ := syscall.SyscallN(
		comMethod(device, 5),
		device,
		uintptr(unsafe.Pointer(&stagingDesc)),
		0,
		uintptr(unsafe.Pointer(&staging)),
	)
	defer releaseCOM(staging)

	if err := checkHRESULT("CreateTexture2D", hr); err != nil {
		return nil, err
	}

	// Copy the GPU texture into CPU-readable memory.
	syscall.SyscallN(
		comMethod(context, 47),
		context,
		staging,
		source,
	)

	var mapped mappedSubresource

	hr, _, _ = syscall.SyscallN(
		comMethod(context, 14),
		context,
		staging,
		0,
		d3d11MapRead,
		0,
		uintptr(unsafe.Pointer(&mapped)),
	)

	if err := checkHRESULT("Map", hr); err != nil {
		return nil, err
	}

	defer syscall.SyscallN(
		comMethod(context, 15),
		context,
		staging,
		0,
	)

	if mapped.Data == 0 {
		return nil, fmt.Errorf("Map returned a null data pointer")
	}

	width := int(desc.Width)
	height := int(desc.Height)
	rowPitch := int(mapped.RowPitch)

	if width <= 0 || height <= 0 || rowPitch < width*4 {
		return nil, fmt.Errorf("invalid texture dimensions or row pitch")
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	pixels := unsafe.Slice(
		(*byte)(unsafe.Pointer(mapped.Data)),
		rowPitch*height,
	)

	for y := 0; y < height; y++ {
		srcRow := pixels[y*rowPitch : y*rowPitch+width*4]
		dstRow := img.Pix[y*img.Stride : y*img.Stride+width*4]

		for x := 0; x < width; x++ {
			offset := x * 4

			dstRow[offset] = srcRow[offset+2]
			dstRow[offset+1] = srcRow[offset+1]
			dstRow[offset+2] = srcRow[offset]
			dstRow[offset+3] = 255
		}
	}

	return img, nil
}
