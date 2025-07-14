package main

import (
	"encoding/json"
	"fmt"
	_ "net/http/pprof"
	"os"

	agent "serverless-hosted-runner/agent"
	common "serverless-hosted-runner/common"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type RunnerConfig struct {
	Size   string `json:"size"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
	Labels string `json:"labels"`
}

func insert_pool(msg_body string, mns_url string, access_key string, access_sec string) error {
	qAgent := agent.CreateAliMNSAgent(mns_url, access_key, access_sec, agent.DEFAULT_POOL_Q, nil, nil)

	qAgent.NotifyAgent(msg_body)
	return nil
}

func runner_config() {
	config := RunnerConfig{}
	err := json.Unmarshal([]byte(os.Getenv("SLS_RUNNER_CONFIG")), &config)
	if err != nil {
		logrus.Errorln("convert runner config to struct failure,", err)
	}
	os.Setenv("SLS_RUNNER_SIZE", config.Size)
	os.Setenv("SLS_RUNNER_CPU", config.CPU)
	os.Setenv("SLS_RUNNER_MEMORY", config.Memory)
	os.Setenv("SLS_RUNNER_LABELS", config.Labels)
}

func main() {
	runner_config()
	crypto := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))
	msg := common.PoolMsg{os.Getenv("SLS_RUNNER_TYPE"), os.Getenv("SLS_RUNNER_NAME"), crypto.EncryptMsg(os.Getenv("SLS_RUNNER_PAT")),
		os.Getenv("SLS_RUNNER_URL"), os.Getenv("SLS_RUNNER_SIZE"), os.Getenv("SLS_TENANT_KEY"), os.Getenv("SLS_TENANT_SECRET"),
		os.Getenv("SLS_TENANT_REGION"), os.Getenv("SLS_SG_ID"), os.Getenv("SLS_VSWITCH_ID"), os.Getenv("SLS_RUNNER_CPU"),
		os.Getenv("SLS_RUNNER_MEMORY"), "", "", os.Getenv("SLS_RUNNER_LABELS"), "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", ""}
	out, err := json.Marshal(msg)
	if err != nil {
		logrus.Errorln("convert struct to byte fail,", err)
	}
	fmt.Println("msg body:", string(out))
	error := insert_pool(string(out), os.Getenv("SLS_MNS_URL"), os.Getenv("SLS_ALI_KEY"), os.Getenv("SLS_ALI_SEC"))
	if err != nil {
		logrus.Errorln("insert msg fail,", error)
	}
}
