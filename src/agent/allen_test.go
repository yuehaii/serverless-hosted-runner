package agent

import (
	"testing"
)

func TestAllenAgent(t *testing.T) {
	aln := CreateAllenStoreAgent(nil)
	aln.InitAgent()
	aln.MonitorOnAgent()
}
