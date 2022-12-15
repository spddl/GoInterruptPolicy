package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagDevObjName         string
	flagDevicePriority     int
	flagDevicePolicy       int
	flagMsiSupported       int
	flagMessageNumberLimit int
	flagCPU                string
	flagRestart            bool
	flagHelp               bool

	CLIMode bool
)

func init() {
	flag.StringVar(&flagDevObjName, "devobj", "", "\\Device\\00000123")
	flag.StringVar(&flagCPU, "cpu", "", "e.g. 0,1,2,4")
	flag.IntVar(&flagDevicePriority, "priority", -1, "0=Undefined, 1=Low, 2=Normal, 3=High")
	flag.IntVar(&flagDevicePolicy, "policy", -1, "0=Default, 1=All Close Proc, 2=One Close Proc, 3=All Proc in Machine, 4=Specified Proc, 5=Spread Messages Across All Proc")
	flag.IntVar(&flagMsiSupported, "msisupported", -1, "0=Off, 1=On")
	flag.IntVar(&flagMessageNumberLimit, "msilimit", -1, "Message Signaled Interrupt Limit")
	flag.BoolVar(&flagRestart, "restart", false, "Restart target device")
	flag.BoolVar(&flagHelp, "help", false, "Print Defaults")

	flag.Parse()
	if flagHelp {
		fmt.Printf("Usage: %s [OPTIONS] argument ...\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	if flagDevObjName != "" || flagDevicePriority != -1 || flagDevicePolicy != -1 || flagMsiSupported != -1 || flagMessageNumberLimit != -1 || flagRestart {
		CLIMode = true
	}

	if flagDevObjName != "" {
		fmt.Println("DevObjName:", flagDevObjName)
	}
	if flagDevicePriority != -1 {
		var prio string
		switch flagDevicePriority {
		case 0:
			prio = "Undefined"
		case 1:
			prio = "Low"
		case 2:
			prio = "Normal"
		case 3:
			prio = "High"
		default:
			prio = fmt.Sprintf("%d", flagDevicePriority)
		}
		fmt.Println("DevicePriority:", prio)
	}
	if flagDevicePolicy != -1 {
		var policy string
		switch flagDevicePolicy {
		case IrqPolicyMachineDefault: // 0x00
			policy = "Default"
		case IrqPolicyAllCloseProcessors: // 0x01
			policy = "All Close Proc"
		case IrqPolicyOneCloseProcessor: // 0x02
			policy = "One Close Proc"
		case IrqPolicyAllProcessorsInMachine: // 0x03
			policy = "All Proc in Machine"
		case IrqPolicySpecifiedProcessors: // 0x04
			policy = "Specified Proc"
		case IrqPolicySpreadMessagesAcrossAllProcessors: // 0x05
			policy = "Spread Messages Across All Proc"
		default:
			policy = fmt.Sprintf("%d", flagDevicePolicy)
		}
		fmt.Println("DevicePolicy:", policy)
	}
}
