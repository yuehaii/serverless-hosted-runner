// Package dispatcher of micro service dispatcher
package dispatcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	agent "serverless-hosted-runner/agent"
	common "serverless-hosted-runner/common"
	listener "serverless-hosted-runner/network/grpc"
	tfc "serverless-hosted-runner/tfcontroller"
	"slices"
	"strconv"
	"strings"

	ali_mns "github.com/aliyun/aliyun-mns-go-sdk"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type Dispatcher interface {
	HandleEvents(w http.ResponseWriter, req *http.Request)
	Refresh()
	Init()
}

type EciDispatcher struct {
	curPath            string
	poolPrefix         string
	interval           int64
	imageVer           string
	lazyRegs           string
	defaultLabels      []string
	allenRegs          string
	sysCtl             common.IUnixSysCtl
	cryptCtl           common.Cryptography
	customizedResolver bool
	checkConnectivity  bool
	disAddr            string
	rCompetition       bool
	evenProxyEnable    bool
	gitAgent           agent.IGit
	ctxLogLevel        string
	poolMode           bool
	cloudPr            string
	eventPush          bool
	tfCtl              string
}

func EciDispatcherConstruct(imageVer string, lazyRegs string, allenRegs string,
	ctxLogLevel string, poolMode bool, cloudPr string, eventPush bool, tfCtl string) Dispatcher {
	return &EciDispatcher{"/go/bin/", agent.NotificationQueue, int64(10), imageVer,
		lazyRegs, []string{"serverless-hosted-runner", "eci-runner"}, allenRegs,
		common.CreateUnixSysCtl(), common.DefaultCryptography(""), false, false, "",
		true, true, agent.CreateGitAgent(), ctxLogLevel, poolMode, cloudPr, eventPush, tfCtl}
}

func FnDispatcherConstruct() Dispatcher {
	return nil
}

func (dis EciDispatcher) responseBack(w http.ResponseWriter, msg string, logstr string, status int) {
	logrus.Info(logstr)
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "text/plain")
	if _, err := w.Write([]byte(msg)); err != nil {
		logrus.Errorf("fail to write response back, %v", err)
	}
}

func (dis EciDispatcher) parseRegistration(item common.PoolMsg) {
	logrus.Infof("parseRegistration, Repos: %s", item.Repos)
	repos := strings.Split(item.Repos, ",")
	url := item.URL
	for _, r := range repos {
		if len(r) > 0 {
			item.Type = "Repo"
			item.Name = r
			item.URL = url + "/" + r
			logrus.Infof("parseRegistration, item.Name: %s", item.Name)
			store := common.EnvStore(&item, item.Name, r)
			store.Save()
			key, runnerType := store.GetKey()
			pat, patType := store.GetPat()
			labels, labelType := store.GetLabels()
			logrus.Infof("parseRegistration, key: %s, pat: %s, labels: %s, runnerType %s, patType %s, labelType %s",
				key, pat, labels, runnerType, patType, labelType)

			iv, _ := strconv.Atoi(item.PullInterval)
			wf := agent.CreateWorkflowAgent(item.Type, item.Name, item.URL, dis.createRunner,
				dis.removeRunner, dis.notifyRelease, dis.checkLabels, r, item.Name, iv, labels)
			wf.InitAgent()
			wf.MonitorOnAgent()
			wf = nil
		}
	}
}

func (dis EciDispatcher) lazyRegistration() {
	logrus.Infof("lazyRegistration start. lazyRegs: %s", dis.lazyRegs)
	arrLazyRegs := []common.PoolMsg{}
	_ = json.Unmarshal([]byte(dis.lazyRegs), &arrLazyRegs)
	if len(arrLazyRegs) > 0 {
		for _, item := range arrLazyRegs {
			dis.parseRegistration(item)
		}
	}
	if dis.evenProxyEnable {
		eAgent := agent.CreateKafkaEventAgent(dis.createRunner, dis.removeRunner,
			os.Getenv("KAFKA_INS_USERNAME"), os.Getenv("KAFKA_INS_PWD"), os.Getenv("KAFKA_INS_ENDPOINT"),
			os.Getenv("KAFKA_INS_TOPIC"), os.Getenv("KAFKA_INS_CONSUMER"),
			arrLazyRegs)
		eAgent.MonitorOnAgent()
	}
}

func (dis EciDispatcher) allenRegistration() {
	logrus.Infof("allen registration start")
	aln := agent.CreateAllenStoreAgent(dis.parseRegistration)
	aln.InitAgent()
	aln.MonitorOnAgent()
}

