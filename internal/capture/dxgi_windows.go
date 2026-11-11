package capture

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Display describes an active desktop output and its position
// within the Windows virtual desktop.
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

// rect mirrors the Win32 RECT structure.
type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// outputDesc mirrors DXGI_OUTPUT_DESC from dxgi.h.
//
// AttachedToDesktop is a Win32 BOOL (32-bit integer).
// Rotation is a DXGI_MODE_ROTATION enum.
// Monitor is an opaque HMONITOR handle.
type outputDesc struct {
	DeviceName         [32]uint16
	DesktopCoordinates rect
	AttachedToDesktop  int32
	Rotation           uint32
	Monitor            uintptr
}

// ListDisplays returns all outputs attached to the Windows desktop.
// Displays are enumerated by adapter and then by output index.
func ListDisplays() ([]Display, error) {
	factory, err := newDXGIFactory()
	if err != nil {
		return nil, err
	}
	defer releaseCOM(factory)

	var displays []Display

	for adapterIndex := uint32(0); ; adapterIndex++ {
		adapter, err := newDXGIAdapter(factory, adapterIndex)

		if errors.Is(err, errDXGINotFound) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf(
				"enumerate adapter %d: %w",
				adapterIndex,
				err,
			)
		}

		outputs, err := listAdapterOutputs(adapter, adapterIndex)
		releaseCOM(adapter)

		if err != nil {
			return nil, fmt.Errorf(
				"list outputs for adapter %d: %w",
				adapterIndex,
				err,
			)
		}

		displays = append(displays, outputs...)
	}

	return displays, nil
}

// listAdapterOutputs enumerates active desktop outputs on an adapter.
func listAdapterOutputs(
	adapter unsafe.Pointer,
	adapterIndex uint32,
) ([]Display, error) {
	var displays []Display

	for outputIndex := uint32(0); ; outputIndex++ {
		output, err := newDXGIOutput(adapter, outputIndex)

		if errors.Is(err, errDXGINotFound) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf(
				"enumerate output %d: %w",
				outputIndex,
				err,
			)
		}

		desc, err := getOutputDesc(output)
		releaseCOM(output)

		if err != nil {
			return nil, fmt.Errorf(
				"describe output %d: %w",
				outputIndex,
				err,
			)
		}

		// Ignore outputs that are not attached to the desktop.
		if desc.AttachedToDesktop == 0 {
			continue
		}

		bounds := desc.DesktopCoordinates

		displays = append(displays, Display{
			AdapterIndex: adapterIndex,
			OutputIndex:  outputIndex,
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

// getOutputDesc retrieves the native DXGI description of an output.
func getOutputDesc(output unsafe.Pointer) (outputDesc, error) {
	var desc outputDesc

	hr, _, _ := syscall.SyscallN(
		comMethod(output, dxgiOutputGetDesc),
		uintptr(output),
		uintptr(unsafe.Pointer(&desc)),
	)

	if err := checkHRESULT("IDXGIOutput::GetDesc", hr); err != nil {
		return outputDesc{}, err
	}

	return desc, nil
}
