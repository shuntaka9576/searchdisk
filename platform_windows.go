//go:build windows

package main

import (
	"os"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetLogicalDrives = kernel32.NewProc("GetLogicalDrives")
	procGetDriveTypeW    = kernel32.NewProc("GetDriveTypeW")
)

const driveFixed = 3

// listFixedDrives は固定ディスク (DRIVE_FIXED) のルートパスを返す。
func listFixedDrives() []string {
	bitmask, _, _ := procGetLogicalDrives.Call()
	var drives []string
	for i := range 26 {
		if bitmask&(1<<uint(i)) == 0 {
			continue
		}
		root := string(rune('A'+i)) + `:\`
		ptr, err := syscall.UTF16PtrFromString(root)
		if err != nil {
			continue
		}
		t, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(ptr)))
		if t == driveFixed {
			drives = append(drives, root)
		}
	}
	return drives
}

// creationTime は Windows の作成日時を返す。
func creationTime(fi os.FileInfo) time.Time {
	if d, ok := fi.Sys().(*syscall.Win32FileAttributeData); ok {
		return time.Unix(0, d.CreationTime.Nanoseconds())
	}
	return fi.ModTime()
}

// isReparse は ReparsePoint 属性が立っているかを返す (ジャンクション等)。
func isReparse(fi os.FileInfo) bool {
	if d, ok := fi.Sys().(*syscall.Win32FileAttributeData); ok {
		return d.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
	}
	return false
}