func (dis *EciDispatcher) Init() {
	dis.disAddr = dis.sysCtl.Addr()
	if dis.customizedResolver {
		go dis.sysCtl.SetResolvers()
	}
	if dis.checkConnectivity {
		go dis.sysCtl.NetworkConnectivity()
	}
	lis := listener.CreateListener(dis.removeRunner, dis.disAddr)
	go lis.Start()
}

func (dis EciDispatcher) Refresh() {
	logrus.Infof("refresh pool start")
	common.SetContextLogLevel(dis.ctxLogLevel)

	if dis.lazyRegs != "" && dis.lazyRegs != "none" {
		go dis.lazyRegistration()
	}
	if dis.allenRegs == "allen" {
		go dis.allenRegistration()
	}

	if dis.poolMode {
		qAgent := agent.CreateAliMNSAgent(os.Getenv("TF_VAR_MNS_URL"), os.Getenv("ALICLOUD_ACCESS_KEY"),
			os.Getenv("ALICLOUD_SECRET_KEY"), agent.DefaultPoolQueue, dis.checkMsg, nil)
		qAgent.MonitorOnAgent()
	}
}

func (dis EciDispatcher) updatePool(msg common.PoolMsg, store common.Store) {
	logrus.Infof("update pool begin...")
	num, err := strconv.Atoi(store.GetPreSize())
	if err != nil {
		logrus.Errorf("updatePool strconv failure: %s", err)
		num = 0
	}
	logrus.Infof("updatePool Size %s, Type %s, Name %s, Pat %s, AnyChange %t", msg.Size,
		msg.Type, msg.Name, msg.Pat, store.AnyChange())
	if msg.Size == "0" && msg.Type == "Pool" {
		logrus.Infof("release pool, name: %s, num: %d", msg.Name, num)
		go dis.releasePool(msg.Name, num)
	} else if msg.Type == "Pool" && msg.Name != "" && msg.Pat != "null" && store.AnyChange() {
		logrus.Infof("release/recreate pool, name: %s, num: %d", msg.Name, num)
		dis.releasePool(msg.Name, num) // upinsert
		out, err := exec.Command("/bin/bash", dis.curPath+"create_runner.sh", "create_pool",
			dis.poolPrefix+"-"+msg.Name, msg.Size, msg.Name, msg.URL, msg.Pat, dis.imageVer,
			msg.Key, msg.Secret, msg.Region, msg.SecGpID, msg.VSwitchID, "pool", msg.CPU,
			msg.Memory, common.Ternary(msg.Labels == "", "none", msg.Labels).(string),
			common.Ternary(msg.ChargeLabels == "", "none", msg.ChargeLabels).(string),
			common.Ternary(msg.RunnerGroup == "", "default", msg.RunnerGroup).(string),
			dis.ctxLogLevel, dis.cloudPr,
			msg.ArmClientID, msg.ArmClientSecret, msg.ArmSubscriptionID, msg.ArmTenantID,
			msg.ArmEnvironment, msg.ArmRPRegistration, msg.ArmResourceGroupName, msg.ArmSubnetID,
			msg.ArmLogAnaWorkspaceID, msg.ArmLogAnaWorkspaceKey,
			msg.GcpCredential, msg.GcpProject, msg.GcpRegion, msg.GcpSA, msg.GcpApikey, msg.GcpDind,
			msg.GcpVpc, msg.GcpSubnet).Output()
		if err != nil {
			logrus.Errorf("error %s", err)
		}
		fmt.Printf("pool creation %s", out)
	}
}

func (dis EciDispatcher) notifyRelease(msg string) {
	if dis.eventPush {
		qAgent := agent.CreateAliMNSAgent(os.Getenv("TF_VAR_MNS_URL"), os.Getenv("ALICLOUD_ACCESS_KEY"),
			os.Getenv("ALICLOUD_SECRET_KEY"), agent.NotificationQueue, nil, nil)
		qAgent.NotifyAgent(msg)
		logrus.Infof("notifyRelease msg: %s", msg)
	}
}

func (dis EciDispatcher) releasePool(orgName string, num int) {
	pName := dis.poolPrefix + "-" + orgName
	for id := 1; id <= num; id++ {
		dis.notifyRelease(pName + "-" + strconv.Itoa(id))
	}
	fmt.Printf("release pool. output - %s\n",
		dis.removeRunner("pool_completed", pName, "", orgName, pName, []string{}, "", ""))
}

