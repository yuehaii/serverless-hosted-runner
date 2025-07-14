package tfcontroller

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"strings"
	"time"

	common "serverless-hosted-runner/common"

	"github.com/gofrs/flock"
	jwt "github.com/golang-jwt/jwt/v5"
	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
	cp "github.com/otiai10/copy"
)

var (
	tflocks_pool = make(map[string]*flock.Flock)
	sys_delay    = make(map[string]int)
)

type ITfController interface {
	Envs(map[string]string)
	EnvsBase64(map[string]string)
	GenTfConfigs(string) error
	TfConfigsExists() bool
	InitTerraform() error
	Init() error
	Plan() (plan_diff bool, err error)
	Apply() error
	Destroy() error
	State(bool) (bool, string)
	Finished(string) (bool, error)
	MarkAsFinish(string, bool) error
	TfFilePath() string
	CleanTfAndLock() error
	FileState(string) (bool, string)
	CleanHCL() error
}

type TfController struct {
	install        bool
	tf_version     string
	exec_path      string
	file_path      string
	working_dir    string
	tf             *tfexec.Terraform
	locktimeout    string
	vars           map[string]string
	var_file       string
	cache_dir      string
	pr_services    map[string]string
	pr             string
	prebuild_cache bool
	match_hcl      bool
	store          common.Store
	hcl_filepath   string
	flock_prefix   string
	flock_suffix   string
	delay_init     int
	delay_delta    int
	dis_dir        string
}

func CreationController(working_dir string, vars map[string]string, store common.Store, pr string,
	dynamic_labels func([]string, *string, *string, *string, *string, *string, *string) string, labels []string) ITfController {
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
	image_ver := common.Ternary(repo_image_ver == "", vars["image_ver"], repo_image_ver).(string)
	large_disk := ""
	specific_labels += dynamic_labels(labels, &cpu, &mem, &vswitch_id, &sec_gp_id, &image_ver, &large_disk)
	vars["image_ver"] = image_ver //  label > repo > default dispacher ver
	logrus.Infof("pat_type: %s, runner_type: %s", pat_type, runner_type)
	if pat_type != runner_type && (pat_type == "repo" && runner_type == "org") {
		runner_type = pat_type
		reg_url = reg_url + "/" + vars["repo_name"]
	}
	vars_append := map[string]string{
		"key": key, "specific_labels": specific_labels, "runner_type": runner_type,
		"pat": pat, "pat_type": pat_type, "reg_url": reg_url, "cpu": cpu, "mem": mem,
		"gcp_dind": gcp_dind, "vswitch_id": vswitch_id, "sec_gp_id": sec_gp_id,
		"large_disk": large_disk,
	}
	maps.Copy(vars, vars_append)
	working_path := vars["repo_name"] + "-" + vars["runer_id"]
	if vars["runner_type"] == "org" {
		working_path = vars["org_name"] + "-" + vars["runer_id"]
	}
	if applyMarker(working_dir+working_path, false) {
		vars["runer_id"] += "-sls-comp"
		working_path += "-sls-comp"
	}
	return tfController(working_path, working_dir+working_path, vars, store, strings.TrimSpace(pr), working_dir)
}

func DestroyController(file_path, working_dir string, store common.Store, pr string, dis_dir string) ITfController {
	gcp_dind, _ := store.GetGcpDind()
	key, _ := store.GetKey()
	if len(key) <= 0 {
		logrus.Warnf("org and repo token not exist. please run the 'register serverless runner' workflow first.")
		return nil
	}
	vars := map[string]string{
		"key": key, "gcp_dind": gcp_dind,
	}
	return tfController(file_path, working_dir, vars, store, strings.TrimSpace(pr), dis_dir)
}

