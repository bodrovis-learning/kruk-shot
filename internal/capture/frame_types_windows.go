package capture

import "golang.org/x/sys/windows"

const (
	// DXGI HRESULT values (dxgi.h).
	dxgiErrorWaitTimeout = 0x887A0027 // No new frame before timeout.
	dxgiErrorAccessLost  = 0x887A0026 // Duplication interface is invalid.

	// IDXGIOutputDuplication vtable indices (dxgi1_2.h).
	dxgiDuplicationAcquireNextFrame = 8
	dxgiDuplicationReleaseFrame     = 14

	// Initial capture: wait up to 10 seconds for a desktop update.
	initialFrameAttempts  = 20
	initialFrameTimeoutMS = 500

	// Subsequent captures: check pending updates without waiting.
	cachedFrameChecks = 3
)

// IID_ID3D11Texture2D from d3d11.h.
var iidID3D11Texture2D = windows.GUID{
	Data1: 0x6f15aaf2,
	Data2: 0xd208,
	Data3: 0x4e89,
	Data4: [8]byte{
		0x9a, 0xb4, 0x48, 0x95,
		0x35, 0xd3, 0x4f, 0x9c,
	},
}

// pointerPosition mirrors DXGI_OUTDUPL_POINTER_POSITION.
type pointerPosition struct {
	X       int32 // Cursor X position.
	Y       int32 // Cursor Y position.
	Visible int32 // Win32 BOOL, not a Go bool.
}

// frameInfo mirrors DXGI_OUTDUPL_FRAME_INFO.
//
// All fields must be present and in the correct order because
// AcquireNextFrame writes directly into this structure.
type frameInfo struct {
	LastPresentTime           int64 // Last desktop image update timestamp.
	LastMouseUpdateTime       int64 // Last pointer update timestamp.
	AccumulatedFrames         uint32
	RectsCoalesced            int32 // Win32 BOOL.
	ProtectedContentMaskedOut int32 // Win32 BOOL.
	PointerPosition           pointerPosition
	TotalMetadataBufferSize   uint32
	PointerShapeBufferSize    uint32
}
