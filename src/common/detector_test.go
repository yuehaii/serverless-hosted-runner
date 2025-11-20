package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemDetector(t *testing.T) {
	mem := CreateMemDector()
	mem.DetectUsage()
	_ = allocateLargeMem(5)
	mem.CompareLast()
	assert.Nil(t, mem)
}

func allocateLargeMem(max int) [][]int {
	var memStore [][]int
	for times := 0; times < max; times++ {
		a := make([]int, 0, 999999)
		memStore = append(memStore, a)
	}
	return memStore
}
