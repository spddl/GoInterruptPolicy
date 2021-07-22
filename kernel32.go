package main

// https://github.com/prometheus-community/windows_exporter/blob/74eac8f29b8083b9e6a4832d739748739e4e3fe0/headers/sysinfoapi/sysinfoapi.go#L44

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// wProcessorArchitecture is a wrapper for the union found in LP_SYSTEM_INFO
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-system_info
type wProcessorArchitecture struct {
	WProcessorArchitecture uint16
	WReserved              uint16
}

// ProcessorArchitecture is an idiomatic wrapper for wProcessorArchitecture
type ProcessorArchitecture uint16

// Idiomatic values for wProcessorArchitecture
const (
	AMD64   ProcessorArchitecture = 9
	ARM     ProcessorArchitecture = 5
	ARM64   ProcessorArchitecture = 12
	IA64    ProcessorArchitecture = 6
	INTEL   ProcessorArchitecture = 0
	UNKNOWN ProcessorArchitecture = 0xffff
)

// SystemInfo is an idiomatic wrapper for LpSystemInfo
type SystemInfo struct {
	Arch                      ProcessorArchitecture
	PageSize                  uint32
	MinimumApplicationAddress uintptr
	MaximumApplicationAddress uintptr
	ActiveProcessorMask       uint
	NumberOfProcessors        uint32
	ProcessorType             uint32
	AllocationGranularity     uint32
	ProcessorLevel            uint16
	ProcessorRevision         uint16
}

// LpSystemInfo is a wrapper for LPSYSTEM_INFO
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/ns-sysinfoapi-system_info
type lpSystemInfo struct {
	Arch                        wProcessorArchitecture
	DwPageSize                  uint32
	LpMinimumApplicationAddress uintptr
	LpMaximumApplicationAddress uintptr
	DwActiveProcessorMask       uint
	DwNumberOfProcessors        uint32
	DwProcessorType             uint32
	DwAllocationGranularity     uint32
	WProcessorLevel             uint16
	WProcessorRevision          uint16
}

var (
	// Library
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	// Functions
	// activateActCtx *windows.LazyProc
	procGetSystemInfo = kernel32.NewProc("GetSystemInfo")
)

// GetSystemInfo is an idiomatic wrapper for the GetSystemInfo function from sysinfoapi
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getsysteminfo
func GetSystemInfo() SystemInfo {
	var info lpSystemInfo
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&info)))
	return SystemInfo{
		Arch:                      ProcessorArchitecture(info.Arch.WProcessorArchitecture),
		PageSize:                  info.DwPageSize,
		MinimumApplicationAddress: info.LpMinimumApplicationAddress,
		MaximumApplicationAddress: info.LpMinimumApplicationAddress,
		ActiveProcessorMask:       info.DwActiveProcessorMask,
		NumberOfProcessors:        info.DwNumberOfProcessors,
		ProcessorType:             info.DwProcessorType,
		AllocationGranularity:     info.DwAllocationGranularity,
		ProcessorLevel:            info.WProcessorLevel,
		ProcessorRevision:         info.WProcessorRevision,
	}
}
