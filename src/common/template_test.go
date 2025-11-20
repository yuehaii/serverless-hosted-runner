package common

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTernary(t *testing.T) {
	val1, val2 := 1, 2
	val := Ternary(val1 > val2, val1, val2)
	assert.Equal(t, val, 2)

	runnerRepoURL := "https://git.build.ingka.ikea.com/api/v3/qweqeqweqewq"
	enID := "git.build.ingka.ikea.com"
	entokenFqdn := "https://git.build.ingka.ikea.com/api/v3/"
	tokenFqdn := "https://api.github.com/"
	pref := Ternary(strings.Contains(runnerRepoURL, enID), entokenFqdn, tokenFqdn).(string)
	assert.Equal(t, pref, entokenFqdn)

	tStr := "Pool"
	v := Ternary(tStr == "Pool", int64(60), int64(5)).(int64)
	assert.Equal(t, v, 60)
}

func TestTernaryComparable(t *testing.T) {
	assert.Equal(t, TernaryComparable(true, nil, nil, fmt.Errorf("invalid")), nil)

	assert.Equal(t, TernaryComparable(false, nil, fmt.Errorf("invalid")), fmt.Errorf("invalid"))

	assert.Equal(t, TernaryComparable(true, nil, fmt.Errorf("invalid")), nil)

	assert.Equal(t, TernaryComparable(true, 2, 3.0), 2)

	assert.Equal(t, TernaryComparable(false, "this", "is a test", "my type", "is", "string"), "is a test")

	// cannot infer T and raise error during build. it is safe to prevent runtime panic
	// assert.Equal(t, TernaryComparable(true, nil, nil, nil), nil)
}
