//go:build debug

package main

import "log"

// AMD 5900x
func Fake5900x() {
	log.Println("AMD 5900x")
	size := 0x20 * 64
	cpuSet := make([]SYSTEM_CPU_SET_INFORMATION, size)
	cpuSet[0].Size = 32
	var lastCoreIndex byte
	var count = 768
	var index = 0x100
	for i := 0; i < count; i++ {
		cs := cpuSet[i].CpuSet()
		cs.Id = uint32(index + i)
		cs.LogicalProcessorIndex = byte(i)
		if i%2 != 0 {
			cs.CoreIndex = lastCoreIndex
		} else {
			cs.CoreIndex = byte(i)
			lastCoreIndex = byte(i)
		}
		if i > 11 {
			cs.LastLevelCacheIndex = 12
		}
	}
	SystemCpuSets = cpuSet[:count]
}

// Intel i9 13900
func Fake13900() {
	log.Println("Intel i9 13900")
	size := 0x20 * 64
	cpuSet := make([]SYSTEM_CPU_SET_INFORMATION, size)
	cpuSet[0].Size = 32
	var lastCoreIndex byte
	var count = 1024
	var index = 0x100
	for i := 0; i < count; i++ {
		cs := cpuSet[i].CpuSet()
		cs.Id = uint32(index + i)
		cs.LogicalProcessorIndex = byte(i)
		if i < 16 && i%2 != 0 {
			cs.CoreIndex = lastCoreIndex
		} else {
			cs.CoreIndex = byte(i)
			lastCoreIndex = byte(i)
		}
		if i < 16 {
			cs.EfficiencyClass = 1
		}
	}
	SystemCpuSets = cpuSet[:count]
}

func Fake8Threads() {
	log.Println("Fake8Threads")
	size := 0x20 * 64
	cpuSet := make([]SYSTEM_CPU_SET_INFORMATION, size)
	cpuSet[0].Size = 32
	var count = 256
	var index = 0x100
	for i := 0; i < count; i++ {
		cs := cpuSet[i].CpuSet()
		cs.Id = uint32(index + i)
		cs.LogicalProcessorIndex = byte(i)
		cs.CoreIndex = byte(i)
	}
	SystemCpuSets = cpuSet[:count]
}

// ?
func FakeNumaCCD12Core() {
	log.Println("FakeNumaCCD12Core")
	size := 0x20 * 64
	cpuSet := make([]SYSTEM_CPU_SET_INFORMATION, size)
	cpuSet[0].Size = 32
	var count = 384
	var index = 0x100
	for i := 0; i < count; i++ {
		cs := cpuSet[i].CpuSet()
		cs.Id = uint32(index + i)
		cs.LogicalProcessorIndex = byte(i)
		cs.CoreIndex = byte(i)

		if i > 5 {
			cs.LastLevelCacheIndex = 6
			cs.NumaNodeIndex = 6
		}
	}
	SystemCpuSets = cpuSet[:count]
}

// ?
func Fake2CCD12CoreHT() {
	log.Println("Fake2CCD12CoreHT")
	size := 0x20 * 64
	cpuSet := make([]SYSTEM_CPU_SET_INFORMATION, size)
	cpuSet[0].Size = 32
	var lastCoreIndex byte
	var count = 768
	var index = 0x100
	for i := 0; i < count; i++ {
		cs := cpuSet[i].CpuSet()
		cs.Id = uint32(index + i)
		cs.LogicalProcessorIndex = byte(i)
		if i%2 != 0 {
			cs.CoreIndex = lastCoreIndex
		} else {
			cs.CoreIndex = byte(i)
			lastCoreIndex = byte(i)
		}

		if i > 11 {
			cs.LastLevelCacheIndex = 12
		}
	}
	SystemCpuSets = cpuSet[:count]
}
