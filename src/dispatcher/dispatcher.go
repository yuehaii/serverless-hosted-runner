package main

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
	cur_path            string
	pool_prefix         string
	interval            int64
	image_ver           string
	lazy_regs           string
	default_labels      []string
	allen_regs          string
	sys_ctl             common.IUnixSysCtl
	crypt_ctl           common.Cryptography
	customized_resolver bool
	check_connectivity  bool
	dis_addr            string
	r_competition       bool
}

func EciDispatcherConstruct(image_ver string, lazy_regs string, allen_regs string) Dispatcher {
	return &EciDispatcher{"/go/bin/", agent.NOTIFICATION_Q, int64(10), image_ver,
		lazy_regs, []string{"serverless-hosted-runner", "eci-runner"}, allen_regs,
		common.CreateUnixSysCtl(), common.DefaultCryptography(""), false, false, "", true}
}

func FnDispatcherConstruct() Dispatcher {
	return nil
}

func (dis EciDispatcher) responseBack(w http.ResponseWriter, msg string, logstr string, status int) {
	logrus.Infof(logstr)
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(msg))
}

func (dis EciDispatcher) parseRegistration(item common.PoolMsg) {
	logrus.Infof("parseRegistration, Repos: %s", item.Repos)
	repos := strings.Split(item.Repos, ",")
	url := item.Url
	for _, r := range repos {
		if len(r) > 0 {
			item.Type = "Repo"
			item.Name = r
			item.Url = url + "/" + r
			logrus.Infof("parseRegistration, item.Name: %s", item.Name)

			store := common.EnvStore(&item, item.Name, r)
			store.Save()
			key, runner_type := store.GetKey()
			pat, pat_type := store.GetPat()
			labels, label_type := store.GetLabels()
			logrus.Infof("parseRegistration, key: %s, pat: %s, labels: %s, runner_type %s, pat_type %s, label_type %s",
				key, pat, labels, runner_type, pat_type, label_type)

			iv, _ := strconv.Atoi(item.PullInterval)
			wf := agent.CreateWorkflowAgent(item.Type, item.Name, item.Url, dis.createRunner,
				dis.removeRunner, dis.notifyRelease, dis.checkLabels, r, item.Name, iv, labels)
			wf.InitAgent()
			wf.MonitorOnAgent()
			wf = nil
		}
	}
}

func (dis EciDispatcher) lazyRegistration() {
	logrus.Infof("lazyRegistration start. lazy_regs: %s", dis.lazy_regs)
	arr_lazy_regs := []common.PoolMsg{}
	_ = json.Unmarshal([]byte(dis.lazy_regs), &arr_lazy_regs)
	if len(arr_lazy_regs) > 0 {
		for _, item := range arr_lazy_regs {
			dis.parseRegistration(item)
		}
	}
}

func (dis EciDispatcher) allenRegistration() {
	logrus.Infof("allen registration start")
	aln := agent.CreateAllenStoreAgent(dis.parseRegistration)
	aln.InitAgent()
	aln.MonitorOnAgent()
}

func (dis *EciDispatcher) Init() {
	dis.dis_addr = dis.sys_ctl.Addr()
	if dis.customized_resolver {
		go dis.sys_ctl.SetResolvers()
	}
	if dis.check_connectivity {
		go dis.sys_ctl.NetworkConnectivity()
	}
	lis := listener.CreateListener(dis.removeRunner, dis.dis_addr)
	go lis.Start()
}