func (dis EciDispatcher) msgInvisible(t string) int64 {
	return common.Ternary(t == "Pool", int64(60), int64(5)).(int64)
}

func (dis EciDispatcher) checkMsg(obj interface{}, para interface{}) bool {
	q, _ := obj.(ali_mns.AliMNSQueue)
	endChan := make(chan bool)
	respChan := make(chan ali_mns.MessageReceiveResponse)
	errChan := make(chan error)
	go func() {
		select {
		case resp := <-respChan:
			{
				msg := common.PoolMsg{}
				if err := json.Unmarshal([]byte(resp.MessageBody), &msg); err != nil {
					logrus.Errorln(err)
				}
				fmt.Println("Unmarshal data, ", msg.Type, msg.Name, msg.Pat, msg.URL, msg.Size, msg.Key,
					msg.Secret, msg.Region, msg.SecGpID, msg.VSwitchID,
					msg.CPU, msg.Memory)
				if ret, e := q.ChangeMessageVisibility(resp.ReceiptHandle, dis.msgInvisible(msg.Type)); e != nil {
					fmt.Println("visibility error", e)
				} else {
					fmt.Println("visibility changed", ret, "delete msg now:", ret.ReceiptHandle)
					store := common.EnvStore(&msg, msg.Name, msg.Name)
					store.Save()
					if e := q.DeleteMessage(ret.ReceiptHandle); e != nil {
						fmt.Println(e)
					}
					go dis.updatePool(msg, store)
					if msg.Type == "Repo" {
						iv, _ := strconv.Atoi(msg.PullInterval)
						wf := agent.CreateWorkflowAgent(msg.Type, msg.Name, msg.URL, dis.createRunner,
							dis.removeRunner, dis.notifyRelease, dis.checkLabels, "", "", iv, msg.Labels)
						wf.InitAgent()
						wf.MonitorOnAgent()
					}
					endChan <- false
				}

				endChan <- false
			}
		case err := <-errChan:
			{
				if err != nil && !ali_mns.ERR_MNS_MESSAGE_NOT_EXIST.IsEqual(err) {
					logrus.Errorln(err)
				}
				endChan <- false
			}
		}
	}()
	q.ReceiveMessage(respChan, errChan, dis.interval)
	return <-endChan
}

func (dis EciDispatcher) HandleEvents(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		dis.responseBack(w, "Event Parsing Error:", "Parsing event body error.", http.StatusBadRequest)
		return
	}
	githubEvent := req.Header["X-Github-Event"][0]
	if githubEvent != "workflow_job" && githubEvent != "ping" {
		dis.responseBack(w, "Unsupported Event", "Request is not workflow_job or ping event.", http.StatusBadRequest)
		return
	} else if githubEvent == "ping" {
		dis.responseBack(w, "Ping Finish", "The dispatcher running normally.", http.StatusOK)
		return
	}

	eventData := common.EventBody{}
	if err := json.Unmarshal([]byte(body), &eventData); err != nil {
		logrus.Errorf("Dispacher handle event, fail to unmarshal event: %v", err)
	}
	orgName := eventData.Organization.Login
	if orgName == "" {
		orgName = eventData.Sender.Login
	}
	repoName := eventData.Repository.Name
	if eventData.Action == "queued" {
		logrus.Infof("Checking queued condition...")
		fmt.Printf("create a new runner. output - %s\n",
			dis.createRunner(eventData.Action,
				strconv.FormatInt(eventData.WorkflowJob.RunID, 10)+"-"+strconv.FormatInt(eventData.WorkflowJob.ID, 10),
				repoName, eventData.Repository.HTMLURL, orgName,
				eventData.Repository.Owner.Login, eventData.WorkflowJob.Labels))
	} else if eventData.Action == "completed" && eventData.WorkflowJob.RunnerID > 0 {
		logrus.Infof("Checking completed condition...")
		dis.notifyRelease(eventData.WorkflowJob.RunnerName)
		if !strings.Contains(eventData.WorkflowJob.RunnerName, dis.poolPrefix) {
			fmt.Printf("remove a runner. output - %s\n",
				dis.removeRunner(eventData.Action, eventData.WorkflowJob.RunnerName, repoName, orgName,
					strconv.FormatInt(eventData.WorkflowJob.RunID, 10)+"-"+strconv.FormatInt(eventData.WorkflowJob.ID, 10),
					eventData.WorkflowJob.Labels, eventData.Repository.HTMLURL, eventData.Repository.Owner.Login))
		}
	} else {
		fmt.Printf("skip the action - %s\n", eventData.Action)
	}
	dis.responseBack(w, "Exist dispatcher.", "The dispatcher event handle finished.", http.StatusOK)
}

