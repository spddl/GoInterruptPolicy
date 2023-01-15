package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

var (
	modole32 = windows.NewLazySystemDLL("ole32.dll")

	procCLSIDFromString = modole32.NewProc("CLSIDFromString")
)

func clsidFromString(lpsz *uint16, pclsid *windows.GUID) (hr int32) {
	r0, _, _ := syscall.SyscallN(procCLSIDFromString.Addr(), uintptr(unsafe.Pointer(lpsz)), uintptr(unsafe.Pointer(pclsid)))
	hr = int32(r0)
	return
}

// sys	clsidFromString(lpsz *uint16, pclsid *windows.GUID) (hr int32) = ole32.CLSIDFromString

// FromString parses "{XXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}" string to GUID.
func FromString(str string) (*windows.GUID, error) {
	strUTF16, err := syscall.UTF16PtrFromString(str)
	if err != nil {
		return nil, err
	}

	guid := &windows.GUID{}

	hr := clsidFromString(strUTF16, guid)
	if hr < 0 {
		return nil, syscall.Errno(hr)
	}

	return guid, nil
}

// ToString function converts GUID to string
// "{XXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX}".
//
// The resulting string is uppercase.
func ToString(guid *windows.GUID) string {
	return fmt.Sprintf("{%06X-%04X-%04X-%04X-%012X}", guid.Data1, guid.Data2, guid.Data3, guid.Data4[:2], guid.Data4[2:])
}