func (dis EciDispatcher) Refresh() {
	logrus.Infof("refresh pool start")
	common.SetContextLogLevel(*ctx_log_level)

	if dis.lazy_regs != "" && dis.lazy_regs != "none" {
		go dis.lazyRegistration()
	}
	if dis.allen_regs == "allen" {
		go dis.allenRegistration()
	}

	if *pool_mode {
		qAgent := agent.CreateAliMNSAgent(os.Getenv("TF_VAR_MNS_URL"), os.Getenv("ALICLOUD_ACCESS_KEY"),
			os.Getenv("ALICLOUD_SECRET_KEY"), agent.DEFAULT_POOL_Q, dis.checkMsg, nil)
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
		out, err := exec.Command("/bin/bash", dis.cur_path+"create_runner.sh", "create_pool",
			dis.pool_prefix+"-"+msg.Name, msg.Size, msg.Name, msg.Url, msg.Pat, dis.image_ver,
			msg.Key, msg.Secret, msg.Region, msg.SecGpId, msg.VSwitchId, "pool", msg.Cpu,
			msg.Memory, common.Ternary(msg.Labels == "", "none", msg.Labels).(string),
			common.Ternary(msg.ChargeLabels == "", "none", msg.ChargeLabels).(string),
			common.Ternary(msg.RunnerGroup == "", "default", msg.RunnerGroup).(string),
			*ctx_log_level, *cloud_pr,
			msg.ArmClientId, msg.ArmClientSecret, msg.ArmSubscriptionId, msg.ArmTenantId,
			msg.ArmEnvironment, msg.ArmRPRegistration, msg.ArmResourceGroupName, msg.ArmSubnetId,
			msg.ArmLogAnaWorkspaceId, msg.ArmLogAnaWorkspaceKey,
			msg.GcpCredential, msg.GcpProject, msg.GcpRegion, msg.GcpSA, msg.GcpApikey, msg.GcpDind,
			msg.GcpVpc, msg.GcpSubnet).Output()
		if err != nil {
			logrus.Errorf("error %s", err)
		}
		fmt.Printf("pool creation %s", out)
	}
}

func (dis EciDispatcher) notifyRelease(msg string) {
	if *event_push {
		qAgent := agent.CreateAliMNSAgent(os.Getenv("TF_VAR_MNS_URL"), os.Getenv("ALICLOUD_ACCESS_KEY"),
			os.Getenv("ALICLOUD_SECRET_KEY"), agent.NOTIFICATION_Q, nil, nil)
		qAgent.NotifyAgent(msg)
		logrus.Infof("notifyRelease msg: %s", msg)
	}
}