func (dis EciDispatcher) checkLabels(labels []string, repoName string,
	orgName string, runnerType string, specificLabels string) bool {
	for idx, item := range labels {
		logrus.Infof("#%d label: %s", idx, item)
		if strings.Contains(item, ",") {
			itemArr := strings.Split(strings.TrimSpace(item), ",")
			labels = append(labels, itemArr...)
		}
	}
	logrus.Infof("checkLabels. repoName: %s, orgName: %s, runnerType: %s, specificLabels: %s",
		repoName, orgName, runnerType, specificLabels)
	if (slices.Contains(labels, repoName) && strings.EqualFold(runnerType, "repo")) ||
		(slices.Contains(labels, orgName) && strings.EqualFold(runnerType, "org")) {
		logrus.Warnf("wf label contains repo/org name. permit.")
		return true
	}
	if len(specificLabels) > 0 {
		specificLabelsArr := strings.Split(specificLabels, ",")
		for _, val := range specificLabelsArr {
			if slices.Contains(labels, strings.TrimSpace(val)) {
				logrus.Warnf("wf label contains customized label. permit.")
				return true
			}
		}
	}
	for _, val := range dis.defaultLabels {
		logrus.Infof("#default label: %s", strings.TrimSpace(val))
		if slices.Contains(labels, strings.TrimSpace(val)) {
			logrus.Warnf("wf label contains default label. permit.")
			return true
		}
	}
	logrus.Infof("not supported wf labels")
	return false
}

// cpu-1.0, memory-2.0
func (dis EciDispatcher) checkDynamicLabels(labels []string, oldCPU *string,
	oldMemory *string, oldVsw *string, oldSg *string, oldImg *string, largeDisk *string) string {
	cpu, memory, vsw, sg, targetLabel, img, bucket := "", "", "", "", "", "", ""
	for _, label := range labels {
		if strings.Contains(label, "cpu-") {
			cpu = strings.ReplaceAll(label, "cpu-", "")
			targetLabel += "," + label
		} else if strings.Contains(label, "memory-") {
			memory = strings.ReplaceAll(label, "memory-", "")
			targetLabel += "," + label
		} else if strings.Contains(label, "vsw-") {
			vsw = label
			targetLabel += "," + label
		} else if strings.Contains(label, "sg-") {
			sg = label
			targetLabel += "," + label
		} else if strings.Contains(label, "img-") {
			img = strings.ReplaceAll(label, "img-", "")
			targetLabel += "," + label
		} else if strings.Contains(label, "sid-") {
			targetLabel += "," + label
		} else if strings.Contains(label, "disk-") {
			bucket = strings.ReplaceAll(label, "disk-", "")
			targetLabel += "," + label
		}
	}
	if len(cpu) > 0 && len(memory) > 0 {
		*oldCPU = cpu
		*oldMemory = memory
	}
	if len(vsw) > 4 && len(sg) > 3 {
		*oldVsw = vsw
		*oldSg = sg
	}
	if len(img) > 0 {
		*oldImg = img
	}
	if len(bucket) > 0 {
		*largeDisk = bucket
	}
	logrus.Infof("checkDynamicLabels. oldCPU:%s, oldMemory:%s, cpu:%s, memory:%s, oldVsw:%s, oldSg:%s, vsw:%s, sg:%s, img:%s, targetLabel:%s, bucket:%s",
		*oldCPU, *oldMemory, cpu, memory, *oldVsw, *oldSg, vsw, sg, img, targetLabel, bucket)
	return targetLabel
}

func (dis EciDispatcher) runnerInfoVerification(store common.Store, labels []string, repoName, orgName string) (bool, string) {
	key, runnerType := store.GetKey()
	pat, _ := store.GetPat()
	specificLabels, _ := store.GetLabels()

	if len(key) <= 0 || len(runnerType) <= 0 || pat == "null" {
		return false, "org and repo token not exist, skip creation. key:" + key + ",type:" + runnerType + ",pat:" + pat
	} else if runnerType == "pool" {
		return false, "pool should not be created here. skip it."
	} else if !dis.checkLabels(labels, repoName, orgName, runnerType, specificLabels) {
		return false, "wf lablel dose not specify the repo/org name or runner label:" + specificLabels
	}
	return true, "runner info verification pass."
}

