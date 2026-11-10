package capture

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const dxgiErrorNotFound = 0x887A0002

type Display struct {
	AdapterIndex uint32
	OutputIndex  uint32
	Name         string
	X            int32
	Y            int32
	Width        int32
	Height       int32
	Rotation     uint32
}

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type outputDesc struct {
	DeviceName         [32]uint16
	DesktopCoordinates rect
	AttachedToDesktop  int32
	Rotation           uint32
	Monitor            uintptr
}

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
)

// ListDisplays returns all active desktop displays.
func ListDisplays() ([]Display, error) {
	var factory uintptr

	hr, _, _ := createDXGIFactory1.Call(
		uintptr(unsafe.Pointer(&iidIDXGIFactory1)),
		uintptr(unsafe.Pointer(&factory)),
	)
	defer releaseCOM(factory)

	if err := checkHRESULT("CreateDXGIFactory1", hr); err != nil {
		return nil, err
	}

	var displays []Display

	for i := uint32(0); ; i++ {
		var adapter uintptr

		hr, _, _ := syscall.SyscallN(
			comMethod(factory, 12), // EnumAdapters1
			factory,
			uintptr(i),
			uintptr(unsafe.Pointer(&adapter)),
		)

		if uint32(hr) == dxgiErrorNotFound {
			releaseCOM(adapter)
			break
		}

		if err := checkHRESULT("EnumAdapters1", hr); err != nil {
			releaseCOM(adapter)
			return nil, err
		}

		outputs, err := listAdapterOutputs(adapter, i)
		releaseCOM(adapter)

		if err != nil {
			return nil, err
		}

		displays = append(displays, outputs...)
	}

	return displays, nil
}

// listAdapterOutputs enumerates active outputs on the given adapter.
func listAdapterOutputs(
	adapter uintptr,
	adapterIndex uint32,
) ([]Display, error) {
	var displays []Display

	for i := uint32(0); ; i++ {
		var output uintptr

		hr, _, _ := syscall.SyscallN(
			comMethod(adapter, 7), // EnumOutputs
			adapter,
			uintptr(i),
			uintptr(unsafe.Pointer(&output)),
		)

		if uint32(hr) == dxgiErrorNotFound {
			releaseCOM(output)
			break
		}

		if err := checkHRESULT("EnumOutputs", hr); err != nil {
			releaseCOM(output)
			return nil, err
		}

		var desc outputDesc

		hr, _, _ = syscall.SyscallN(
			comMethod(output, 7), // GetDesc
			output,
			uintptr(unsafe.Pointer(&desc)),
		)

		releaseCOM(output)

		if err := checkHRESULT("GetDesc", hr); err != nil {
			return nil, err
		}

		if desc.AttachedToDesktop == 0 {
			continue
		}

		bounds := desc.DesktopCoordinates

		displays = append(displays, Display{
			AdapterIndex: adapterIndex,
			OutputIndex:  i,
			Name:         windows.UTF16ToString(desc.DeviceName[:]),
			X:            bounds.Left,
			Y:            bounds.Top,
			Width:        bounds.Right - bounds.Left,
			Height:       bounds.Bottom - bounds.Top,
			Rotation:     desc.Rotation,
		})
	}

	return displays, nil
}