func tfController(file_path, working_dir string, vars map[string]string, store common.Store, pr string, dis_dir string) ITfController {

	return &TfController{false, "1.10.5", "/usr/local/bin/terraform", file_path, working_dir, nil,
		"20m", vars, "ubuntu_runner.tfvars", dis_dir + "tf_plugin_cache",
		map[string]string{"ali": "alicloud_eci_container_group", "azure": "azurerm_container_group",
			"gcp": "google_cloud_run", "gcp_dind": "gcp_runner_batch_job_module"}, pr, false, false,
		store, "/.terraform.lock.hcl", "/var/lock/sls-", ".lock.", 300, 60, dis_dir}
}

func applyMarker(working_dir string, init bool) bool {
	_, err := os.Stat(working_dir)
	if err != nil {
		logrus.Errorf("apply marker, runner directory dose not exists: " + working_dir)
		return false
	}
	marker_f := working_dir + "/.marker"
	if init {
		marker_f, err := os.Create(marker_f)
		if err == nil {
			logrus.Warnf("marker initialized for %s", working_dir)
			defer marker_f.Close()
		} else {
			logrus.Errorf("fail to create apply marker for %v" + working_dir)
		}
		return false
	} else {
		f_des, err := os.Stat(marker_f)
		if err != nil {
			logrus.Errorf("marker file dose not exists: " + working_dir)
			return false
		} else {
			cur_t := time.Now()
			f_t := f_des.ModTime()
			logrus.Warnf("marker %s, cur_t %v, f_t %v, sub %v mins", working_dir, cur_t, f_t, cur_t.Sub(f_t).Minutes())
			return cur_t.Sub(f_t).Minutes() > 5
		}
	}
}

func (tfc *TfController) genAccessToken(cred string, key string) string {
	jctl := common.CreateGcpJsWebTokenCtl(jwt.SigningMethodRS256, key, cred)
	if jctl == nil {
		logrus.Errorf("fail to create jwt controller")
		return ""
	}
	if err := jctl.SetIssFromGCPCredential(); err != nil {
		logrus.Errorf("fail to get iss from gcp credential, %s", err)
		return ""
	}
	if tk, _ := jctl.ExchangeApiKey(); len(tk) == 0 {
		logrus.Warnf("fail to gen gcp token with go ctl.")
		return ""
	} else {
		logrus.Infof("success gen gcp token with go ctl: '%s'", tk)
		return tk
	}
}

func (tfc *TfController) cpTfConfig(src, dest string) error {
	_, err := os.Stat(src)
	if err == nil {
		src_file, err := os.Open(src)
		if err == nil {
			defer src_file.Close()
			dest_file, err := os.Create(dest)
			if err == nil {
				defer dest_file.Close()
				lens, err := io.Copy(dest_file, src_file)
				if err == nil {
					logrus.Infof("success cp %v size file from %s to %s", lens, src, dest)
				}
			}
		}
	}
	return err
}

func (tfc *TfController) setupCache() {
	if len(tfc.cache_dir) > 0 {
		tfc.tf.SetEnv(map[string]string{
			"TF_PLUGIN_CACHE_DIR": tfc.cache_dir,
			// it will cause unmatch issue in gcp pr
			"TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE": common.Ternary(tfc.pr == "ali", "true", "").(string),
		})
		_, err := os.Stat(tfc.cache_dir)
		if err != nil {
			if err := os.MkdirAll(tfc.cache_dir, os.ModePerm); err != nil {
				logrus.Errorf("fail to create cache dir: %s", err)
			}
		} else {
			logrus.Infof("cache dir already exists.")
		}
	}
}

func (tfc *TfController) preBuildHcl() {
	cacheHclPath := tfc.cache_dir + tfc.hcl_filepath
	initOps := []tfexec.InitOption{
		tfexec.Upgrade(false),
	}

	runner_dir := tfc.dis_dir + "runner/" + tfc.pr
	tf, err := tfexec.NewTerraform(runner_dir, tfc.exec_path)
	if err != nil {
		logrus.Errorf(runner_dir+" fail to new tf to create cache: %s", err)
		return
	}
	err = tf.Init(context.Background(), initOps...)
	if err != nil {
		logrus.Errorf(tfc.cache_dir+" fail to run init to create cache: %s", err)
	}
	err = tfc.cpTfConfig(runner_dir+tfc.hcl_filepath, cacheHclPath)
	if err != nil {
		logrus.Errorf(cacheHclPath+" can't be setup match. %s", err)
	}
	os.Remove(runner_dir + tfc.hcl_filepath)
	os.RemoveAll(runner_dir + "/.terraform")
}