func (dis EciDispatcher) createRunner(act string, runnerID string, repoName string,
	repoURL string, orgName string, ownerName string, labels []string) string {
	store := common.EnvStore(nil, orgName, repoName)
	if verified, inf := dis.runnerInfoVerification(store, labels, repoName, orgName); !verified {
		logrus.Warnf("fail to verify runner information %s", inf)
		return "fail to verify runner information. " + inf
	}
	if dis.tfCtl == "go" {
		itfCtl, msg, errTf := dis.createRunnerTfCtl(store, act, runnerID, repoName, repoURL, orgName, ownerName, labels)
		if errTf != nil && itfCtl != nil {
			tfErrMsg := errTf.Error()
			isFileBusy := dis.sysCtl.IsFileBusy(tfErrMsg)
			isSysBusy := dis.sysCtl.IsSysBusy(tfErrMsg)
			logrus.Warnf("tf gen error message is %s, isFileBusy %v, isSysBusy %v",
				tfErrMsg, isFileBusy, isSysBusy)
			if err := itfCtl.MarkAsFinish("gen", isSysBusy || isFileBusy); err != nil {
				logrus.Errorf("fail to mark gen as finish: %v, tf err:%v", err, errTf)
				return err.Error() + msg + tfErrMsg
			}
			if isSysBusy || isFileBusy {
				if isFileBusy {
					logrus.Warnf("file busy detected in tf msg %s", tfErrMsg)
					if err := itfCtl.CleanTfAndLock(); err != nil {
						logrus.Errorf("fail to CleanTfAndLock: %v", err)
					}
				}
				logrus.Warnf("sys, file busy. begin plugin reload checking. msg:%s, err:%v", msg, errTf)
				if err := dis.sysCtl.ReloadPlugin(); err != nil {
					logrus.Errorf("create runner, fail to ReloadPlugin: %v", err)
				}
			}
		}
		logrus.Warnf("runner creation with tf controller. msg:%s, err:%v", msg, errTf)
		return msg
	} else {
		// if need pool(not recommended), pls use cmd mode
		return dis.createRunnerCmd(store, act, runnerID, repoName, repoURL, orgName, ownerName, labels)
	}
}

func (dis EciDispatcher) createRunnerTfCtl(store common.Store, act string, runnerID string, repoName string,
	repoURL string, orgName string, ownerName string, labels []string) (tfc.ITfController, string, error) {
	logrus.Infof("Tf Controller Creating runner...")

	ctl := tfc.CreateController(dis.curPath,
		map[string]string{"act": act, "runer_id": runnerID, "repo_name": repoName,
			"repo_url": repoURL, "org_name": orgName, "owner_name": ownerName,
			"image_ver": dis.imageVer, "ctx_log_level": dis.ctxLogLevel, "dis_ip": dis.disAddr,
		}, store, dis.cloudPr, dis.checkDynamicLabels, labels, dis.gitAgent)

	logrus.Infof("Creating runner %s with tf ctl...", ctl.TfFilePath())
	if isfinish, err := ctl.Finished("gen"); !isfinish {
		if err != nil {
			return nil, err.Error(), err
		} else {
			msg := "can't gen lock" + ctl.TfFilePath()
			return nil, msg, fmt.Errorf("can't gen lock %s", ctl.TfFilePath())
		}
	}
	if err := ctl.GenTfConfigs(dis.curPath + "runner/"); err != nil {
		return ctl, "fail to generate tf configurations for" + ctl.TfFilePath(), err
	}
	if err := ctl.InitTerraform(); err != nil {
		return ctl, "fail to init tf controller for" + ctl.TfFilePath(), err
	}
	if state, des := ctl.State(false); state {
		msg := des + " state of runner path exists. skip it: " + ctl.TfFilePath()
		return nil, msg, nil
	}

	initAndApply := func() (tfc.ITfController, string, error) {
		if err := ctl.Init(); err != nil {
			return ctl, "fail to init tf configurations.", err
		}
		if err := ctl.Apply(); err != nil {
			return ctl, "fail to apply tf configurations.", err
		}
		return nil, "Success init and apply", nil
	}

	if interCtl, interMsg, interErr := initAndApply(); interErr != nil {
		logrus.Errorf("initAndApply err %s, %s", ctl.TfFilePath(), interErr.Error())
		if strings.Contains(interErr.Error(),
			"installed provider plugins are not consistent") {
			logrus.Warnf("hcl inconsist with pr of %s, clean and init apply again", ctl.TfFilePath())
			if err := interCtl.CleanHCL(); err != nil {
				logrus.Errorf("create runner, fail to clean hcl, %v", err)
			}
			if interCtlHcl, interMsgHcl, interErrHcl := initAndApply(); interErrHcl != nil {
				logrus.Errorf("hcl clean fail "+ctl.TfFilePath()+", %s", interErrHcl.Error())
				return interCtlHcl, interMsgHcl, interErrHcl
			}
		} else {
			return interCtl, interMsg, interErr
		}
	}

	if state, des := ctl.State(true); !state {
		msg := "fail to apply with nil err, but state of " + ctl.TfFilePath() + " does not contains service. " + des
		return ctl, msg, fmt.Errorf("fail to apply with nil err, but state of %s does not contains service %s", ctl.TfFilePath(), des)

	}

	if !dis.rCompetition {
		logrus.Warnf("runner competition not detected. safe reset destroy for %s", ctl.TfFilePath())
		store.ResetDestory(ctl.TfFilePath())
		if err := ctl.MarkAsFinish("del", false); err != nil {
			logrus.Errorf("fail to mark del as finish after creation of runner %s: %v", ctl.TfFilePath(), err)
		}
	}

	return nil, "Finish runner creation." + ctl.TfFilePath(), nil
}

