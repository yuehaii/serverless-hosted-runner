package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolvers(t *testing.T) {
	sys := CreateUnixSysCtl()
	addr := sys.Addr()
	sys.SetResolvers()
	sys.NetworkConnectivity()
	assert.Nil(t, addr)
}
