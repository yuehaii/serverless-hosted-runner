package common

import (
	"fmt"
	"runtime"
	"time"
)

type IDetector interface {
	DetectUsage()
}

type IMemoryDector interface {
	IDetector
	CompareLast()
}

type MemoryDector struct {
	detectTime        time.Time
	altHeapUsage      uint64
	hisHeapUsage      uint64
	osMemUsage        uint64
	gcCycle           uint32
	thresholdMbSecond uint32
}

func CreateMemDector() IMemoryDector {
	return &MemoryDector{time.Time{}, 0, 0, 0, 0, 5}
}

func CreateCPUDector() IDetector {
	return nil
}

func (d *MemoryDector) DetectUsage() {
	d.memUsage()
}

func (d MemoryDector) CompareLast() {
	lastTime := d.detectTime
	lastAllocate := d.altHeapUsage
	lastHistory := d.hisHeapUsage
	lastSystem := d.osMemUsage
	lastGcCycles := d.gcCycle
	d.memUsage()
	fmt.Printf("Time span: %s - %s", lastTime.String(), d.detectTime.String())
	fmt.Printf("\tRuntime Allocated Heap Increased %v MiB", d.byteToMega(d.altHeapUsage-lastAllocate))
	fmt.Printf("\tHistory Allocated Heap Increased %v MiB", d.byteToMega(d.hisHeapUsage-lastHistory))
	fmt.Printf("\tSystem Allocated Increased %v MiB", d.byteToMega(d.osMemUsage-lastSystem))
	fmt.Printf("\tCompleted GC Cycles Increased %v\n", d.gcCycle-lastGcCycles)
	d.needGc(time.Since(lastTime), d.byteToMega(d.altHeapUsage-lastAllocate))
}

func (d *MemoryDector) memUsage() {
	var mstate runtime.MemStats
	runtime.ReadMemStats(&mstate)
	d.detectTime = time.Now()
	d.altHeapUsage = mstate.Alloc
	d.hisHeapUsage = mstate.TotalAlloc
	d.osMemUsage = mstate.Sys
	d.gcCycle = mstate.NumGC
	fmt.Printf("%s, Runtime Allocated Heap = %v MiB", d.detectTime.String(), d.byteToMega(mstate.Alloc))
	fmt.Printf("\tHistory Allocated Heap = %v MiB", d.byteToMega(mstate.TotalAlloc))
	fmt.Printf("\tSystem Allocated = %v MiB", d.byteToMega(mstate.Sys))
	fmt.Printf("\tCompleted GC Cycles = %v\n", mstate.NumGC)
}

func (d MemoryDector) needGc(duration time.Duration, usageMb uint64) {
	if usageMb/uint64(duration.Seconds()) > uint64(d.thresholdMbSecond) {
		fmt.Printf("Mem usage %vMb/s, it exceed %vMb/s threshold and trigger GC.",
			usageMb/uint64(duration.Seconds()), d.thresholdMbSecond)
		runtime.GC()
	}
}

func (d MemoryDector) byteToMega(data uint64) uint64 {
	return data / 1024 / 1024
}