func (dis EciDispatcher) createRunnerCmd(store common.Store, act string, runnerID string, repoName string,
	repoURL string, orgName string, ownerName string, labels []string) string {
	logrus.Infof("CMD Creating runner...")

	key, runnerType := store.GetKey()
	pat, patType := store.GetPat()
	specificLabels, _ := store.GetLabels()
	regURL, _ := store.GetURL()
	cpu, _ := store.GetCPU()
	mem, _ := store.GetMemory()
	secGpID, _ := store.GetSecGpID()
	vswitchID, _ := store.GetVSwitchID()
	gcpDind, _ := store.GetGcpDind()
	repoImageVer, _ := store.GetImageVersion()
	imageVer := common.Ternary(repoImageVer == "", dis.imageVer, repoImageVer).(string)
	largeDisk := ""

	if len(key) <= 0 || len(runnerType) <= 0 || pat == "null" {
		logrus.Warnf("Skip the runner creation. key: %s, type: %s, pat: %s", key, runnerType, pat)
		return "org and repo token not exist. please run the 'register serverless runner' workflow first."
	} else if runnerType == "pool" {
		return "pool should not be created here. skip it."
	} else if !dis.checkLabels(labels, repoName, orgName, runnerType, specificLabels) {
		return "wf lablel dose not specify the repo/org name or runner label:" + specificLabels
	}

	specificLabels += dis.checkDynamicLabels(labels, &cpu, &mem, &vswitchID, &secGpID, &imageVer, &largeDisk)

	logrus.Infof("Create runner. patType: %s, runnerType: %s", patType, runnerType)
	if patType != runnerType && (patType == "repo" && runnerType == "org") {
		runnerType = patType
		regURL = regURL + "/" + repoName
	}

	sec, _ := store.GetSecret()
	region, _ := store.GetRegion()
	chargeLabels, _ := store.GetChargeLabels()
	runnerGroup, _ := store.GetRunnerGroup()
	armClientID, _ := store.GetArmClientID()
	armClientSecret, _ := store.GetArmClientSecret()
	armSubscriptionID, _ := store.GetArmSubscriptionID()
	armTenantID, _ := store.GetArmTenantID()
	armEnvironment, _ := store.GetArmEnvironment()
	armRpRegistration, _ := store.GetArmRPRegistration()
	armResourceGroupName, _ := store.GetArmResourceGroupName()
	armSubnetID, _ := store.GetArmSubnetID()
	armLogAnaWorkspaceID, _ := store.GetArmLogAnalyticsWorkspaceID()
	armLogAnaWorkspaceKey, _ := store.GetArmLogAnalyticsWorkspaceKey()
	gcpCredentials, _ := store.GetGcpCredential()
	gcpProject, _ := store.GetGcpProject()
	gcpRegion, _ := store.GetGcpRegion()
	gcpSa, _ := store.GetGcpSA()
	gcpApikey, _ := store.GetGcpApikey()
	gcpVpc, _ := store.GetGcpVpc()
	gcpSubnet, _ := store.GetGcpSubnet()
	// following vars will not be sync into cmd and renewed by go clt
	aciLocation, _ := store.GetAciLocation()
	aciSku, _ := store.GetAciSku()
	aciNetworkType, _ := store.GetAciNetworkType()
	out, err := exec.Command("/bin/bash", dis.curPath+"create_runner.sh", act, runnerID,
		repoName, regURL, orgName, ownerName, pat, imageVer, key, sec, region,
		secGpID, vswitchID, runnerType, cpu, mem,
		common.Ternary(specificLabels == "", "none", strings.ReplaceAll(specificLabels, " ", "")).(string),
		common.Ternary(chargeLabels == "", "none", strings.ReplaceAll(chargeLabels, " ", "")).(string),
		common.Ternary(runnerGroup == "", "default", strings.ReplaceAll(runnerGroup, " ", "")).(string),
		dis.ctxLogLevel, dis.cloudPr,
		armClientID, armClientSecret, armSubscriptionID, armTenantID, armEnvironment, armRpRegistration,
		armResourceGroupName, armSubnetID, armLogAnaWorkspaceID, armLogAnaWorkspaceKey,
		gcpCredentials, gcpProject, gcpRegion, gcpSa, gcpApikey, gcpDind, gcpVpc, gcpSubnet,
		aciLocation, aciSku, aciNetworkType, largeDisk,
	).Output()
	if err != nil {
		logrus.Errorf("error %s", err)
	} else {
		logrus.Info("ResetDestory " + repoName + "-" + runnerID)
		store.ResetDestory(repoName + "-" + runnerID)
	}
	output := string(out)
	logrus.Warnf("Finish runner creation, output: %s", output)

	return output
}

