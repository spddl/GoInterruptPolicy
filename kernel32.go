package main

// https://github.com/prometheus-community/windows_exporter/blob/74eac8f29b8083b9e6a4832d739748739e4e3fe0/headers/sysinfoapi/sysinfoapi.go#L44

import (
	"syscall"
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
	libKernel32 = windows.NewLazySystemDLL("kernel32.dll")

	// Functions
	getSystemCpuSetInformation = libKernel32.NewProc("GetSystemCpuSetInformation")
	getSystemInfo              = libKernel32.NewProc("GetSystemInfo")
)

// GetSystemInfo is an idiomatic wrapper for the GetSystemInfo function from sysinfoapi
// https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getsysteminfo
func GetSystemInfo() SystemInfo {
	var info lpSystemInfo
	getSystemInfo.Call(uintptr(unsafe.Pointer(&info)))
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

// The SystemInformationClass constants have been derived from the SYSTEM_INFORMATION_CLASS enum definition.
const (
	SystemAllowedCpuSetsInformation = 0xA8
	SystemCpuSetInformation         = 0xAF

	ProcessDefaultCpuSetsInformation = 0x42

	// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	PROCESS_SET_LIMITED_INFORMATION   = 0x2000
)

const (
	SYSTEM_CPU_SET_INFORMATION_PARKED                      uint32 = 0x1
	SYSTEM_CPU_SET_INFORMATION_ALLOCATED                   uint32 = 0x2
	SYSTEM_CPU_SET_INFORMATION_ALLOCATED_TO_TARGET_PROCESS uint32 = 0x4
	SYSTEM_CPU_SET_INFORMATION_REALTIME                    uint32 = 0x8
)

// https://github.com/zzl/go-win32api/blob/d9f481c2ab64b5df06e8d62e1bac33dd9141de43/win32/System.SystemInformation.go

type SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1_Anonymous struct {
	Bitfield_ byte
}

type SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1 struct {
	SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1_Anonymous
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1) AllFlags() *byte {
	return (*byte)(unsafe.Pointer(t))
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1) AllFlagsVal() byte {
	return *(*byte)(unsafe.Pointer(t))
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1) Anonymous() *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1_Anonymous {
	return (*SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1_Anonymous)(unsafe.Pointer(t))
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1) AnonymousVal() SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1_Anonymous {
	return *(*SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1_Anonymous)(unsafe.Pointer(t))
}

type SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous2 struct {
	Data [1]uint32
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous2) Reserved() *uint32 {
	return (*uint32)(unsafe.Pointer(t))
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous2) ReservedVal() uint32 {
	return *(*uint32)(unsafe.Pointer(t))
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous2) SchedulingClass() *byte {
	return (*byte)(unsafe.Pointer(t))
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous2) SchedulingClassVal() byte {
	return *(*byte)(unsafe.Pointer(t))
}

type SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet struct {
	Id                    uint32
	Group                 uint16
	LogicalProcessorIndex byte
	CoreIndex             byte
	LastLevelCacheIndex   byte
	NumaNodeIndex         byte
	EfficiencyClass       byte
	SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous1
	SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet_Anonymous2
	AllocationTag uint64
}

type SYSTEM_CPU_SET_INFORMATION_Anonymous struct {
	Data [3]uint64
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous) CpuSet() *SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet {
	return (*SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet)(unsafe.Pointer(t))
}

func (t *SYSTEM_CPU_SET_INFORMATION_Anonymous) CpuSetVal() SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet {
	return *(*SYSTEM_CPU_SET_INFORMATION_Anonymous_CpuSet)(unsafe.Pointer(t))
}

type CPU_SET_INFORMATION_TYPE int32

type SYSTEM_CPU_SET_INFORMATION struct {
	Size uint32
	Type CPU_SET_INFORMATION_TYPE
	SYSTEM_CPU_SET_INFORMATION_Anonymous
}

func GetSystemCpuSetInformation(
	information *SYSTEM_CPU_SET_INFORMATION,
	bufferLength uint32,
	returnedLength *uint32,
	process uintptr,
	flags uint32,
) uint32 {
	r1, _, _ := syscall.SyscallN(getSystemCpuSetInformation.Addr(),
		uintptr(unsafe.Pointer(information)),
		uintptr(bufferLength),
		uintptr(unsafe.Pointer(returnedLength)),
		process,
		uintptr(flags),
	)
	return uint32(r1)
}
