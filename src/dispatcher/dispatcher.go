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
	"slices"
	"strconv"
	"strings"

	ali_mns "github.com/aliyun/aliyun-mns-go-sdk"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type Dispatcher interface {
	HandleEvents(w http.ResponseWriter, req *http.Request)
	Refresh()
}

type EciDispatcher struct {
	cur_path       string
	pool_prefix    string
	interval       int64
	image_ver      string
	lazy_regs      string
	default_labels []string
	allen_regs     string
}

func EciDispatcherConstruct(image_ver string, lazy_regs string, allen_regs string) Dispatcher {
	return EciDispatcher{"/go/bin/", agent.NOTIFICATION_Q, int64(10), image_ver,
		lazy_regs, []string{"serverless-hosted-runner", "eci-runner"}, allen_regs}
}

func FnDispatcherConstruct() Dispatcher {
	return nil
}

func (dis EciDispatcher) response_back(w http.ResponseWriter, msg string, logstr string, status int) {
	logrus.Infof(logstr)
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte(msg))
}

func (dis EciDispatcher) parse_registration(item common.PoolMsg) {
	logrus.Infof("parse_registration, Repos: %s", item.Repos)
	repos := strings.Split(item.Repos, ",")
	url := item.Url
	for _, r := range repos {
		if len(r) > 0 {
			item.Type = "Repo"
			item.Name = r
			item.Url = url + "/" + r
			logrus.Infof("parse_registration, item.Name: %s", item.Name)

			store := common.EnvStore(&item, item.Name, r)
			store.Save()
			key, runner_type := store.GetKey()
			pat, pat_type := store.GetPat()
			labels, label_type := store.GetLabels()
			logrus.Infof("parse_registration, key: %s, pat: %s, labels: %s, runner_type %s, pat_type %s, label_type %s",
				key, pat, labels, runner_type, pat_type, label_type)

			iv, _ := strconv.Atoi(item.PullInterval)
			wf := agent.CreateWorkflowAgent(item.Type, item.Name, item.Url, dis.createRunner,
				dis.removeRunner, dis.notify_release, dis.check_labels, r, item.Name, iv, labels)
			wf.InitAgent()
			wf.MonitorOnAgent()
			wf = nil
		}
	}
}

func (dis EciDispatcher) lazy_registration() {
	logrus.Infof("lazy_registration start. lazy_regs: %s", dis.lazy_regs)
	arr_lazy_regs := []common.PoolMsg{}
	_ = json.Unmarshal([]byte(dis.lazy_regs), &arr_lazy_regs)
	if len(arr_lazy_regs) > 0 {
		for _, item := range arr_lazy_regs {
			dis.parse_registration(item)
		}
	}
}

func (dis EciDispatcher) allen_registration() {
	logrus.Infof("allen registration start")
	aln := agent.CreateAllenStoreAgent(dis.parse_registration)
	aln.InitAgent()
	aln.MonitorOnAgent()
}

func (dis EciDispatcher) Refresh() {
	logrus.Infof("refresh pool start")
	common.SetContextLogLevel(*ctx_log_level)

	if dis.lazy_regs != "" && dis.lazy_regs != "none" {
		go dis.lazy_registration()
	}
	if dis.allen_regs == "allen" {
		go dis.allen_registration()
	}

	if *pool_mode {
		qAgent := agent.CreateAliMNSAgent(os.Getenv("TF_VAR_MNS_URL"), os.Getenv("ALICLOUD_ACCESS_KEY"),
			os.Getenv("ALICLOUD_SECRET_KEY"), agent.DEFAULT_POOL_Q, dis.check_msg, nil)
		qAgent.MonitorOnAgent()
	}
}

func (dis EciDispatcher) update_pool(msg common.PoolMsg, store common.Store) {
	logrus.Infof("update pool begin...")
	num, err := strconv.Atoi(store.GetPreSize())
	if err != nil {
		logrus.Errorf("update_pool strconv failure: %s", err)
		num = 0
	}
	logrus.Infof("update_pool Size %s, Type %s, Name %s, Pat %s, AnyChange %t", msg.Size,
		msg.Type, msg.Name, msg.Pat, store.AnyChange())
	if msg.Size == "0" && msg.Type == "Pool" {
		logrus.Infof("release pool, name: %s, num: %d", msg.Name, num)
		go dis.release_pool(msg.Name, num)
	} else if msg.Type == "Pool" && msg.Name != "" && msg.Pat != "null" && store.AnyChange() {
		logrus.Infof("release/recreate pool, name: %s, num: %d", msg.Name, num)
		dis.release_pool(msg.Name, num) // upinsert
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
			msg.GcpCredential, msg.GcpProject, msg.GcpRegion).Output()
		if err != nil {
			logrus.Errorf("error %s", err)
		}
		fmt.Printf("pool creation %s", out)
	}
}

