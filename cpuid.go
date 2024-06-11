package main

import "github.com/intel-go/cpuid"

func isAMD() bool {
	return cpuid.VendorIdentificatorString == "AuthenticAMD"
}

func isIntel() bool {
	return cpuid.VendorIdentificatorString == "GenuineIntel"
}
