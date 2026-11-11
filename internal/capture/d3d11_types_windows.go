package capture

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// Values from d3d11.h.
	d3d11UsageStaging  = 3       // D3D11_USAGE_STAGING: GPU copy, CPU read/write.
	d3d11CPUAccessRead = 0x20000 // D3D11_CPU_ACCESS_READ: allow CPU reads.
	d3d11MapRead       = 1       // D3D11_MAP_READ: map resource for reading.

	// d3d11SDKVersion corresponds to D3D11_SDK_VERSION in d3d11.h.
	// This value is required by D3D11CreateDevice and is not a DirectX version.
	d3d11SDKVersion = 7

	// Value from dxgiformat.h.
	dxgiFormatBGRA8 = 87 // DXGI_FORMAT_B8G8R8A8_UNORM: 8 bits per channel.

	// COM vtable indices, including inherited interface methods.
	// The indices are defined by the native Direct3D 11 interface layout.
	d3d11DeviceCreateTexture2D = 5
	d3d11ContextMap            = 14
	d3d11ContextUnmap          = 15
	d3d11ContextCopyResource   = 47

	// BGRA and RGBA both use four bytes per pixel.
	bytesPerPixel = 4

	// D3D_DRIVER_TYPE_UNKNOWN is required when supplying an adapter.
	d3d11DriverTypeUnknown = 0

	// ID3D11Texture2D::GetDesc (d3d11.h).
	d3d11Texture2DGetDesc = 10
)

// d3d11CreateDevice dynamically loads D3D11CreateDevice from d3d11.dll.
// The function creates a Direct3D 11 device and its immediate context.
var d3d11CreateDevice = windows.NewLazySystemDLL("d3d11.dll").
	NewProc("D3D11CreateDevice")

// mappedSubresource mirrors D3D11_MAPPED_SUBRESOURCE.
//
// Data points to mapped GPU resource memory. It remains valid only
// until ID3D11DeviceContext::Unmap is called.
//
// RowPitch is the distance in bytes between consecutive rows.
// It may be larger than Width * bytesPerPixel due to alignment.
type mappedSubresource struct {
	Data       unsafe.Pointer
	RowPitch   uint32
	DepthPitch uint32
}

// sampleDesc mirrors DXGI_SAMPLE_DESC.
type sampleDesc struct {
	Count   uint32
	Quality uint32
}

// texture2DDesc mirrors D3D11_TEXTURE2D_DESC.
//
// Field order and sizes must match the native Windows structure.
type texture2DDesc struct {
	Width          uint32
	Height         uint32
	MipLevels      uint32
	ArraySize      uint32
	Format         uint32
	SampleDesc     sampleDesc
	Usage          uint32
	BindFlags      uint32
	CPUAccessFlags uint32
	MiscFlags      uint32
}
