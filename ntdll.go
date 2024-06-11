package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"syscall"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	// Library
	libNtdll *windows.LazyDLL

	// Functions
	ntQueryInformationProcess  *windows.LazyProc
	ntQuerySystemInformationEx *windows.LazyProc
	ntQueryKey                 *windows.LazyProc
)

func init() {
	// Library
	libNtdll = windows.NewLazySystemDLL("ntdll.dll")

	// Functions
	ntQueryKey = libNtdll.NewProc("NtQueryKey")
	ntQueryInformationProcess = libNtdll.NewProc("NtQueryInformationProcess")
	ntQuerySystemInformationEx = libNtdll.NewProc("NtQuerySystemInformationEx")
}

// The KeyInformationClass constants have been derived from the KEY_INFORMATION_CLASS enum definition.
type KeyInformationClass uint32

const (
	KeyBasicInformation KeyInformationClass = iota
	KeyNodeInformation
	KeyFullInformation
	KeyNameInformation
	KeyCachedInformation
	KeyFlagsInformation
	KeyVirtualizationInformation
	KeyHandleTagsInformation
	MaxKeyInfoClass
)

const STATUS_BUFFER_TOO_SMALL = 0xC0000023

// OUT-parameter: KeyInformation, ResultLength.
// *OPT-parameter: KeyInformation.
func NtQueryKey(
	keyHandle uintptr,
	keyInformationClass KeyInformationClass,
	keyInformation *byte,
	length uint32,
	resultLength *uint32,
) int {
	r0, _, _ := ntQueryKey.Call(keyHandle,
		uintptr(keyInformationClass),
		uintptr(unsafe.Pointer(keyInformation)),
		uintptr(length),
		uintptr(unsafe.Pointer(resultLength)))
	return int(r0)
}

func GetRegistryLocation(regHandle uintptr) (string, error) {
	var size uint32 = 0
	result := NtQueryKey(regHandle, KeyNameInformation, nil, 0, &size)
	if result == STATUS_BUFFER_TOO_SMALL {
		buf := make([]byte, size)
		if result := NtQueryKey(regHandle, KeyNameInformation, &buf[0], size, &size); result == 0 {
			regPath, err := DecodeUTF16(buf)
			if err != nil {
				log.Println(err)
			}

			tempRegPath := replaceRegistryMachine(regPath)
			tempRegPath = generalizeControlSet(tempRegPath)

			return `HKEY_LOCAL_MACHINE\` + tempRegPath, nil
		}
	} else {
		return "", fmt.Errorf("error: 0x%X", result)
	}
	return "", nil
}

func DecodeUTF16(b []byte) (string, error) {
	if len(b)%2 != 0 {
		return "", fmt.Errorf("must have even length byte slice")
	}

	u16s := make([]uint16, 1)
	ret := &bytes.Buffer{}
	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}

// NtQueryInformationProcess is a wrapper for ntdll.NtQueryInformationProcess.
// The handle must have the PROCESS_QUERY_INFORMATION access right.
// Returns an error of type NTStatus.
func NtQueryInformationProcess(
	processHandle windows.Handle,
	processInformationClass int32,
	processInformation windows.Pointer,
	processInformationLength uint32,
	returnLength *uint32,
) error {
	r1, _, err := ntQueryInformationProcess.Call(
		uintptr(processHandle),
		uintptr(processInformationClass),
		uintptr(unsafe.Pointer(processInformation)),
		uintptr(processInformationLength),
		uintptr(unsafe.Pointer(returnLength)))
	if int(r1) < 0 {
		return os.NewSyscallError("NtQueryInformationProcess", err)
	}
	return nil
}

func NtQuerySystemInformationEx(
	systemInformationClass int32,
	inputBuffer unsafe.Pointer,
	inputBufferLength uint32,
	systemInformation *uint64,
	systemInformationLength uint32,
	returnLength *uint32,
) uint32 {
	r1, _, _ := syscall.SyscallN(ntQuerySystemInformationEx.Addr(),
		uintptr(systemInformationClass),
		uintptr(inputBuffer),
		uintptr(inputBufferLength),
		uintptr(unsafe.Pointer(systemInformation)),
		uintptr(systemInformationLength),
		uintptr(unsafe.Pointer(returnLength)),
	)
	return uint32(r1)
}
