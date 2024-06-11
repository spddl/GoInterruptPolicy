package main

import (
	"log"
	"strconv"
	"time"

	"golang.org/x/sys/windows/registry"
)

type Device struct {
	Idata               DevInfoData
	reg                 registry.Key
	IrqPolicy           int32
	DeviceDesc          string
	DeviceIDs           []string
	DevObjName          string
	Driver              string
	LocationInformation string
	FriendlyName        string
	LastChange          time.Time

	// AffinityPolicy
	DevicePolicy          uint32
	DevicePriority        uint32
	AssignmentSetOverride Bits

	// MessageSignaledInterruptProperties
	MsiSupported       uint32
	MessageNumberLimit uint32
	MaxMSILimit        uint32
	InterruptTypeMap   Bits
}

const (
	// https://docs.microsoft.com/en-us/windows-hardware/drivers/kernel/interrupt-affinity-and-priority
	IrqPolicyMachineDefault                    = iota // 0
	IrqPolicyAllCloseProcessors                       // 1
	IrqPolicyOneCloseProcessor                        // 2
	IrqPolicyAllProcessorsInMachine                   // 3
	IrqPolicySpecifiedProcessors                      // 4
	IrqPolicySpreadMessagesAcrossAllProcessors        // 5
)

type Bits uint64

var CPUMap map[Bits]string

var CPUBits []Bits
var InterruptTypeMap = map[Bits]string{
	0: "unknown",
	1: "LineBased",
	2: "Msi",

	4: "MsiX",
}

var sysInfo SystemInfo
var handle DevInfo

const ZeroBit = Bits(0)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	sysInfo = GetSystemInfo()
	CPUMap = make(map[Bits]string, sysInfo.NumberOfProcessors)
	var index Bits = 1
	for i := 0; i < int(sysInfo.NumberOfProcessors); i++ {
		indexString := strconv.Itoa(i)
		CPUMap[index] = indexString
		CPUBits = append(CPUBits, index)
		index *= 2
	}
}

func Set(b, flag Bits) Bits    { return b | flag }
func Clear(b, flag Bits) Bits  { return b &^ flag }
func Toggle(b, flag Bits) Bits { return b ^ flag }
func Has(b, flag Bits) bool    { return b&flag != 0 }

// https://gist.github.com/chiro-hiro/2674626cebbcb5a676355b7aaac4972d
func i64tob(val uint64) []byte {
	r := make([]byte, 8)
	for i := uint64(0); i < 8; i++ {
		r[i] = byte((val >> (i * 8)) & 0xff)
	}
	return r
}

func btoi64(val []byte) uint64 {
	r := uint64(0)
	for i := uint64(0); i < 8; i++ {
		r |= uint64(val[i]) << (8 * i)
	}
	return r
}

func btoi32(val []byte) uint32 {
	r := uint32(0)
	for i := uint32(0); i < 4; i++ {
		r |= uint32(val[i]) << (8 * i)
	}
	return r
}

func btoi16(val []byte) uint16 {
	r := uint16(0)
	for i := uint16(0); i < 2; i++ {
		r |= uint16(val[i]) << (8 * i)
	}
	return r
}
