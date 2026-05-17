//go:build windows

package main

import (
	"syscall"
)

const (
	idBtnOpen     = 1001
	idBtnDesktop  = 1002
	idBtnExplorer = 1003
	idBtnRefresh  = 1004
	idBtnInfo     = 1005
	idBtnCopyPath = 1006
	idBtnPreview  = 1007

	SW_SHOWNORMAL = 1

	DRIVE_REMOVABLE = 2
	DRIVE_FIXED     = 3
	DRIVE_REMOTE    = 4
	DRIVE_RAMDISK   = 6

	FILE_ATTRIBUTE_REPARSE_POINT = 0x00000400

	FO_COPY               = 0x0002
	FOF_RENAMEONCOLLISION = 0x0008
	FOF_ALLOWUNDO         = 0x0040
	FOF_NOCONFIRMMKDIR    = 0x0200

	CSIDL_DESKTOPDIRECTORY = 0x0010
	SHGFP_TYPE_CURRENT     = 0

	maxResults = 500

	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002
	GMEM_ZEROINIT  = 0x0040
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")

	procOpenClipboard           = user32.NewProc("OpenClipboard")
	procEmptyClipboard          = user32.NewProc("EmptyClipboard")
	procSetClipboardData        = user32.NewProc("SetClipboardData")
	procCloseClipboard          = user32.NewProc("CloseClipboard")
	procRegisterClipboardFormat = user32.NewProc("RegisterClipboardFormatW")

	procGetLogicalDrives = kernel32.NewProc("GetLogicalDrives")
	procGetDriveType     = kernel32.NewProc("GetDriveTypeW")
	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalFree       = kernel32.NewProc("GlobalFree")

	procShellExecute    = shell32.NewProc("ShellExecuteW")
	procSHFileOperation = shell32.NewProc("SHFileOperationW")
	procSHGetFolderPath = shell32.NewProc("SHGetFolderPathW")
)

type POINT struct {
	X int32
	Y int32
}

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type SHFILEOPSTRUCT struct {
	Hwnd                  uintptr
	WFunc                 uint32
	PFrom                 *uint16
	PTo                   *uint16
	FFlags                uint16
	FAnyOperationsAborted int32
	HNameMappings         uintptr
	LpszProgressTitle     *uint16
}
