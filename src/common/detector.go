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
	d.mem_usage()
}

func (d MemoryDector) CompareLast() {
	last_t := d.detect_time
	last_allocate := d.alt_heap_usage
	last_history := d.his_heap_usage
	last_system := d.os_mem_usage
	last_gc_cycles := d.gc_cycle
	d.mem_usage()
	fmt.Printf("Time span: %s - %s", last_t.String(), d.detect_time.String())
	fmt.Printf("\tRuntime Allocated Heap Increased %v MiB", d.byte_to_mega(d.alt_heap_usage-last_allocate))
	fmt.Printf("\tHistory Allocated Heap Increased %v MiB", d.byte_to_mega(d.his_heap_usage-last_history))
	fmt.Printf("\tSystem Allocated Increased %v MiB", d.byte_to_mega(d.os_mem_usage-last_system))
	fmt.Printf("\tCompleted GC Cycles Increased %v\n", d.gc_cycle-last_gc_cycles)
	d.need_gc(time.Since(last_t), d.byte_to_mega(d.alt_heap_usage-last_allocate))
}

func (d *MemoryDector) mem_usage() {
	var mstate runtime.MemStats
	runtime.ReadMemStats(&mstate)
	d.detect_time = time.Now()
	d.alt_heap_usage = mstate.Alloc
	d.his_heap_usage = mstate.TotalAlloc
	d.os_mem_usage = mstate.Sys
	d.gc_cycle = mstate.NumGC
	fmt.Printf("%s, Runtime Allocated Heap = %v MiB", d.detect_time.String(), d.byte_to_mega(mstate.Alloc))
	fmt.Printf("\tHistory Allocated Heap = %v MiB", d.byte_to_mega(mstate.TotalAlloc))
	fmt.Printf("\tSystem Allocated = %v MiB", d.byte_to_mega(mstate.Sys))
	fmt.Printf("\tCompleted GC Cycles = %v\n", mstate.NumGC)
}

func (d MemoryDector) need_gc(duration time.Duration, usage_mb uint64) {
	if usage_mb/uint64(duration.Seconds()) > uint64(d.threshold_mb_second) {
		fmt.Printf("Mem usage %vMb/s, it exceed %vMb/s threshold and trigger GC.",
			usage_mb/uint64(duration.Seconds()), d.threshold_mb_second)
		runtime.GC()
	}
}

func (d MemoryDector) byte_to_mega(data uint64) uint64 {
	return data / 1024 / 1024
}
