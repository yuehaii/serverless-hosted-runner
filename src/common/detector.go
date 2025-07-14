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
	detect_time         time.Time
	alt_heap_usage      uint64
	his_heap_usage      uint64
	os_mem_usage        uint64
	gc_cycle            uint32
	threshold_mb_second uint32
}

func CreateMemDector() IMemoryDector {
	return &MemoryDector{time.Time{}, 0, 0, 0, 0, 5}
}

func CreateCpuDector() IDetector {
	return nil
}

func (d *MemoryDector) DetectUsage() {
	d.memUsage()
}

func (d MemoryDector) CompareLast() {
	last_t := d.detect_time
	last_allocate := d.alt_heap_usage
	last_history := d.his_heap_usage
	last_system := d.os_mem_usage
	last_gc_cycles := d.gc_cycle
	d.memUsage()
	fmt.Printf("Time span: %s - %s", last_t.String(), d.detect_time.String())
	fmt.Printf("\tRuntime Allocated Heap Increased %v MiB", d.byteToMega(d.alt_heap_usage-last_allocate))
	fmt.Printf("\tHistory Allocated Heap Increased %v MiB", d.byteToMega(d.his_heap_usage-last_history))
	fmt.Printf("\tSystem Allocated Increased %v MiB", d.byteToMega(d.os_mem_usage-last_system))
	fmt.Printf("\tCompleted GC Cycles Increased %v\n", d.gc_cycle-last_gc_cycles)
	d.needGc(time.Since(last_t), d.byteToMega(d.alt_heap_usage-last_allocate))
}

func (d *MemoryDector) memUsage() {
	var mstate runtime.MemStats
	runtime.ReadMemStats(&mstate)
	d.detect_time = time.Now()
	d.alt_heap_usage = mstate.Alloc
	d.his_heap_usage = mstate.TotalAlloc
	d.os_mem_usage = mstate.Sys
	d.gc_cycle = mstate.NumGC
	fmt.Printf("%s, Runtime Allocated Heap = %v MiB", d.detect_time.String(), d.byteToMega(mstate.Alloc))
	fmt.Printf("\tHistory Allocated Heap = %v MiB", d.byteToMega(mstate.TotalAlloc))
	fmt.Printf("\tSystem Allocated = %v MiB", d.byteToMega(mstate.Sys))
	fmt.Printf("\tCompleted GC Cycles = %v\n", mstate.NumGC)
}

func (d MemoryDector) needGc(duration time.Duration, usage_mb uint64) {
	if usage_mb/uint64(duration.Seconds()) > uint64(d.threshold_mb_second) {
		fmt.Printf("Mem usage %vMb/s, it exceed %vMb/s threshold and trigger GC.",
			usage_mb/uint64(duration.Seconds()), d.threshold_mb_second)
		runtime.GC()
	}
}

func (d MemoryDector) byteToMega(data uint64) uint64 {
	return data / 1024 / 1024
}
