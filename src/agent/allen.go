package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	cloud "serverless-hosted-runner/cloud"
	common "serverless-hosted-runner/common"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type IAllenStore interface {
	Agent
}

type AllenStore struct {
	pull_interval    string
	db_scan_interval int
	scan_marker      string
	db_handler       common.IPostgresDB
	parse_reg        common.ParseRegistration
	vs_detector      cloud.IVSWitch
}

func CreateAllenStoreAgent(parser common.ParseRegistration) IAllenStore {
	return &AllenStore{"2", 5, "allen_db_scaned_configs", nil, parser, nil}
}

func (aln *AllenStore) InitAgent() {
	aln.db_handler = common.CreatePostgresDB()
	aln.db_handler.InitConnection()
	aln.vs_detector = cloud.CreateVSWitch()
}

func (aln AllenStore) MonitorOnAgent() {
	mem := common.CreateMemDector()
	for {
		logrus.Infof("MonitorOnAgent, connect allen configuration...")
		mem.DetectUsage()
		if aln.db_handler.Connect() == nil {
			aln.listRegistration()
			aln.parseRegistrations()
			aln.db_handler.Disconnect()
		}
		logrus.Infof("MonitorOnAgent, wait " + strconv.Itoa(aln.db_scan_interval) + " minutes and check again")
		time.Sleep(time.Duration(aln.db_scan_interval) * time.Minute)
		mem.CompareLast()
	}
}

func (aln AllenStore) NotifyAgent(msg string) {
	// aln.db_handler.UpdateStatus("configured", "registration succeed")
}

func (aln AllenStore) listRegistration() {
	_, err := aln.db_handler.ListRow(aln.markRegistration(""))
	if err != nil {
		logrus.Errorf("listRegistration failure: %s", err)
	}
}

func (aln AllenStore) saveConfig(conf string) (common.PoolMsg, error) {
	msg := common.AllenMsg{}
	err := json.Unmarshal([]byte(conf), &msg)
	if err != nil {
		return common.PoolMsg{}, err
	}
	fmt.Println("allen, saveConfig Unmarshal data, ", msg.Type, msg.Name, msg.Pat, msg.Url, msg.Size, msg.Key,
		msg.Secret, msg.Region, msg.SecGpId, msg.VSwitchId,
		msg.Cpu, msg.Memory)
	conv_msg := msg.ConvertPoolMsg()
	conv_msg.PullInterval = aln.pull_interval
	store := common.EnvStore(&conv_msg, msg.Name, msg.Name)
	store.Save()
	return conv_msg, nil
}

func (aln AllenStore) markRegistration(reg_id string) string {
	cur_marker := os.Getenv(aln.scan_marker)
	if reg_id == "" {
		return cur_marker
	}
	if cur_marker != "" {
		os.Setenv(aln.scan_marker, cur_marker+","+reg_id)
		return cur_marker + "," + reg_id
	} else {
		os.Setenv(aln.scan_marker, reg_id)
		return reg_id
	}
}

func (aln AllenStore) parseRegistrations() {
	has_next := true
	err := error(nil)
	for has_next {
		var cur_idx string
		var msg_val string
		var status_val string
		var comment_val string

		has_next, err = aln.db_handler.IterateRows(&cur_idx, &msg_val, &status_val, &comment_val)
		if err != nil {
			aln.db_handler.UpdateStatus(cur_idx, "failed", fmt.Sprintf("%s", err))
			continue
		}
		if msg_val == "" {
			continue
		}
		conv_str, er := aln.saveConfig(msg_val)
		if er != nil {
			logrus.Errorf("saveConfig failure: %s", err)
		}
		aln.parse_reg(conv_str)
		aln.markRegistration(cur_idx)
		go aln.updateAllenStore(cur_idx, conv_str)
	}

}

func (aln AllenStore) updateAllenStore(cur_idx string, conv_str common.PoolMsg) {
	aln.vs_detector.Info(conv_str.VSwitchId, conv_str.Key, conv_str.Secret, conv_str.Region)
	aln.db_handler.UpdateStatus(cur_idx, "configured",
		fmt.Sprintf("registration succeed. vswitch %s, %d ip left",
			conv_str.VSwitchId, aln.vs_detector.LeftIPs()))
}