func (tfc TfController) CleanHCL() error {
	return os.Remove(tfc.working_dir + tfc.hcl_filepath)
}

func (tfc *TfController) preMatchHcl() {
	tfc.setupCache()
	if tfc.prebuild_cache {
		tfc.preBuildHcl()
	}
	if tfc.match_hcl {
		tfc.matchHcl()
	}
}

func (tfc *TfController) postMatchHcl() {
	if tfc.match_hcl {
		tfc.matchHcl()
	}
}

// match hcl can reduce the pr version not match caused delay
func (tfc *TfController) matchHcl() {
	runnerHclPath := tfc.working_dir + tfc.hcl_filepath
	cacheHclPath := tfc.cache_dir + tfc.hcl_filepath
	_, err_cache_lock := os.Stat(cacheHclPath)
	_, err_runner_lock := os.Stat(runnerHclPath)
	if err_cache_lock == nil && err_runner_lock != nil {
		logrus.Infof("match hcl from cache to runner " + tfc.working_dir)
		err := tfc.cpTfConfig(cacheHclPath, runnerHclPath)
		if err != nil {
			logrus.Errorf("fail to match hcl from cache to runner: %s", err)
		}
	} else if err_cache_lock != nil && err_runner_lock == nil {
		logrus.Infof("match hcl from runner to cache " + tfc.working_dir)
		err := tfc.cpTfConfig(runnerHclPath, cacheHclPath)
		if err != nil {
			logrus.Errorf("fail to init cache hcl from runner to cache: %s", err)
		}
	} else if err_cache_lock != nil && err_runner_lock != nil {
		logrus.Infof("%s, %s, hcl dose not exists for cache and runner. welcome first runner."+tfc.working_dir,
			err_cache_lock, err_runner_lock)
	}

}

func (tfc *TfController) buildVars(op string) (error, []string) {
	if op == "init" {
		return nil, []string{}
	} else if op == "apply" {
		return tfc.buildApplyVars()
	} else if op == "destroy" {
		return tfc.buildDestroyVars()
	}
	return fmt.Errorf("invalid build var operation, should be init/apply/destroy"), []string{}
}

func (tfc *TfController) buildDestroyVars() (error, []string) {
	vars := []string{}
	if tfc.pr == "azure" {
		arm_subnet_id, _ := tfc.store.GetArmSubnetID()
		arm_log_ana_workspace_id, _ := tfc.store.GetArmLogAnalyticsWorkspaceID()
		arm_log_ana_workspace_key, _ := tfc.store.GetArmLogAnalyticsWorkspaceKey()
		vars = []string{"subnet_ids=" + arm_subnet_id, "workspace_id=" + arm_log_ana_workspace_id,
			"workspace_key=" + arm_log_ana_workspace_key}
	}
	return nil, vars
}

