package main

import (
	"log"
	"strconv"
	"time"

	"golang.org/x/sys/windows/registry"
)

type Device struct {
	reg                 registry.Key
	IrqPolicy           int32
	DeviceDesc          string
	DeviceIDs           []string
	DevObjName          string
	Driver              string
	LocationInformation string
	Regpath             string
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
	InterrupTypeMap    Bits
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
var CPUArray []string
var CPUBits []Bits
var InterrupTypeMap = map[Bits]string{
	0: "unknown",
	1: "LineBased",
	2: "Msi",

	4: "MsiX",
}
var sysInfo SystemInfo

const ZeroBit = Bits(0)

func Set(b, flag Bits) Bits    { return b | flag }
func Clear(b, flag Bits) Bits  { return b &^ flag }
func Toggle(b, flag Bits) Bits { return b ^ flag }
func Has(b, flag Bits) bool    { return b&flag != 0 }

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	sysInfo = GetSystemInfo()
	CPUMap = make(map[Bits]string, sysInfo.NumberOfProcessors)
	var index Bits = 1
	for i := 0; i < int(sysInfo.NumberOfProcessors); i++ {
		name := "CPU" + strconv.Itoa(i)
		CPUMap[index] = name
		CPUArray = append(CPUArray, name)
		CPUBits = append(CPUBits, index)
		index *= 2
	}
}
