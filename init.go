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
	DevicePolicy          int32
	DevicePriority        int32
	AssignmentSetOverride Bits

	// MessageSignaledInterruptProperties
	MessageNumberLimit string
	MsiSupported       int32
}

const (
	// https://docs.microsoft.com/en-us/windows-hardware/drivers/kernel/interrupt-affinity-and-priority
	IrqPolicyMachineDefault                    = iota // 0x00
	IrqPolicyAllCloseProcessors                = iota // 0x01
	IrqPolicyOneCloseProcessor                 = iota // 0x02
	IrqPolicyAllProcessorsInMachine            = iota // 0x03
	IrqPolicySpecifiedProcessors               = iota // 0x04
	IrqPolicySpreadMessagesAcrossAllProcessors = iota // 0x05
)

type Bits uint64

var CPUMap map[Bits]string
var CPUArray []string
var CPUBits []Bits
var sysInfo SystemInfo

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