func (tfc *TfController) buildApplyVars() (error, []string) {
	runner_org := "none"
	charge_labels, _ := tfc.store.GetChargeLabels()
	runner_group, _ := tfc.store.GetRunnerGroup()
	arm_resource_group_name, _ := tfc.store.GetArmResourceGroupName()
	arm_subnet_id, _ := tfc.store.GetArmSubnetID()
	arm_log_ana_workspace_id, _ := tfc.store.GetArmLogAnalyticsWorkspaceID()
	arm_log_ana_workspace_key, _ := tfc.store.GetArmLogAnalyticsWorkspaceKey()
	gcp_project, _ := tfc.store.GetGcpProject()
	gcp_region, _ := tfc.store.GetGcpRegion()
	gcp_sa, _ := tfc.store.GetGcpSA()
	gcp_vpc, _ := tfc.store.GetGcpVpc()
	gcp_subnet, _ := tfc.store.GetGcpSubnet()
	aci_location, _ := tfc.store.GetAciLocation()
	aci_sku, _ := tfc.store.GetAciSku()
	aci_network_type, _ := tfc.store.GetAciNetworkType()
	if tfc.vars["runner_type"] == "org" {
		runner_org = tfc.vars["org_name"]
	}

	vars := []string{"runner_id=" + tfc.vars["runer_id"], "runner_repname=" + tfc.vars["repo_name"], "runner_orgowner=" + tfc.vars["owner_name"],
		"runner_action=" + tfc.vars["act"], "runner_repurl=" + tfc.vars["reg_url"], "runner_token=" + tfc.vars["pat"], "image_ver=" + tfc.vars["image_ver"],
		"security_group_id=" + tfc.vars["sec_gp_id"], "vswitch_id=" + tfc.vars["vswitch_id"], "container_type=" + tfc.vars["runner_type"],
		"runner_orgname=" + runner_org, "network_mode=fixed", "runner_cpu=" + tfc.vars["cpu"], "runner_memory=" + tfc.vars["mem"],
		"runner_labels=" + common.Ternary(tfc.vars["specific_labels"] == "", "none", strings.ReplaceAll(tfc.vars["specific_labels"], " ", "")).(string),
		"charge_labels=" + common.Ternary(charge_labels == "", "none", strings.ReplaceAll(charge_labels, " ", "")).(string),
		"runner_group=" + common.Ternary(runner_group == "", "default", strings.ReplaceAll(runner_group, " ", "")).(string),
		"oss_mount=" + tfc.vars["large_disk"], "cloud_pr=" + tfc.pr, "dis_ip=" + tfc.vars["dis_ip"],
		"ctx_log_level=" + tfc.vars["ctx_log_level"], "subnet_ids=" + arm_subnet_id, "resource_group_name=" + arm_resource_group_name,
		"workspace_id=" + arm_log_ana_workspace_id, "workspace_key=" + arm_log_ana_workspace_key,
		"aci_location=" + aci_location, "aci_sku=" + aci_sku, "aci_network_type=" + aci_network_type,
		"gcp_project=" + gcp_project, "gcp_region=" + gcp_region, "gcp_project_sa_email=" + gcp_sa, "gcp_project_apikey=" + tfc.vars["gcp_project_apikey"],
		"gcp_runner_dind=" + tfc.vars["gcp_dind"], "gcp_vpc=" + gcp_vpc, "gcp_subnet=" + gcp_subnet,
		"IMAGE_RETRIEVE_SERVER=" + os.Getenv("TF_VAR_IMAGE_RETRIEVE_SERVER"),
		"IMAGE_RETRIEVE_USERNAME=" + os.Getenv("TF_VAR_IMAGE_RETRIEVE_USERNAME"),
		"IMAGE_RETRIEVE_PWD=" + os.Getenv("TF_VAR_IMAGE_RETRIEVE_PWD"),
	}
	return nil, vars
}

