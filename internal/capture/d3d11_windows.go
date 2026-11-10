package capture

import "golang.org/x/sys/windows"

const d3d11SDKVersion = 7

var d3d11CreateDevice = windows.NewLazySystemDLL("d3d11.dll").
	NewProc("D3D11CreateDevice")
