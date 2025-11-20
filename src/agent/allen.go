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
	pullInterval   string
	dbScanInterval int
	scanMarker     string
	dbHandler      common.IPostgresDB
	parseReg       common.ParseRegistration
	vsDetector     cloud.IVSWitch
}

func CreateAllenStoreAgent(parser common.ParseRegistration) IAllenStore {
	return &AllenStore{"2", 5, "allen_db_scaned_configs", nil, parser, nil}
}

func (aln *AllenStore) InitAgent() {
	aln.dbHandler = common.CreatePostgresDB()
	aln.dbHandler.InitConnection()
	aln.vsDetector = cloud.CreateVSWitch()
}

func (aln AllenStore) MonitorOnAgent() {
	mem := common.CreateMemDector()
	for {
		logrus.Infof("MonitorOnAgent, connect allen configuration...")
		mem.DetectUsage()
		if aln.dbHandler.Connect() == nil {
			aln.listRegistration()
			aln.parseRegistrations()
			aln.dbHandler.Disconnect()
		}
		logrus.Infof("MonitorOnAgent, wait %s minutes and check again", strconv.Itoa(aln.dbScanInterval))
		time.Sleep(time.Duration(aln.dbScanInterval) * time.Minute)
		mem.CompareLast()
	}
}

func (aln AllenStore) NotifyAgent(msg string) {
	// aln.dbHandler.UpdateStatus("configured", "registration succeed")
}

func (aln AllenStore) listRegistration() {
	_, err := aln.dbHandler.ListRow(aln.markRegistration(""))
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
	fmt.Println("allen, saveConfig Unmarshal data, ", msg.Type, msg.Name, msg.Pat, msg.URL, msg.Size, msg.Key,
		msg.Secret, msg.Region, msg.SecGpID, msg.VSwitchID,
		msg.CPU, msg.Memory)
	convMsg := msg.ConvertPoolMsg()
	convMsg.PullInterval = aln.pullInterval
	store := common.EnvStore(&convMsg, msg.Name, msg.Name)
	store.Save()
	return convMsg, nil
}

func (aln AllenStore) markRegistration(regID string) string {
	curMarker := os.Getenv(aln.scanMarker)
	if regID == "" {
		return curMarker
	}
	if curMarker != "" {
		if err := os.Setenv(aln.scanMarker, curMarker+","+regID); err != nil {
			logrus.Errorf("markRegistration, fail to set env for marker: %s", err)
		}
		return curMarker + "," + regID
	} else {
		if err := os.Setenv(aln.scanMarker, regID); err != nil {
			logrus.Errorf("markRegistration, fail to set env: %s", err)
		}
		return regID
	}
}

func (aln AllenStore) parseRegistrations() {
	hasNext := true
	err := error(nil)
	for hasNext {
		var curIdx string
		var msgVal string
		var statusVal string
		var commentVal string

		hasNext, err = aln.dbHandler.IterateRows(&curIdx, &msgVal, &statusVal, &commentVal)
		if err != nil {
			if err = aln.dbHandler.UpdateStatus(curIdx, "failed", fmt.Sprintf("%s", err)); err != nil {
				logrus.Errorf("parseRegistrations, allen db UpdateStatus failure: %s", err)
			}
			continue
		}
		if msgVal == "" {
			continue
		}
		convStr, er := aln.saveConfig(msgVal)
		if er != nil {
			logrus.Errorf("saveConfig failure: %s", err)
		}
		aln.parseReg(convStr)
		aln.markRegistration(curIdx)
		go aln.updateAllenStore(curIdx, convStr)
	}

}

func (aln AllenStore) updateAllenStore(curIdx string, convStr common.PoolMsg) {
	aln.vsDetector.Info(convStr.VSwitchID, convStr.Key, convStr.Secret, convStr.Region)
	if err := aln.dbHandler.UpdateStatus(curIdx, "configured",
		fmt.Sprintf("registration succeed. vswitch %s, %d ip left",
			convStr.VSwitchID, aln.vsDetector.LeftIPs())); err != nil {
		logrus.Errorf("updateAllenStore, allen db UpdateStatus failure: %s", err)
	}
}