func (tfc *TfController) buildEnvs() {
	sec, _ := tfc.store.GetSecret()
	region, _ := tfc.store.GetRegion()
	arm_client_id, _ := tfc.store.GetArmClientId()
	arm_client_secret, _ := tfc.store.GetArmClientSecret()
	arm_subscription_id, _ := tfc.store.GetArmSubscriptionId()
	arm_tenant_id, _ := tfc.store.GetArmTenantId()
	arm_environment, _ := tfc.store.GetArmEnvironment()
	arm_rp_registration, _ := tfc.store.GetArmRPRegistration()
	gcp_credentials, _ := tfc.store.GetGcpCredential()
	gcp_project, _ := tfc.store.GetGcpProject()
	gcp_region, _ := tfc.store.GetGcpRegion()
	gcp_sa, _ := tfc.store.GetGcpSA()
	gcp_apikey, _ := tfc.store.GetGcpApikey()
	if tfc.pr == "gcp" || tfc.pr == "gcp_dind" {
		tfc.Envs(map[string]string{
			"GOOGLE_PROJECT":          gcp_project,
			"GOOGLE_REGION":           gcp_region,
			"GOOGLE_PROJECT_SA_EMAIL": gcp_sa,
			"GOOGLE_RUNNER_DIND":      tfc.vars["gcp_dind"],
		})
		tfc.EnvsBase64(map[string]string{
			"GOOGLE_CREDENTIALS":            gcp_credentials,
			"GOOGLE_CREDENTIALS_PRIVATEKEY": gcp_apikey,
		})
	} else if tfc.pr == "azure" {
		tfc.Envs(map[string]string{
			"ARM_CLIENT_ID":                       arm_client_id,
			"ARM_CLIENT_SECRET":                   arm_client_secret,
			"ARM_SUBSCRIPTION_ID":                 arm_subscription_id,
			"ARM_TENANT_ID":                       arm_tenant_id,
			"ARM_ENVIRONMENT":                     arm_environment,
			"ARM_RESOURCE_PROVIDER_REGISTRATIONS": arm_rp_registration,
		})
	} else if tfc.pr == "ali" {
		tfc.Envs(map[string]string{
			// ali
			"ALICLOUD_ACCESS_KEY": tfc.vars["key"],
			"ALICLOUD_SECRET_KEY": sec,
			"ALICLOUD_REGION":     region,
		})
	}
}

func (tfc TfController) appendApplyLog() (apdlogs string) {
	if tfc.pr == "gcp_dind" {
		f_bytes_code, err := os.ReadFile(tfc.working_dir + "/http_response_code.log")
		if err == nil {
			apdlogs += string(f_bytes_code)
		} else {
			logrus.Errorf("fail to get the log %s, %s", tfc.working_dir+"/http_response_code.log", err)
		}

		f_bytes_body, err := os.ReadFile(tfc.working_dir + "/http_response_body.log")
		if err == nil {
			apdlogs += string(f_bytes_body)
		} else {
			logrus.Errorf("fail to get the log %s, %s", tfc.working_dir+"/http_response_body.log", err)
		}
		return apdlogs
	} else {
		// job log
		return ""
	}
}

func (tfc TfController) sysCmd(cmdstr string, args ...string) string {
	if out, err := exec.Command(cmdstr, args...).Output(); err != nil {
		return string(out)
	} else {
		logrus.Errorf("fail to run sys cmd: %s", err)
		return err.Error()
	}
}

func (tfc TfController) TfFilePath() string {
	return tfc.file_path
}

func (tfc *TfController) EnvsBase64(envs map[string]string) {
	if tfc.tf != nil {
		decode_envs := map[string]string{}
		for key, env_encode := range envs {
			decode, err := base64.StdEncoding.DecodeString(env_encode)
			if err == nil {
				decode_envs[key] = string(decode)
			} else {
				logrus.Errorf("fail to decode base64 env %s, %v", key, err)
			}
		}
		if len(decode_envs["GOOGLE_CREDENTIALS"]) > 0 && len(decode_envs["GOOGLE_CREDENTIALS_PRIVATEKEY"]) > 0 {
			decode_envs["GOOGLE_PROJECT_APIKEY"] = tfc.genAccessToken(decode_envs["GOOGLE_CREDENTIALS"],
				decode_envs["GOOGLE_CREDENTIALS_PRIVATEKEY"])
			if len(decode_envs["GOOGLE_PROJECT_APIKEY"]) > 0 {
				tfc.vars["gcp_project_apikey"] = decode_envs["GOOGLE_PROJECT_APIKEY"]
			} else {
				logrus.Errorf("can't get the access key to create gcp services: ", fmt.Errorf("fail to retrieve GOOGLE_PROJECT_APIKEY"))
			}
		}
		for idx, e := range decode_envs {
			logrus.Debugf(tfc.working_dir+" DEBUGING checking decode envs %s = %s ,", idx, e)
		}
		tfc.tf.SetEnv(decode_envs)
	} else {
		logrus.Warnf("fail to set envs. Please run InitTerraform to initialize terraform first")
	}
}

