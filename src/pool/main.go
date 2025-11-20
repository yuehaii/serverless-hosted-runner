// Package pool mode
package main

import (
	"encoding/json"
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

func insertPool(msgBody string, mnsURL string, accessKey string, accessSec string) error {
	qAgent := agent.CreateAliMNSAgent(mnsURL, accessKey, accessSec, agent.DefaultPoolQueue, nil, nil)

	qAgent.NotifyAgent(msgBody)
	return nil
}

func runnerConfig() {
	config := RunnerConfig{}
	err := json.Unmarshal([]byte(os.Getenv("SLS_RUNNER_CONFIG")), &config)
	if err != nil {
		logrus.Errorln("convert runner config to struct failure,", err)
	}
	if err := os.Setenv("SLS_RUNNER_SIZE", config.Size); err != nil {
		logrus.Errorln("fail to set config runner Size env,", err)
	}
	if err := os.Setenv("SLS_RUNNER_CPU", config.CPU); err != nil {
		logrus.Errorln("fail to set config runner CPU env,", err)
	}
	if err := os.Setenv("SLS_RUNNER_MEMORY", config.Memory); err != nil {
		logrus.Errorln("fail to set config runner Memory env,", err)
	}
	if err := os.Setenv("SLS_RUNNER_LABELS", config.Labels); err != nil {
		logrus.Errorln("fail to set config runner Labels env,", err)
	}
}

func main() {
	runnerConfig()
	crypto := common.DefaultCryptography(os.Getenv("SLS_ENC_KEY"))
	msg := common.PoolMsg{
		Type:                  os.Getenv("SLS_RUNNER_TYPE"),
		Name:                  os.Getenv("SLS_RUNNER_NAME"),
		Pat:                   crypto.EncryptMsg(os.Getenv("SLS_RUNNER_PAT")),
		URL:                   os.Getenv("SLS_RUNNER_URL"),
		Size:                  os.Getenv("SLS_RUNNER_SIZE"),
		Key:                   os.Getenv("SLS_TENANT_KEY"),
		Secret:                os.Getenv("SLS_TENANT_SECRET"),
		Region:                os.Getenv("SLS_TENANT_REGION"),
		SecGpID:               os.Getenv("SLS_SG_ID"),
		VSwitchID:             os.Getenv("SLS_VSWITCH_ID"),
		CPU:                   os.Getenv("SLS_RUNNER_CPU"),
		Memory:                os.Getenv("SLS_RUNNER_MEMORY"),
		Repos:                 os.Getenv("SLS_RUNNER_REPOS"),
		PullInterval:          os.Getenv("SLS_RUNNER_PULLINTERVAL"),
		Labels:                os.Getenv("SLS_RUNNER_LABELS"),
		ChargeLabels:          os.Getenv("SLS_RUNNER_CHARGELABELS"),
		RunnerGroup:           os.Getenv("SLS_RUNNER_RUNNERGROUP"),
		ArmClientID:           os.Getenv("SLS_RUNNER_ARMCLIENTID"),
		ArmClientSecret:       os.Getenv("SLS_RUNNER_ARMCLIENTSECRET"),
		ArmSubscriptionID:     os.Getenv("SLS_RUNNER_ARMSUBSCRIPTIONID"),
		ArmTenantID:           os.Getenv("SLS_RUNNER_ARMTENANTID"),
		ArmEnvironment:        os.Getenv("SLS_RUNNER_ARMENV"),
		ArmRPRegistration:     os.Getenv("SLS_RUNNER_ARMPRREGISTRATION"),
		ArmResourceGroupName:  os.Getenv("SLS_RUNNER_ARMRGNAME"),
		ArmSubnetID:           os.Getenv("SLS_RUNNER_ARMSUBNETID"),
		ArmLogAnaWorkspaceID:  os.Getenv("SLS_RUNNER_ARMLOGANAWSID"),
		ArmLogAnaWorkspaceKey: os.Getenv("SLS_RUNNER_ARMLOGANAWSKEY"),
		GcpCredential:         os.Getenv("SLS_RUNNER_GCPCREDENTIAL"),
		GcpProject:            os.Getenv("SLS_RUNNER_GCPPROJECT"),
		GcpRegion:             os.Getenv("SLS_RUNNER_GCPREGION"),
		GcpSA:                 os.Getenv("SLS_RUNNER_GCPSA"),
		GcpApikey:             os.Getenv("SLS_RUNNER_GCPAPIKEY"),
		GcpDind:               os.Getenv("SLS_RUNNER_GCPDIND"),
		GcpVpc:                os.Getenv("SLS_RUNNER_GCPVPC"),
		GcpSubnet:             os.Getenv("SLS_RUNNER_GCPSUBNET"),
		ImageVersion:          os.Getenv("SLS_RUNNER_IMAGEVERSION"),
		AciLocation:           os.Getenv("SLS_RUNNER_ACILOCATION"),
		AciSku:                os.Getenv("SLS_RUNNER_ACISKU"),
		AciNetworkType:        os.Getenv("SLS_RUNNER_ACINETWORKTYPE")}
	out, err := json.Marshal(msg)
	if err != nil {
		logrus.Errorln("convert struct to byte fail,", err)
	}
	error := insertPool(string(out), os.Getenv("SLS_MNS_URL"), os.Getenv("SLS_ALI_KEY"), os.Getenv("SLS_ALI_SEC"))
	if err != nil {
		logrus.Errorln("insert msg fail,", error)
	}
}