func (dis EciDispatcher) removeRunner(act string, runnerName string, repoName string,
	orgName string, runWf string, labels []string, url string, owner string) string {
	logrus.Printf("removeRunner paras: act %s, runnerName %s, repoName %s, orgName %s, runWf %s, labels %v, url %s, owner %s",
		act, runnerName, repoName, orgName, runWf, labels, url, owner)
	store := common.EnvStore(nil, orgName, repoName)
	if store.IsDestory(runnerName + "-" + runWf) {
		logrus.Warn("workflow " + repoName + "-" + runWf + " occupied runner " + runnerName + " already removed.")
		return "workflow " + repoName + "-" + runWf + " occupied runner " + runnerName + " already removed."
	}
	if dis.tfCtl == "go" {
		itfCtl, msg, errTf := dis.removeRunnerTfCtl(store, act, runnerName, repoName,
			orgName, runWf, labels, url, runWf, owner)
		if itfCtl != nil {
			if errTf != nil {
				tfErrMsg := errTf.Error()
				isFileBusy := dis.sysCtl.IsFileBusy(tfErrMsg)
				isSysBusy := dis.sysCtl.IsSysBusy(tfErrMsg)
				logrus.Warnf("tf del error message is %s", tfErrMsg)
				if err := itfCtl.MarkAsFinish("del", isSysBusy || isFileBusy); err != nil {
					logrus.Warn("after desrtoy, fail to mark del as finish" + err.Error() + errTf.Error() + msg)
					return err.Error() + errTf.Error() + msg
				}
				if isFileBusy || isSysBusy {
					logrus.Warnf("sys,file busy during destroy. plugin reload checking. msg:%s, err:%v", msg, errTf)
					if err := dis.sysCtl.ReloadPlugin(); err != nil {
						logrus.Errorf("remove runner, fail to ReloadPlugin: %v", err)
					}
				}
			}
			if err := itfCtl.MarkAsFinish("gen", false); err != nil {
				logrus.Warn("after desrtoy, fail to mark gen as finish" + err.Error() + errTf.Error() + msg)
				return err.Error() + errTf.Error() + msg
			}
		}
		logrus.Warnf("runner removing with tf controller. msg:%s, err:%s", msg, errTf)
		return msg
	} else {
		return dis.removeRunnerCmd(store, act, runnerName, repoName, orgName, runWf)
	}
}