func (tfc *TfController) Envs(envs map[string]string) {
	if tfc.tf != nil {
		for idx, e := range envs {
			logrus.Debugf(tfc.working_dir+" DEBUGING checking normal envs %s = %s ,", idx, e)
		}
		tfc.tf.SetEnv(envs)
	} else {
		logrus.Warnf("fail to set envs. Please run InitTerraform to initialize terraform first")
	}
}

func (tfc *TfController) Finished(op string) (bool, error) {
	if tflocks_pool[tfc.file_path+op] == nil {
		logrus.Warnf("new flock for " + tfc.flock_prefix + tfc.file_path + tfc.flock_suffix + op)
		tflocks_pool[tfc.file_path+op] = flock.New(tfc.flock_prefix + tfc.file_path + tfc.flock_suffix + op)
	} else if tflocks_pool[tfc.file_path+op].Locked() {
		// TODO: tflocks_pool thread safe and is RWMutex, not R. TryLock still return true when it is in locked state.
		logrus.Warnf("lock in use: " + tfc.flock_prefix + tfc.file_path + tfc.flock_suffix + op)
		return false, fmt.Errorf(op + ", the lock " + tfc.file_path + " is in use")
	}
	logrus.Warnf(tfc.file_path+op+", before try flock: p %s, path %s, info %v, state %s",
		tflocks_pool[tfc.file_path+op], tflocks_pool[tfc.file_path+op].Path(),
		tflocks_pool[tfc.file_path+op].String(), tflocks_pool[tfc.file_path+op].Locked())
	locked, err := tflocks_pool[tfc.file_path+op].TryLock()
	if err != nil {
		logrus.Warnf(tfc.file_path+op+", fail to get flock: %s", err)
		return false, err
	}
	if locked {
		logrus.Warnf(tfc.file_path+op+", success flock: p %s, path %s, info %v, state %s",
			tflocks_pool[tfc.file_path+op], tflocks_pool[tfc.file_path+op].Path(),
			tflocks_pool[tfc.file_path+op].String(), tflocks_pool[tfc.file_path+op].Locked())
		return true, nil
	} else {
		logrus.Warnf(tfc.file_path + op + ", can't lock " + tfc.file_path)
		return false, fmt.Errorf(op + ", can't lock " + tfc.file_path + ", in use")
	}
}

func (tfc *TfController) MarkAsFinish(op string, delay bool) error {
	if delay {
		if sys_delay[tfc.file_path+op] <= 0 {
			sys_delay[tfc.file_path+op] = tfc.delay_init
		}
		logrus.Warnf("%s, %s, mark as finish, delay %d seconds start", tfc.file_path, op, sys_delay[tfc.file_path+op])
		time.Sleep(time.Duration(sys_delay[tfc.file_path+op]) * time.Second)
		sys_delay[tfc.file_path+op] = sys_delay[tfc.file_path+op] + tfc.delay_delta
		logrus.Warnf("%s, %s, mark as finish, delay %d end", tfc.file_path, op, sys_delay[tfc.file_path+op])
		// TODO: if delayed several time during stress test, it is better to unload all pr
	}
	if tflocks_pool[tfc.file_path+op] != nil {
		logrus.Infof("unlock " + tfc.flock_prefix + tfc.file_path + tfc.flock_suffix + op)
		err := tflocks_pool[tfc.file_path+op].Unlock()
		if err != nil {
			logrus.Warnf(tfc.file_path+op+", fail to unlock: %s", err)
			return err
		}
		logrus.Infof(tfc.file_path+op+", success unlock: path %s, info %v, state %s",
			tflocks_pool[tfc.file_path+op].Path(), tflocks_pool[tfc.file_path+op].String(),
			tflocks_pool[tfc.file_path+op].Locked())
	} else {
		logrus.Warnf(tfc.file_path + op + ", nil flock handler")
		return fmt.Errorf(op + ", nil flock handler for " + tfc.file_path)
	}
	if err := os.Remove(tfc.flock_prefix + tfc.file_path + tfc.flock_suffix + op); err != nil {
		logrus.Warnf("can not remove the lock file: %s, err: %s",
			tfc.flock_prefix+tfc.file_path+tfc.flock_suffix+op, err)
	}
	return nil
}