func (dis EciDispatcher) notify_release(msg string) {
	if *event_push {
		qAgent := agent.CreateAliMNSAgent(os.Getenv("TF_VAR_MNS_URL"), os.Getenv("ALICLOUD_ACCESS_KEY"),
			os.Getenv("ALICLOUD_SECRET_KEY"), agent.NOTIFICATION_Q, nil, nil)
		qAgent.NotifyAgent(msg)
		logrus.Infof("notify_release msg: %s", msg)
	}
}

func (dis EciDispatcher) release_pool(org_name string, num int) {
	// TODO. if the previous pool not consume the event, new created pool would be deleted.
	p_name := dis.pool_prefix + "-" + org_name
	for id := 1; id <= num; id++ {
		dis.notify_release(p_name + "-" + strconv.Itoa(id))
	}
	fmt.Printf("release pool. output - %s\n",
		dis.removeRunner("pool_completed", p_name, "", org_name, p_name))
}

func (dis EciDispatcher) msg_invisible(t string) int64 {
	return common.Ternary(t == "Pool", int64(60), int64(5)).(int64)
}

func (dis EciDispatcher) check_msg(obj interface{}, para interface{}) bool {
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
				if ret, e := q.ChangeMessageVisibility(resp.ReceiptHandle, dis.msg_invisible(msg.Type)); e != nil {
					fmt.Println("visibility error", e)
				} else {
					fmt.Println("visibility changed", ret, "delete msg now:", ret.ReceiptHandle)
					store := common.EnvStore(&msg, msg.Name, msg.Name)
					store.Save()
					if e := q.DeleteMessage(ret.ReceiptHandle); e != nil {
						fmt.Println(e)
					}
					go dis.update_pool(msg, store)
					if msg.Type == "Repo" {
						iv, _ := strconv.Atoi(msg.PullInterval)
						wf := agent.CreateWorkflowAgent(msg.Type, msg.Name, msg.Url, dis.createRunner,
							dis.removeRunner, dis.notify_release, dis.check_labels, "", "", iv, msg.Labels)
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
		dis.response_back(w, "Event Parsing Error:", "Parsing event body error.", http.StatusBadRequest)
		return
	}
	github_event := req.Header["X-Github-Event"][0]
	if github_event != "workflow_job" && github_event != "ping" {
		dis.response_back(w, "Unsupported Event", "Request is not workflow_job or ping event.", http.StatusBadRequest)
		return
	} else if github_event == "ping" {
		dis.response_back(w, "Ping Finish", "The dispatcher running normally.", http.StatusOK)
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
		dis.notify_release(event_data.WorkflowJob.RunnerName)
		if !strings.Contains(event_data.WorkflowJob.RunnerName, dis.pool_prefix) {
			fmt.Printf("remove a runner. output - %s\n",
				dis.removeRunner(event_data.Action, event_data.WorkflowJob.RunnerName, repo_name, org_name,
					strconv.FormatInt(event_data.WorkflowJob.RunID, 10)+"-"+strconv.FormatInt(event_data.WorkflowJob.ID, 10)))
		}
	} else {
		fmt.Printf("skip the action - %s\n", event_data.Action)
	}
	dis.response_back(w, "Exist dispatcher.", "The dispatcher event handle finished.", http.StatusOK)
}

func (dis EciDispatcher) check_labels(labels []string, repo_name string,
	org_name string, runner_type string, specific_labels string) bool {
	for idx, item := range labels {
		logrus.Infof("#%d label: %s", idx, item)
		if strings.Contains(item, ",") {
			item_arr := strings.Split(strings.TrimSpace(item), ",")
			labels = append(labels, item_arr...)
		}
	}
	logrus.Infof("check_labels. repo_name: %s, org_name: %s, runner_type: %s, specific_labels: %s",
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
func (dis EciDispatcher) check_labels_cpu_memory(labels []string, old_cpu *string,
	old_memory *string) string {
	cpu, memory, target_label := "", "", ""
	for _, label := range labels {
		if strings.Contains(label, "cpu-") {
			cpu = strings.ReplaceAll(label, "cpu-", "")
			target_label += "," + label
		} else if strings.Contains(label, "memory-") {
			memory = strings.ReplaceAll(label, "memory-", "")
			target_label += "," + label
		}
	}
	if len(cpu) > 0 && len(memory) > 0 {
		logrus.Infof("check_labels_cpu_memory. old_cpu: %s, old_memory: %s, cpu: %s, memory: %s, target_label: %s",
			*old_cpu, *old_memory, cpu, memory, target_label)
		*old_cpu = cpu
		*old_memory = memory
		return target_label
	}
	return ""
}

func (dis EciDispatcher) createRunner(act string, runer_id string, repo_name string,
	repo_url string, org_name string, owner_name string, labels []string) string {
	store := common.EnvStore(nil, org_name, repo_name)
	key, runner_type := store.GetKey()
	sec, _ := store.GetSecret()
	region, _ := store.GetRegion()
	sec_gp_id, _ := store.GetSecGpId()
	reg_url, _ := store.GetUrl()
	vswitch_id, _ := store.GetVSwitchId()
	pat, pat_type := store.GetPat()
	cpu, _ := store.GetCpu()
	mem, _ := store.GetMemory()
	specific_labels, _ := store.GetLabels()
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

	if len(key) <= 0 || len(runner_type) <= 0 || pat == "null" {
		logrus.Warnf("Skip the runner creation. key: %s, type: %s, pat: %s", key, runner_type, pat)
		return "org and repo token not exist. please run the 'register serverless runner' workflow first."
	} else if runner_type == "pool" {
		return "pool should not be created here. skip it."
	} else if !dis.check_labels(labels, repo_name, org_name, runner_type, specific_labels) {
		return "wf lablel dose not specify the repo/org name or runner label:" + specific_labels
	}
	specific_labels += dis.check_labels_cpu_memory(labels, &cpu, &mem)

	logrus.Infof("Create runner. pat_type: %s, runner_type: %s", pat_type, runner_type)
	if pat_type != runner_type && (pat_type == "repo" && runner_type == "org") {
		runner_type = pat_type
		reg_url = reg_url + "/" + repo_name
	}

	logrus.Infof("Creating runner...")
	out, err := exec.Command("/bin/bash", dis.cur_path+"create_runner.sh", act, runer_id,
		repo_name, reg_url, org_name, owner_name, pat, dis.image_ver, key, sec, region,
		sec_gp_id, vswitch_id, runner_type, cpu, mem,
		common.Ternary(specific_labels == "", "none", strings.ReplaceAll(specific_labels, " ", "")).(string),
		common.Ternary(charge_labels == "", "none", strings.ReplaceAll(charge_labels, " ", "")).(string),
		common.Ternary(runner_group == "", "default", strings.ReplaceAll(runner_group, " ", "")).(string),
		*ctx_log_level, *cloud_pr,
		arm_client_id, arm_client_secret, arm_subscription_id, arm_tenant_id, arm_environment, arm_rp_registration,
		arm_resource_group_name, arm_subnet_id, arm_log_ana_workspace_id, arm_log_ana_workspace_key,
		gcp_credentials, gcp_project, gcp_region,
	).Output()
	if err != nil {
		logrus.Errorf("error %s", err)
	}
	output := string(out)
	logrus.Warnf("Finish runner creation, output: %s", output)
	return output
}

func (dis EciDispatcher) removeRunner(act string, runer_name string, repo_name string,
	org_name string, run_wf string) string {
	store := common.EnvStore(nil, org_name, repo_name)
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
	}

	logrus.Infof("Remove runner...")
	out, err := exec.Command("/bin/bash", dis.cur_path+"remove_runner.sh", act, runer_name, key, sec,
		region, org_name, repo_name, run_wf, *cloud_pr,
		arm_client_id, arm_client_secret, arm_subscription_id, arm_tenant_id, arm_environment,
		arm_rp_registration, arm_subnet_id, arm_log_ana_workspace_id, arm_log_ana_workspace_key).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output := string(out)
	logrus.Warnf("Remove runner finish, output: %s", output)
	return output
}