func (dis EciDispatcher) releasePool(org_name string, num int) {
	p_name := dis.pool_prefix + "-" + org_name
	for id := 1; id <= num; id++ {
		dis.notifyRelease(p_name + "-" + strconv.Itoa(id))
	}
	fmt.Printf("release pool. output - %s\n",
		dis.removeRunner("pool_completed", p_name, "", org_name, p_name, []string{}, "", ""))
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
				json.Unmarshal([]byte(resp.MessageBody), &msg)
				fmt.Println("Unmarshal data, ", msg.Type, msg.Name, msg.Pat, msg.Url, msg.Size, msg.Key,
					msg.Secret, msg.Region, msg.SecGpId, msg.VSwitchId,
					msg.Cpu, msg.Memory)
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
						wf := agent.CreateWorkflowAgent(msg.Type, msg.Name, msg.Url, dis.createRunner,
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
	github_event := req.Header["X-Github-Event"][0]
	if github_event != "workflow_job" && github_event != "ping" {
		dis.responseBack(w, "Unsupported Event", "Request is not workflow_job or ping event.", http.StatusBadRequest)
		return
	} else if github_event == "ping" {
		dis.responseBack(w, "Ping Finish", "The dispatcher running normally.", http.StatusOK)
		return
	}

	event_data := EventBody{}
	json.Unmarshal([]byte(body), &event_data)
	org_name := event_data.Organization.Login
	if org_name == "" {
		org_name = event_data.Sender.Login
	}
	repo_name := event_data.Repository.Name
	if event_data.Action == "queued" {
		logrus.Infof("Checking queued condition...")
		fmt.Printf("create a new runner. output - %s\n",
			dis.createRunner(event_data.Action,
				strconv.FormatInt(event_data.WorkflowJob.RunID, 10)+"-"+strconv.FormatInt(event_data.WorkflowJob.ID, 10),
				repo_name, event_data.Repository.HTMLURL, org_name,
				event_data.Repository.Owner.Login, event_data.WorkflowJob.Labels))
	} else if event_data.Action == "completed" && event_data.WorkflowJob.RunnerID > 0 {
		logrus.Infof("Checking completed condition...")
		dis.notifyRelease(event_data.WorkflowJob.RunnerName)
		if !strings.Contains(event_data.WorkflowJob.RunnerName, dis.pool_prefix) {
			fmt.Printf("remove a runner. output - %s\n",
				dis.removeRunner(event_data.Action, event_data.WorkflowJob.RunnerName, repo_name, org_name,
					strconv.FormatInt(event_data.WorkflowJob.RunID, 10)+"-"+strconv.FormatInt(event_data.WorkflowJob.ID, 10),
					event_data.WorkflowJob.Labels, event_data.Repository.HTMLURL, event_data.Repository.Owner.Login))
		}
	} else {
		fmt.Printf("skip the action - %s\n", event_data.Action)
	}
	dis.responseBack(w, "Exist dispatcher.", "The dispatcher event handle finished.", http.StatusOK)
}

func (dis EciDispatcher) checkLabels(labels []string, repo_name string,
	org_name string, runner_type string, specific_labels string) bool {
	for idx, item := range labels {
		logrus.Infof("#%d label: %s", idx, item)
		if strings.Contains(item, ",") {
			item_arr := strings.Split(strings.TrimSpace(item), ",")
			labels = append(labels, item_arr...)
		}
	}
	logrus.Infof("checkLabels. repo_name: %s, org_name: %s, runner_type: %s, specific_labels: %s",
		repo_name, org_name, runner_type, specific_labels)
	if (slices.Contains(labels, repo_name) && strings.EqualFold(runner_type, "repo")) ||
		(slices.Contains(labels, org_name) && strings.EqualFold(runner_type, "org")) {
		logrus.Warnf("wf label contains repo/org name. permit.")
		return true
	}
	if len(specific_labels) > 0 {
		specific_labels_arr := strings.Split(specific_labels, ",")
		for _, val := range specific_labels_arr {
			if slices.Contains(labels, strings.TrimSpace(val)) {
				logrus.Warnf("wf label contains customized label. permit.")
				return true
			}
		}
	}
	for _, val := range dis.default_labels {
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
func (dis EciDispatcher) checkDynamicLabels(labels []string, old_cpu *string,
	old_memory *string, old_vsw *string, old_sg *string, old_img *string, large_disk *string) string {
	cpu, memory, vsw, sg, target_label, img, bucket := "", "", "", "", "", "", ""
	for _, label := range labels {
		if strings.Contains(label, "cpu-") {
			cpu = strings.ReplaceAll(label, "cpu-", "")
			target_label += "," + label
		} else if strings.Contains(label, "memory-") {
			memory = strings.ReplaceAll(label, "memory-", "")
			target_label += "," + label
		} else if strings.Contains(label, "vsw-") {
			vsw = label
			target_label += "," + label
		} else if strings.Contains(label, "sg-") {
			sg = label
			target_label += "," + label
		} else if strings.Contains(label, "img-") {
			img = strings.ReplaceAll(label, "img-", "")
			target_label += "," + label
		} else if strings.Contains(label, "sid-") {
			target_label += "," + label
		} else if strings.Contains(label, "disk-") {
			bucket = strings.ReplaceAll(label, "disk-", "")
			target_label += "," + label
		}
	}
	if len(cpu) > 0 && len(memory) > 0 {
		*old_cpu = cpu
		*old_memory = memory
	}
	if len(vsw) > 4 && len(sg) > 3 {
		*old_vsw = vsw
		*old_sg = sg
	}
	if len(img) > 0 {
		*old_img = img
	}
	if len(bucket) > 0 {
		*large_disk = bucket
	}
	logrus.Infof("checkDynamicLabels. old_cpu:%s, old_memory:%s, cpu:%s, memory:%s, old_vsw:%s, old_sg:%s, vsw:%s, sg:%s, img:%s, target_label:%s, bucket:%s",
		*old_cpu, *old_memory, cpu, memory, *old_vsw, *old_sg, vsw, sg, img, target_label, bucket)
	return target_label
}

func (dis EciDispatcher) runnerInfoVerification(store common.Store, labels []string, repo_name, org_name string) (bool, string) {
	key, runner_type := store.GetKey()
	pat, _ := store.GetPat()
	specific_labels, _ := store.GetLabels()

	if len(key) <= 0 || len(runner_type) <= 0 || pat == "null" {
		return false, fmt.Sprint("org and repo token not exist. please run the 'register serverless runner' workflow first."+
			"Skip the runner creation. key: %s, type: %s, pat: %s", key, runner_type, pat)
	} else if runner_type == "pool" {
		return false, "pool should not be created here. skip it."
	} else if !dis.checkLabels(labels, repo_name, org_name, runner_type, specific_labels) {
		return false, "wf lablel dose not specify the repo/org name or runner label:" + specific_labels
	}
	return true, "runner info verification pass."
}

func (dis EciDispatcher) createRunner(act string, runer_id string, repo_name string,
	repo_url string, org_name string, owner_name string, labels []string) string {
	store := common.EnvStore(nil, org_name, repo_name)
	if verified, inf := dis.runnerInfoVerification(store, labels, repo_name, org_name); !verified {
		logrus.Warnf("fail to verify runner information. " + inf)
		return "fail to verify runner information. " + inf
	}
	if *tf_ctl == "go" {
		itf_ctl, err_tf, msg := dis.createRunnerTfCtl(store, act, runer_id, repo_name, repo_url, org_name, owner_name, labels)
		if err_tf != nil && itf_ctl != nil {
			tf_err_msg := err_tf.Error()
			is_file_busy := dis.sys_ctl.IsFileBusy(tf_err_msg)
			is_sys_busy := dis.sys_ctl.IsSysBusy(tf_err_msg)
			logrus.Warnf("tf gen error message is %s, is_file_busy %v, is_sys_busy %v",
				tf_err_msg, is_file_busy, is_sys_busy)
			if err := itf_ctl.MarkAsFinish("gen", is_sys_busy || is_file_busy); err != nil {
				logrus.Errorf("fail to mark gen as finish: %v, tf err:%v", err, err_tf)
				return err.Error() + msg + tf_err_msg
			}
			if is_sys_busy || is_file_busy {
				if is_file_busy {
					logrus.Warnf("file busy detected in tf msg %s", tf_err_msg)
					if err := itf_ctl.CleanTfAndLock(); err != nil {
						logrus.Errorf("fail to CleanTfAndLock: %v", err)
					}
				}
				logrus.Warnf("sys, file busy. begin plugin reload checking. msg:%s, err:%v", msg, err_tf)
				dis.sys_ctl.ReloadPlugin()
			}
		}
		logrus.Warnf("runner creation with tf controller. msg:%s, err:%v", msg, err_tf)
		return msg
	} else {
		// if need pool(not recommended), pls use cmd mode
		return dis.createRunnerCmd(store, act, runer_id, repo_name, repo_url, org_name, owner_name, labels)
	}
}

func (dis EciDispatcher) createRunnerTfCtl(store common.Store, act string, runer_id string, repo_name string,
	repo_url string, org_name string, owner_name string, labels []string) (tfc.ITfController, error, string) {
	logrus.Infof("Tf Controller Creating runner...")

	ctl := tfc.CreationController(dis.cur_path,
		map[string]string{"act": act, "runer_id": runer_id, "repo_name": repo_name,
			"repo_url": repo_url, "org_name": org_name, "owner_name": owner_name,
			"image_ver": dis.image_ver, "ctx_log_level": *ctx_log_level, "dis_ip": dis.dis_addr,
		}, store, *cloud_pr, dis.checkDynamicLabels, labels)

	logrus.Infof("Creating runner " + ctl.TfFilePath() + " with tf ctl...")
	if isfinish, err := ctl.Finished("gen"); !isfinish {
		if err != nil {
			return nil, err, err.Error()
		} else {
			msg := "can't gen lock" + ctl.TfFilePath()
			return nil, fmt.Errorf(msg), msg
		}
	}
	if err := ctl.GenTfConfigs(dis.cur_path + "runner/"); err != nil {
		return ctl, err, "fail to generate tf configurations for" + ctl.TfFilePath()
	}
	if err := ctl.InitTerraform(); err != nil {
		return ctl, err, "fail to init tf controller for" + ctl.TfFilePath()
	}
	if state, des := ctl.State(false); state {
		msg := des + " state of runner path exists. skip it: " + ctl.TfFilePath()
		return nil, nil, msg
	}

	initAndApply := func() (tfc.ITfController, error, string) {
		if err := ctl.Init(); err != nil {
			return ctl, err, "fail to init tf configurations."
		}
		if err := ctl.Apply(); err != nil {
			return ctl, err, "fail to apply tf configurations."
		}
		return nil, nil, "Success init and apply"
	}

	if inter_ctl, inter_err, inter_msg := initAndApply(); inter_err != nil {
		logrus.Errorf("initAndApply err "+ctl.TfFilePath()+", %s", inter_err.Error())
		if strings.Contains(inter_err.Error(),
			"installed provider plugins are not consistent") {
			logrus.Warnf("hcl inconsist with pr of " + ctl.TfFilePath() + ", clean and init apply again")
			inter_ctl.CleanHCL()
			if inter_ctl_hcl, inter_err_hcl, inter_msg_hcl := initAndApply(); inter_err_hcl != nil {
				logrus.Errorf("hcl clean fail "+ctl.TfFilePath()+", %s", inter_err_hcl.Error())
				return inter_ctl_hcl, inter_err_hcl, inter_msg_hcl
			}
		} else {
			return inter_ctl, inter_err, inter_msg
		}
	}

	if state, des := ctl.State(true); !state {
		msg := "fail to apply with nil err, but state of " + ctl.TfFilePath() + " does not contains service. " + des
		return ctl, fmt.Errorf(msg), msg

	}
	if !dis.r_competition {
		logrus.Warnf("runner competition not detected. safe reset destroy for %s", ctl.TfFilePath())
		store.ResetDestory(ctl.TfFilePath())
		if err := ctl.MarkAsFinish("del", false); err != nil {
			logrus.Errorf("fail to mark del as finish after creation of runner %s: %v", ctl.TfFilePath(), err)
		}
	}

	return nil, nil, "Finish runner creation." + ctl.TfFilePath()
}

func (dis EciDispatcher) createRunnerCmd(store common.Store, act string, runer_id string, repo_name string,
	repo_url string, org_name string, owner_name string, labels []string) string {
	logrus.Infof("CMD Creating runner...")

	key, runner_type := store.GetKey()
	pat, pat_type := store.GetPat()
	specific_labels, _ := store.GetLabels()
	reg_url, _ := store.GetUrl()
	cpu, _ := store.GetCpu()
	mem, _ := store.GetMemory()
	sec_gp_id, _ := store.GetSecGpId()
	vswitch_id, _ := store.GetVSwitchId()
	gcp_dind, _ := store.GetGcpDind()
	repo_image_ver, _ := store.GetImageVersion()
	image_ver := common.Ternary(repo_image_ver == "", dis.image_ver, repo_image_ver).(string)
	large_disk := ""

	if len(key) <= 0 || len(runner_type) <= 0 || pat == "null" {
		logrus.Warnf("Skip the runner creation. key: %s, type: %s, pat: %s", key, runner_type, pat)
		return "org and repo token not exist. please run the 'register serverless runner' workflow first."
	} else if runner_type == "pool" {
		return "pool should not be created here. skip it."
	} else if !dis.checkLabels(labels, repo_name, org_name, runner_type, specific_labels) {
		return "wf lablel dose not specify the repo/org name or runner label:" + specific_labels
	}

	specific_labels += dis.checkDynamicLabels(labels, &cpu, &mem, &vswitch_id, &sec_gp_id, &image_ver, &large_disk)

	logrus.Infof("Create runner. pat_type: %s, runner_type: %s", pat_type, runner_type)
	if pat_type != runner_type && (pat_type == "repo" && runner_type == "org") {
		runner_type = pat_type
		reg_url = reg_url + "/" + repo_name
	}

	sec, _ := store.GetSecret()
	region, _ := store.GetRegion()
	charge_labels, _ := store.GetChargeLabels()
	runner_group, _ := store.GetRunnerGroup()
	arm_client_id, _ := store.GetArmClientId()
	arm_client_secret, _ := store.GetArmClientSecret()
	arm_subscription_id, _ := store.GetArmSubscriptionId()
	arm_tenant_id, _ := store.GetArmTenantId()
	arm_environment, _ := store.GetArmEnvironment()
	arm_rp_registration, _ := store.GetArmRPRegistration()
	arm_resource_group_name, _ := store.GetArmResourceGroupName()
	arm_subnet_id, _ := store.GetArmSubnetID()
	arm_log_ana_workspace_id, _ := store.GetArmLogAnalyticsWorkspaceID()
	arm_log_ana_workspace_key, _ := store.GetArmLogAnalyticsWorkspaceKey()
	gcp_credentials, _ := store.GetGcpCredential()
	gcp_project, _ := store.GetGcpProject()
	gcp_region, _ := store.GetGcpRegion()
	gcp_sa, _ := store.GetGcpSA()
	gcp_apikey, _ := store.GetGcpApikey()
	gcp_vpc, _ := store.GetGcpVpc()
	gcp_subnet, _ := store.GetGcpSubnet()
	// following vars will not be sync into cmd and renewed by go clt
	aci_location, _ := store.GetAciLocation()
	aci_sku, _ := store.GetAciSku()
	aci_network_type, _ := store.GetAciNetworkType()
	out, err := exec.Command("/bin/bash", dis.cur_path+"create_runner.sh", act, runer_id,
		repo_name, reg_url, org_name, owner_name, pat, image_ver, key, sec, region,
		sec_gp_id, vswitch_id, runner_type, cpu, mem,
		common.Ternary(specific_labels == "", "none", strings.ReplaceAll(specific_labels, " ", "")).(string),
		common.Ternary(charge_labels == "", "none", strings.ReplaceAll(charge_labels, " ", "")).(string),
		common.Ternary(runner_group == "", "default", strings.ReplaceAll(runner_group, " ", "")).(string),
		*ctx_log_level, *cloud_pr,
		arm_client_id, arm_client_secret, arm_subscription_id, arm_tenant_id, arm_environment, arm_rp_registration,
		arm_resource_group_name, arm_subnet_id, arm_log_ana_workspace_id, arm_log_ana_workspace_key,
		gcp_credentials, gcp_project, gcp_region, gcp_sa, gcp_apikey, gcp_dind, gcp_vpc, gcp_subnet,
		aci_location, aci_sku, aci_network_type, large_disk,
	).Output()
	if err != nil {
		logrus.Errorf("error %s", err)
	} else {
		logrus.Infof("ResetDestory " + repo_name + "-" + runer_id)
		store.ResetDestory(repo_name + "-" + runer_id)
	}
	output := string(out)
	logrus.Warnf("Finish runner creation, output: %s", output)

	return output
}

func (dis EciDispatcher) removeRunner(act string, runer_name string, repo_name string,
	org_name string, run_wf string, labels []string, url string, owner string) string {
	logrus.Printf("removeRunner paras: act %s, runer_name %s, repo_name %s, org_name %s, run_wf %s, labels %v, url %s, owner %s",
		act, runer_name, repo_name, org_name, run_wf, labels, url, owner)
	store := common.EnvStore(nil, org_name, repo_name)
	if store.IsDestory(runer_name) {
		logrus.Warnf("workflow " + repo_name + "-" + run_wf + " occupied runner " + runer_name + " already removed.")
		return "workflow " + repo_name + "-" + run_wf + " occupied runner " + runer_name + " already removed."
	}
	if *tf_ctl == "go" {
		itf_ctl, err_tf, msg := dis.removeRunnerTfCtl(store, act, runer_name, repo_name,
			org_name, run_wf, labels, url, run_wf, owner)
		if itf_ctl != nil {
			if err_tf != nil {
				tf_err_msg := err_tf.Error()
				is_file_busy := dis.sys_ctl.IsFileBusy(tf_err_msg)
				is_sys_busy := dis.sys_ctl.IsSysBusy(tf_err_msg)
				logrus.Warnf("tf del error message is %s", tf_err_msg)
				if err := itf_ctl.MarkAsFinish("del", is_sys_busy || is_file_busy); err != nil {
					logrus.Warnf("after desrtoy, fail to mark del as finish" + err.Error() + err_tf.Error() + msg)
					return err.Error() + err_tf.Error() + msg
				}
				if is_file_busy || is_sys_busy {
					logrus.Warnf("sys,file busy during destroy. plugin reload checking. msg:%s, err:%v", msg, err_tf)
					dis.sys_ctl.ReloadPlugin()
				}
			}
			if err := itf_ctl.MarkAsFinish("gen", false); err != nil {
				logrus.Warnf("after desrtoy, fail to mark gen as finish" + err.Error() + err_tf.Error() + msg)
				return err.Error() + err_tf.Error() + msg
			}
		}
		logrus.Warnf("runner removing with tf controller. msg:%s, err:%s", msg, err_tf)
		return msg
	} else {
		return dis.removeRunnerCmd(store, act, runer_name, repo_name, org_name, run_wf)
	}
}

func (dis EciDispatcher) removeRunnerTfCtl(store common.Store, act string, runer_name string, repo_name string,
	org_name string, run_wf string, labels []string, url string, runner_id string,
	owner string) (tfc.ITfController, error, string) {
	run_on := common.Ternary(len(runer_name) == 0, repo_name+"-"+run_wf, runer_name).(string)
	logrus.Infof("Removing runner " + dis.cur_path + run_on + " with tf ctl...")
	ctl := tfc.DestroyController(run_on, dis.cur_path+run_on, store, *cloud_pr, dis.cur_path)
	if ctl == nil {
		return nil, nil, "org and repo token dose not exist, please register."
	}
	if isfinish, err := ctl.Finished("del"); !isfinish {
		if err != nil {
			return nil, err, err.Error()
		} else {
			msg := "can't del lock" + repo_name + "-" + run_wf
			return nil, fmt.Errorf(msg), msg
		}
	}
	if !ctl.TfConfigsExists() {

		msg := "Runner config dose not exists. Skip destroy."
		return nil, fmt.Errorf(msg), msg
	}
	if err := ctl.InitTerraform(); err != nil {
		msg := "fail to init Tf controller for destroy."
		return ctl, fmt.Errorf(msg), msg
	}
	if state, des := ctl.State(false); !state {
		msg := des + " runner path: " + common.Ternary(len(runer_name) == 0, repo_name+"-"+run_wf, runer_name).(string)

		return ctl, fmt.Errorf(msg), msg
	}
	if dis.sys_ctl.ExceedReload() && !strings.Contains(runer_name, run_wf) && !strings.Contains(runer_name, "sls-comp") {
		if exist, _ := ctl.FileState(dis.cur_path + run_wf); !exist {
			logrus.Warnf("runner competition detected run on %s, wf %s", runer_name, run_wf)
			itf_ctl, err_tf, msg := dis.createRunnerTfCtl(store, act, run_wf+"-sls-comp-"+dis.crypt_ctl.RandStr(4), repo_name, url, org_name, owner, labels)
			if err_tf != nil && itf_ctl != nil {
				tf_err_msg := err_tf.Error()
				logrus.Warnf("runner competition, tf gen error message is %s, msg %s", tf_err_msg, msg)
				if err := itf_ctl.MarkAsFinish("gen", dis.sys_ctl.IsSysBusy(tf_err_msg)); err != nil {
					logrus.Errorf("runner competition, fail to mark gen as finish: %s, tf err:%s", err, err_tf)
				}
				return ctl, err_tf, "fail to create competition runner"
			}
		}
	}
	if err := ctl.Destroy(); err != nil {
		return ctl, err, "fail to Destroy runner. " + err.Error()
	}
	store.MarkDestory(run_on)
	return ctl, nil, "Success remove runner " + run_on
}

func (dis EciDispatcher) removeRunnerCmd(store common.Store, act string, runer_name string, repo_name string,
	org_name string, run_wf string) string {
	logrus.Infof("CMD Remove runner...")
	key, _ := store.GetKey()
	sec, _ := store.GetSecret()
	region, _ := store.GetRegion()
	arm_client_id, _ := store.GetArmClientId()
	arm_client_secret, _ := store.GetArmClientSecret()
	arm_subscription_id, _ := store.GetArmSubscriptionId()
	arm_tenant_id, _ := store.GetArmTenantId()
	arm_environment, _ := store.GetArmEnvironment()
	arm_rp_registration, _ := store.GetArmRPRegistration()
	arm_subnet_id, _ := store.GetArmSubnetID()
	arm_log_ana_workspace_id, _ := store.GetArmLogAnalyticsWorkspaceID()
	arm_log_ana_workspace_key, _ := store.GetArmLogAnalyticsWorkspaceKey()
	if len(key) <= 0 {
		return "org and repo token not exist. please run the 'register serverless runner' workflow first."
	} else if store.IsDestory(repo_name + "-" + run_wf) {
		logrus.Infof("workflow " + repo_name + "-" + run_wf + " occupied runner " + runer_name + " already removed.")
		return "workflow " + repo_name + "-" + run_wf + " occupied runner " + runer_name + " already removed."
	}
	out, err := exec.Command("/bin/bash", dis.cur_path+"remove_runner.sh", act, runer_name, key, sec,
		region, org_name, repo_name, run_wf, *cloud_pr,
		arm_client_id, arm_client_secret, arm_subscription_id, arm_tenant_id, arm_environment,
		arm_rp_registration, arm_subnet_id, arm_log_ana_workspace_id, arm_log_ana_workspace_key).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	} else {
		logrus.Infof("MarkDestory " + repo_name + "-" + run_wf)
		store.MarkDestory(repo_name + "-" + run_wf)
	}
	output := string(out)
	logrus.Warnf("Remove runner finish, output: %s", output)
	return output
}
