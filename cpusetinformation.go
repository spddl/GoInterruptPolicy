package main

import (
	"log"
	"math"

	"golang.org/x/sys/windows"
)

var SystemCpuSets = []SYSTEM_CPU_SET_INFORMATION{}

const (
	ToolTipTextNumaNode        = "A group-relative value indicating which NUMA node a CPU Set is on. All CPU Sets in a given group that are on the same NUMA node will have the same value for this field."
	ToolTipTextLastLevelCache  = "A group-relative value indicating which CPU Sets share at least one level of cache with each other. This value is the same for all CPU Sets in a group that are on processors that share cache with each other."
	ToolTipTextEfficiencyClass = "A value indicating the intrinsic energy efficiency of a processor for systems that support heterogeneous processors (such as ARM big.LITTLE systems). CPU Sets with higher numerical values of this field have home processors that are faster but less power-efficient than ones with lower values."
)

type CpuSets struct {
	HyperThreading    bool
	CoreCount         int
	MaxThreadsPerCore int
	NumaNode          bool // A group-relative value indicating which NUMA node a CPU Set is on. All CPU Sets in a given group that are on the same NUMA node will have the same value for this field.
	LastLevelCache    bool // A group-relative value indicating which CPU Sets share at least one level of cache with each other. This value is the same for all CPU Sets in a group that are on processors that share cache with each other.
	EfficiencyClass   bool // A value indicating the intrinsic energy efficiency of a processor for systems that support heterogeneous processors (such as ARM big.LITTLE systems). CPU Sets with higher numerical values of this field have home processors that are faster but less power-efficient than ones with lower values.
	CPU               []CpuSet
	Layout            struct {
		Rows int
		Cols int
	}
}

type CpuSet struct {
	Id                    uint32
	CoreIndex             byte
	LogicalProcessorIndex byte
	LastLevelCacheIndex   byte // A group-relative value indicating which CPU Sets share at least one level of cache with each other. This value is the same for all CPU Sets in a group that are on processors that share cache with each other.
	EfficiencyClass       byte // A value indicating the intrinsic energy efficiency of a processor for systems that support heterogeneous processors (such as ARM big.LITTLE systems). CPU Sets with higher numerical values of this field have home processors that are faster but less power-efficient than ones with lower values.
	NumaNodeIndex         byte // A group-relative value indicating which NUMA node a CPU Set is on. All CPU Sets in a given group that are on the same NUMA node will have the same value for this field.
}

func getLayout(total int) (bestCols, bestRows int) {
	ratio := 16.0 / 9.0
	bestCols, bestRows = 1, total
	minDiff := math.MaxFloat64

	for cols := 1; cols <= total; cols++ {
		rows := total / cols
		if cols*rows != total {
			continue
		}
		currentRatio := float64(cols) / float64(rows)
		diff := math.Abs(currentRatio - ratio)

		if diff < minDiff {
			bestCols, bestRows = cols, rows
			minDiff = diff
		}
	}

	return
}

func (cs *CpuSets) Init() {
	size := 0x20 * 64
	SystemCpuSets = make([]SYSTEM_CPU_SET_INFORMATION, size)

	var length uint32
	var hProcess windows.Handle
	success := GetSystemCpuSetInformation(&SystemCpuSets[0], uint32(size), &length, uintptr(hProcess), 0)
	if success != 1 {
		log.Println("err")
	} else {
		SystemCpuSets = SystemCpuSets[:length]
	}

	/// debug
	// Fake13900()
	// Fake5900x()
	// Fake8Threads()
	// FakeNumaCCD12Core()
	// Fake2CCD12CoreHT()

	cs.CoreCount = int(uint32(len(SystemCpuSets)) / SystemCpuSets[0].Size)
	var lastEfficiencyClass, lastLevelCache, lastNumaNodeIndex byte
	var LogicalCores int

	for i := 0; i < cs.CoreCount; i++ {
		cpu := SystemCpuSets[i].CpuSet()
		if i == 0 { // The EfficiencyClass starts with 1 on the Intel Gen12+
			lastEfficiencyClass = cpu.EfficiencyClass
		}

		cs.CPU = append(cs.CPU, CpuSet{
			Id:                    cpu.Id,
			CoreIndex:             cpu.CoreIndex,
			LogicalProcessorIndex: cpu.LogicalProcessorIndex,
			EfficiencyClass:       cpu.EfficiencyClass,
			LastLevelCacheIndex:   cpu.LastLevelCacheIndex,
			NumaNodeIndex:         cpu.NumaNodeIndex,
		})

		// fmt.Printf("(%02d) [%d/%x] %02d/%02d Eff%d CCD%d NUMA%d\n", i, cpu.Id, cpu.Id, cpu.CoreIndex, cpu.LogicalProcessorIndex, cpu.EfficiencyClass, cpu.LastLevelCacheIndex, cpu.NumaNodeIndex)

		if cpu.CoreIndex != cpu.LogicalProcessorIndex {
			if !cs.HyperThreading {
				cs.HyperThreading = true
			}
			if cs.MaxThreadsPerCore < int(cpu.LogicalProcessorIndex-cpu.CoreIndex) {
				cs.MaxThreadsPerCore = int(cpu.LogicalProcessorIndex - cpu.CoreIndex)
			}
		} else {
			LogicalCores++
		}

		if !cs.EfficiencyClass && lastEfficiencyClass != cpu.EfficiencyClass {
			cs.EfficiencyClass = true
		}

		if !cs.LastLevelCache && lastLevelCache != cpu.LastLevelCacheIndex {
			cs.LastLevelCache = true
		}

		if !cs.NumaNode && lastNumaNodeIndex != cpu.NumaNodeIndex {
			cs.NumaNode = true
		}

		lastEfficiencyClass = cpu.EfficiencyClass
		lastLevelCache = cpu.LastLevelCacheIndex
		lastNumaNodeIndex = cpu.NumaNodeIndex
	}

	cs.Layout.Cols, cs.Layout.Rows = getLayout(LogicalCores)
}
