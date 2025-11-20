package common

import (
	"testing"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAnyChange(t *testing.T) {
	msg := PoolMsg{}
	msg.Type = "Repo"
	msg.Name = "cdh-br-ark-impl-alirdspg-ccoecn"
	msg.Pat = ""
	msg.URL = "https://github.com/ingka-group-digital" + "/" + "cdh-br-ark-impl-alirdspg-ccoecn"
	msg.Size = "1"
	msg.Key = "LTAI5t78PSxiJN2XvUAmC9mx"
	msg.Secret = "1yddiLE931hWrT6TXKt24zMGIJGzswssM"
	msg.Region = "cn-shanghai"
	msg.SecGpID = "sg-uf69bmdfil0cobd9uzd9"
	msg.VSwitchID = "vsw-uf6j151zxg7t4t3u69xig"
	msg.CPU = "1.0"
	msg.Memory = "2.0"
	msg.Repos = "cdh-br-ark-impl-alirdspg-ccoecn,serverless-hosted-runner"
	st := EnvStore(&msg, msg.Name, "cdh-br-ark-impl-alirdspg-ccoecn")
	st.Save()
	key, runnerType := st.GetKey()
	pat, patType := st.GetPat()
	logrus.Infof("lazy_registration, key: %s, pat: %s, runnerType %s, patType %s", key, pat, runnerType, patType)
	assert.NotEqual(t, key, "")

	msg2 := PoolMsg{}
	msg2.Type = "Org"
	msg2.Name = "otestname"
	msg2.Pat = "o1231231312"
	msg2.URL = "http://test.com/123o2"
	msg2.Size = "32"
	msg2.Key = "oqweqweqeq2"
	msg2.Secret = "o89988777887772"
	msg2.Region = "ocn-shanghai2"
	msg2.SecGpID = "oa-sdasdasda2"
	msg2.VSwitchID = "oz-xczxcxczxc2"
	msg2.CPU = "3.0"
	msg2.Memory = "5.0"
	st2 := EnvStore(&msg2, msg2.Name, msg2.Name)
	st2.Save()

	isSame := st2.AnyChange()
	assert.Equal(t, isSame, true)

	msg3 := PoolMsg{}
	msg3.Type = "Org"
	msg3.Name = "otestname"
	msg3.Pat = "o1231231312"
	msg3.URL = "http://test.com/123o2"
	msg3.Size = "32"
	msg3.Key = "oqweqweqeq2"
	msg3.Secret = "o89988777887772"
	msg3.Region = "ocn-shanghai2"
	msg3.SecGpID = "oa-sdasdasda2"
	msg3.VSwitchID = "oz-xczxcxczxc2"
	msg3.CPU = "3.0"
	msg3.Memory = "5.0"
	st3 := EnvStore(&msg3, msg3.Name, msg3.Name)
	st3.Save()

	isSame3 := st3.AnyChange()
	assert.Equal(t, isSame3, false)
}

func TestStoreOrg(t *testing.T) {
	msg := PoolMsg{}
	msg.Type = "Org"
	msg.Name = "otestname"
	msg.Pat = "o123123131"
	msg.URL = "http://test.com/123o"
	msg.Size = "3"
	msg.Key = "oqweqweqeq"
	msg.Secret = "o8998877788777"
	msg.Region = "ocn-shanghai"
	msg.SecGpID = "oa-sdasdasda"
	msg.VSwitchID = "oz-xczxcxczxc"
	msg.CPU = "2.0"
	msg.Memory = "4.0"
	st := EnvStore(&msg, "", "")
	st.Save()
	store := EnvStore(nil, msg.Name, "")
	key, _ := store.GetKey()
	sec, _ := store.GetSecret()
	region, _ := store.GetRegion()
	secGpID, _ := store.GetSecGpID()
	vswitchID, _ := store.GetVSwitchID()
	url, _ := store.GetURL()
	size, _ := store.GetSize()
	cpu, _ := store.GetCPU()
	mem, _ := store.GetMemory()
	pat, runnerType := store.GetPat()
	assert.Equal(t, runnerType, "org")
	assert.Equal(t, key, msg.Key)
	assert.Equal(t, sec, msg.Secret)
	assert.Equal(t, region, msg.Region)
	assert.Equal(t, secGpID, msg.SecGpID)
	assert.Equal(t, vswitchID, msg.VSwitchID)
	assert.Equal(t, pat, msg.Pat)
	assert.Equal(t, url, msg.URL)
	assert.Equal(t, size, msg.Size)
	assert.Equal(t, cpu, msg.CPU)
	assert.Equal(t, mem, msg.Memory)

	msg2 := PoolMsg{}
	msg2.Type = "Org"
	msg2.Name = "otestname"
	msg2.Pat = "o1231231312"
	msg2.URL = "http://test.com/123o2"
	msg2.Size = "32"
	msg2.Key = "oqweqweqeq2"
	msg2.Secret = "o89988777887772"
	msg2.Region = "ocn-shanghai2"
	msg2.SecGpID = "oa-sdasdasda2"
	msg2.VSwitchID = "oz-xczxcxczxc2"
	st2 := EnvStore(&msg2, msg2.Name, msg2.Name) // need org name and repo name to check the prev key val
	st2.Save()
	store2 := EnvStore(nil, msg2.Name, "")
	key2 := store2.GetPreKey()
	sec2 := store2.GetPreSecret()
	region2 := store2.GetPreRegion()
	secGpID2 := store2.GetPreSecGpID()
	vswitchID2 := store2.GetPreVSwitchID()
	url2 := store2.GetPreURL()
	size2 := store2.GetPreSize()
	pat2 := store2.GetPrePat()
	assert.Equal(t, key2, msg.Key)
	assert.Equal(t, sec2, msg.Secret)
	assert.Equal(t, region2, msg.Region)
	assert.Equal(t, secGpID2, msg.SecGpID)
	assert.Equal(t, vswitchID2, msg.VSwitchID)
	assert.Equal(t, pat2, msg.Pat)
	assert.Equal(t, url2, msg.URL)
	assert.Equal(t, size2, msg.Size)
	key3, _ := store2.GetKey()
	sec3, _ := store2.GetSecret()
	region3, _ := store2.GetRegion()
	secGpID3, _ := store2.GetSecGpID()
	vswitchID3, _ := store2.GetVSwitchID()
	url3, _ := store2.GetURL()
	size3, _ := store2.GetSize()
	pat3, _ := store2.GetPat()
	assert.Equal(t, key3, msg2.Key)
	assert.Equal(t, sec3, msg2.Secret)
	assert.Equal(t, region3, msg2.Region)
	assert.Equal(t, secGpID3, msg2.SecGpID)
	assert.Equal(t, vswitchID3, msg2.VSwitchID)
	assert.Equal(t, pat3, msg2.Pat)
	assert.Equal(t, url3, msg2.URL)
	assert.Equal(t, size3, msg2.Size)
}

func TestNoneTkRepo(t *testing.T) {
	msg := PoolMsg{}
	msg.Type = "Pool"
	msg.Name = "testname"
	msg.Pat = "asqqwe"
	msg.URL = "http://git.build.ingka.ikea.com/123"
	msg.Size = "3"
	msg.Key = "qweqweqeq"
	msg.Secret = "8998877788777"
	msg.Region = "cn-shanghai"
	msg.SecGpID = "a-sdasdasda"
	msg.VSwitchID = "z-xczxcxczxc"
	// if len(msg.Pat) > 0 { return }
	st := EnvStore(&msg, "", "")
	st.Save()
	// store := EnvStore(nil, "", msg.Name)
	store := EnvStore(nil, msg.Name, "")
	key, _ := store.GetKey()
	sec, _ := store.GetSecret()
	region, _ := store.GetRegion()
	secGpID, _ := store.GetSecGpID()
	vswitchID, _ := store.GetVSwitchID()
	pat, runnerType := store.GetPat()
	assert.Equal(t, runnerType, "repo")
	assert.Equal(t, key, msg.Key)
	assert.Equal(t, sec, msg.Secret)
	assert.Equal(t, region, msg.Region)
	assert.Equal(t, secGpID, msg.SecGpID)
	assert.Equal(t, vswitchID, msg.VSwitchID)
	assert.Equal(t, pat, msg.Pat)
}

func TestStoreRepo(t *testing.T) {
	msg := PoolMsg{}
	msg.Type = "Repo"
	msg.Name = "testname"
	msg.Pat = "123123131"
	msg.URL = "http://test.com/123"
	msg.Size = "3"
	msg.Key = "qweqweqeq"
	msg.Secret = "8998877788777"
	msg.Region = "cn-shanghai"
	msg.SecGpID = "a-sdasdasda"
	msg.VSwitchID = "z-xczxcxczxc"
	st := EnvStore(&msg, "", "")
	st.Save()
	store := EnvStore(nil, "", msg.Name)
	key, _ := store.GetKey()
	sec, _ := store.GetSecret()
	region, _ := store.GetRegion()
	secGpID, _ := store.GetSecGpID()
	vswitchID, _ := store.GetVSwitchID()
	pat, runnerType := store.GetPat()
	assert.Equal(t, runnerType, "repo")
	assert.Equal(t, key, msg.Key)
	assert.Equal(t, sec, msg.Secret)
	assert.Equal(t, region, msg.Region)
	assert.Equal(t, secGpID, msg.SecGpID)
	assert.Equal(t, vswitchID, msg.VSwitchID)
	assert.Equal(t, pat, msg.Pat)
}