func (dis EciDispatcher) removeRunnerTfCtl(store common.Store, act string, runnerName string, repoName string,
	orgName string, runWf string, labels []string, url string, runnerID string,
	owner string) (tfc.ITfController, string, error) {
	runOn := common.Ternary(len(runnerName) == 0, repoName+"-"+runWf, runnerName).(string)
	logrus.Info("Removing runner " + dis.curPath + runOn + " with tf ctl...")
	ctl := tfc.DestroyController(runOn, dis.curPath+runOn, store, dis.cloudPr, dis.curPath, dis.gitAgent)
	if ctl == nil {
		return nil, "org and repo token dose not exist, please register.", nil
	}
	if isfinish, err := ctl.Finished("del"); !isfinish {
		if err != nil {
			return nil, err.Error(), err
		} else {
			msg := "can't del lock" + repoName + "-" + runWf
			return nil, msg, fmt.Errorf("can't del lock %s-%s", repoName, runWf)
		}
	}
	if !ctl.TfConfigsExists() {

		msg := "Runner config dose not exists. Skip destroy."
		return nil, msg, fmt.Errorf("runner config dose not exists. skip destroy")
	}
	if err := ctl.InitTerraform(); err != nil {
		msg := "fail to init Tf controller for destroy."
		return ctl, msg, fmt.Errorf("fail to init Tf controller for destroy")
	}
	if state, des := ctl.State(false); !state {
		msg := des + " runner path: " + common.Ternary(len(runnerName) == 0, repoName+"-"+runWf, runnerName).(string)

		return ctl, msg, fmt.Errorf("%s runner path: %s", des,
			common.Ternary(len(runnerName) == 0, repoName+"-"+runWf, runnerName).(string))
	}
	if dis.sysCtl.ExceedReload() && !strings.Contains(runnerName, runWf) && !strings.Contains(runnerName, "sls-comp") {
		if exist, _ := ctl.FileState(dis.curPath + runWf); !exist {
			logrus.Warnf("runner competition detected run on %s, wf %s", runnerName, runWf)
			itfCtl, msg, errTf := dis.createRunnerTfCtl(store, act, runWf+"-sls-comp-"+dis.cryptCtl.RandStr(4), repoName, url, orgName, owner, labels)
			if errTf != nil && itfCtl != nil {
				tfErrMsg := errTf.Error()
				logrus.Warnf("runner competition, tf gen error message is %s, msg %s", tfErrMsg, msg)
				if err := itfCtl.MarkAsFinish("gen", dis.sysCtl.IsSysBusy(tfErrMsg)); err != nil {
					logrus.Errorf("runner competition, fail to mark gen as finish: %s, tf err:%s", err, errTf)
				}
				return ctl, "fail to create competition runner", errTf
			}
		}
	}
	if err := ctl.Destroy(); err != nil {
		return ctl, "fail to Destroy runner. " + err.Error(), err
	}
	store.MarkDestory(runOn)
	return ctl, "Success remove runner " + runOn, nil
}

func (dis EciDispatcher) removeRunnerCmd(store common.Store, act string, runnerName string, repoName string,
	orgName string, runWf string) string {
	logrus.Infof("CMD Remove runner...")
	key, _ := store.GetKey()
	sec, _ := store.GetSecret()
	region, _ := store.GetRegion()
	armClientID, _ := store.GetArmClientID()
	armClientSecret, _ := store.GetArmClientSecret()
	armSubscriptionID, _ := store.GetArmSubscriptionID()
	armTenantID, _ := store.GetArmTenantID()
	armEnvironment, _ := store.GetArmEnvironment()
	armRpRegistration, _ := store.GetArmRPRegistration()
	armSubnetID, _ := store.GetArmSubnetID()
	armLogAnaWorkspaceID, _ := store.GetArmLogAnalyticsWorkspaceID()
	armLogAnaWorkspaceKey, _ := store.GetArmLogAnalyticsWorkspaceKey()
	if len(key) <= 0 {
		return "org and repo token not exist. please run the 'register serverless runner' workflow first."
	} else if store.IsDestory(repoName + "-" + runWf) {
		logrus.Info("workflow " + repoName + "-" + runWf + " occupied runner " + runnerName + " already removed.")
		return "workflow " + repoName + "-" + runWf + " occupied runner " + runnerName + " already removed."
	}
	out, err := exec.Command("/bin/bash", dis.curPath+"remove_runner.sh", act, runnerName, key, sec,
		region, orgName, repoName, runWf, dis.cloudPr,
		armClientID, armClientSecret, armSubscriptionID, armTenantID, armEnvironment,
		armRpRegistration, armSubnetID, armLogAnaWorkspaceID, armLogAnaWorkspaceKey).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	} else {
		logrus.Info("MarkDestory " + repoName + "-" + runWf)
		store.MarkDestory(repoName + "-" + runWf)
	}
	output := string(out)
	logrus.Warnf("Remove runner finish, output: %s", output)
	return output
}
