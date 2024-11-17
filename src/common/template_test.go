package common 

import( 
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
)

func TestTernary(t *testing.T) {
	val1, val2 := 1, 2
	val := Ternary(val1 > val2, val1, val2) 
	assert.Equal(t, val, 2) 

	runner_repo_url := "https://git.build.ingka.ikea.com/api/v3/qweqeqweqewq"
	en_id := "git.build.ingka.ikea.com"
	entoken_fqdn := "https://git.build.ingka.ikea.com/api/v3/"
	token_fqdn := "https://api.github.com/"
	pref := Ternary(strings.Contains(runner_repo_url, en_id), entoken_fqdn, token_fqdn).(string) 
	assert.Equal(t, pref, entoken_fqdn) 

	t_str := "Pool"
	v := Ternary(t_str == "Pool", int64(60), int64(5)).(int64)
	assert.Equal(t, v, 60) 
}