func (tfc *TfController) TfConfigsExists() bool {
	_, err := os.Stat(tfc.working_dir)
	if err == nil {
		logrus.Infof("runner directory exists." + tfc.working_dir)
		return true
	}
	logrus.Infof(tfc.working_dir+" runner directory dose not exists. info %s", err)
	return false
}

func (tfc *TfController) GenTfConfigs(src_dir string) (err error) {
	src := src_dir + tfc.pr
	if tfc.TfConfigsExists() {
		return nil
	}

	err = cp.Copy(src, tfc.working_dir)
	if err != nil {
		logrus.Errorf("fail to generate terraform configs: %s", err)
	}
	return err
}

func (tfc *TfController) InitTerraform() (err error) {
	if tfc.install {
		installer := &releases.ExactVersion{
			Product: product.Terraform,
			Version: version.Must(version.NewVersion(tfc.tf_version)),
		}
		tfc.exec_path, err = installer.Install(context.Background())
		if err != nil {
			logrus.Errorf(tfc.working_dir+" fail to installing Terraform: %s", err)
		}
	}
	if tfc.pr == "gcp" && tfc.vars["gcp_dind"] == "true" {
		tfc.pr = "gcp_dind"
	}
	if !tfc.TfConfigsExists() {
		logrus.Warnf(tfc.working_dir + " already destroyed. skip it")
		return fmt.Errorf(tfc.working_dir + " dose not exists.")
	}
	tfc.tf, err = tfexec.NewTerraform(tfc.working_dir, tfc.exec_path)
	if err != nil {
		logrus.Errorf(tfc.working_dir+" fail to running NewTerraform: %s", err)
		return err
	}
	tfc.buildEnvs()
	return err
}

func (tfc *TfController) Init() (err error) {
	tfc.preMatchHcl()
	if err != nil {
		logrus.Warnf(tfc.working_dir+" fail to run Plan: %s", err)
	}
	initOps := []tfexec.InitOption{
		tfexec.Upgrade(false),
	}
	err = tfc.tf.Init(context.Background(), initOps...)
	if err != nil {
		logrus.Errorf(tfc.working_dir+" fail to run Init: %s", err)
	}
	tfc.postMatchHcl()
	return err
}

func (tfc *TfController) Plan() (plan_diff bool, err error) {
	planOptions := []tfexec.PlanOption{
		tfexec.Out(tfc.working_dir),
	}
	plan_diff, err = tfc.tf.Plan(context.Background(), planOptions...)
	if err != nil {
		logrus.Errorf(tfc.working_dir+" fail to run Plan: %s", err)
	}
	return plan_diff, err
}

func (tfc *TfController) Apply() (err error) {
	logwriter := logrus.StandardLogger().Writer()
	tfc.tf.SetStdout(logwriter)
	tfc.tf.SetStderr(logwriter)
	defer logwriter.Close()
	tfc.buildEnvs()
	err, vars := tfc.buildVars("apply")
	if err != nil {
		logrus.Errorf("fail to build apply vars " + tfc.working_dir)
		return err
	}
	apply_opt := []tfexec.ApplyOption{
		tfexec.Lock(true),
		tfexec.LockTimeout(tfc.locktimeout),
		tfexec.VarFile(tfc.var_file),
	}
	for _, v := range vars {
		logrus.Debugf(tfc.working_dir + " DEBUGING checking vars start: " + v)
		apply_opt = append(apply_opt, tfexec.Var(v))
	}

	err = tfc.tf.Apply(context.Background(), apply_opt...)
	if err != nil {
		logrus.Errorf(tfc.working_dir+" fail to run Apply: %s, output %s", err, tfc.appendApplyLog())
		return err
	}
	logrus.Infof("success apply resources: %s", tfc.appendApplyLog())
	return nil
}

