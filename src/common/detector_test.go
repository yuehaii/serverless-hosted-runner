package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemDetector(t *testing.T) {
	mem := CreateMemDector()
	mem.DetectUsage()
	allocate_large_mem(5)
	mem.CompareLast()
	assert.Nil(t, mem)
}

func allocate_large_mem(max int) {
	var mem_store [][]int
	for times := 0; times < max; times++ {
		a := make([]int, 0, 999999)
		mem_store = append(mem_store, a)
	}
}