func (tfc *TfController) Destroy() (err error) {
	destroy_opt := []tfexec.DestroyOption{
		tfexec.Lock(true),
		tfexec.LockTimeout(tfc.locktimeout),
		tfexec.VarFile(tfc.var_file),
	}
	err, vars := tfc.buildVars("destroy")
	if err != nil {
		logrus.Errorf("fail to build destroy vars " + tfc.working_dir)
		return err
	}
	for _, v := range vars {
		destroy_opt = append(destroy_opt, tfexec.Var(v))
	}
	if err = tfc.tf.Destroy(context.Background(), destroy_opt...); err != nil {
		logrus.Errorf(tfc.working_dir+" fail to run Destroy: %s", err)
		return err
	}
	if err = os.RemoveAll(tfc.working_dir); err != nil {
		logrus.Errorf("fail to remove dir "+tfc.working_dir+": %s", err)
		return err
	}
	return nil
}

func (tfc *TfController) CleanTfAndLock() (err error) {
	if err = os.RemoveAll(tfc.working_dir); err != nil {
		logrus.Errorf("CleanTfAndLock fail to remove dir "+tfc.working_dir+": %v", err)
		return err
	}
	if err := tfc.MarkAsFinish("gen", false); err != nil {
		logrus.Errorf("CleanTfAndLock fail to remove gen lock "+tfc.working_dir+": %v", err)
	}
	if err := tfc.MarkAsFinish("del", false); err != nil {
		logrus.Errorf("CleanTfAndLock fail to remove del lock "+tfc.working_dir+": %v", err)
	}
	return nil
}

func (tfc *TfController) FileState(filepath string) (bool, string) {
	f := common.Ternary(filepath == "", tfc.working_dir+"/terraform.tfstate", filepath+"/terraform.tfstate").(string)
	f_bytes_code, err := os.ReadFile(f)
	if err == nil {
		if strings.Contains(string(f_bytes_code), tfc.pr_services[tfc.pr]) {
			msg := f + " service recorded in terraform.tfstate"
			logrus.Warnf(msg)
			return true, msg
		} else {
			msg := f + " service not recorded in terraform.tfstate"
			logrus.Warnf(msg)
			return false, msg
		}
	}
	logrus.Errorf("fail to get the log %s, %s", f+" /http_response_code.log", err)
	return false, err.Error()
}

func (tfc *TfController) State(marker bool) (bool, string) {
	state, err := tfc.tf.Show(context.Background())
	if err != nil {
		// TODO: state exists, but tf plugin dose not response during stress test
		msg := tfc.working_dir + " the tf state not exists."
		logrus.Errorf(msg+" error running Show: %s", err)
		return tfc.FileState("")
	}
	if err := state.Validate(); err != nil {
		msg := tfc.working_dir + " the tf state validate fail"
		logrus.Errorf(msg+": %s", err)
		return false, msg
	}
	if state.Values == nil || state.Values.RootModule == nil {
		msg := tfc.working_dir + " the tf state dose not contains values, please init first"
		logrus.Warnf(msg+": %s", err)
		return false, msg
	} else {
		isRsExists := func(resources []*tfjson.StateResource, k string) bool {
			for _, r := range resources {
				logrus.Infof(tfc.working_dir+", resource address:%s, type: %s, key: %s", r.Address, r.Type, k)
				if r.Type == k || strings.Contains(r.Address, k) {
					logrus.Infof(tfc.working_dir+", state exists: %s", k)
					return true
				}
			}
			return false
		}
		if exist := isRsExists(state.Values.RootModule.Resources, tfc.pr_services[tfc.pr]); exist {
			if marker {
				applyMarker(tfc.working_dir, true)
			}
			return true, "service created"
		}
		if state.Values.RootModule.ChildModules != nil {
			for _, child := range state.Values.RootModule.ChildModules {
				if exist := isRsExists(child.Resources, tfc.pr_services[tfc.pr]); exist {
					if marker {
						applyMarker(tfc.working_dir, true)
					}
					return true, "child service created"
				}
			}
		}
	}
	return false, "service not created"
}
