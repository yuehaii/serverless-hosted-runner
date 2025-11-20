// Package tfcontroller
package tfcontroller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"strings"
	"time"

	"serverless-hosted-runner/agent"
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
	tflocksPool = make(map[string]*flock.Flock)
	sysDelay    = make(map[string]int)
)

type ITfController interface {
	Envs(map[string]string)
	EnvsBase64(map[string]string)
	GenTfConfigs(string) error
	TfConfigsExists() bool
	InitTerraform() error
	Init() error
	Plan() (planDiff bool, err error)
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
	install       bool
	tfVersion     string
	execPath      string
	filePath      string
	workingDir    string
	tf            *tfexec.Terraform
	locktimeout   string
	vars          map[string]string
	varFile       string
	cacheDir      string
	prServices    map[string]string
	pr            string
	prebuildCache bool
	matchHclConf  bool
	store         common.Store
	hclFilepath   string
	flockPrefix   string
	flockSuffix   string
	delayInit     int
	delayDelta    int
	disDir        string
	gitAgent      agent.IGit
}

func CreateController(workingDir string, vars map[string]string, store common.Store, pr string,
	dynamicLabels func([]string, *string, *string, *string, *string, *string, *string) string, labels []string, gitAgent agent.IGit) ITfController {
	key, runnerType := store.GetKey()
	pat, patType := store.GetPat()
	specificLabels, _ := store.GetLabels()
	cloudPr, _ := store.GetCloudProvider()
	regURL, _ := store.GetURL()
	cpu, _ := store.GetCPU()
	mem, _ := store.GetMemory()
	secGpID, _ := store.GetSecGpID()
	vswitchID, _ := store.GetVSwitchID()
	gcpDind, _ := store.GetGcpDind()
	repoImageVer, _ := store.GetImageVersion()
	imageVer := common.Ternary(repoImageVer == "", vars["image_ver"], repoImageVer).(string)
	largeDisk := ""
	specificLabels += dynamicLabels(labels, &cpu, &mem, &vswitchID, &secGpID, &imageVer, &largeDisk)
	vars["image_ver"] = imageVer //  label > repo > default dispacher ver
	logrus.Infof("patType: %s, runnerType: %s", patType, runnerType)
	if patType != runnerType && (patType == "repo" && runnerType == "org") {
		runnerType = patType
		regURL = regURL + "/" + vars["repo_name"]
	}
	varsAppend := map[string]string{
		"key": key, "specific_labels": specificLabels, "runner_type": runnerType,
		"pat": pat, "pat_type": patType, "reg_url": regURL, "cpu": cpu, "mem": mem,
		"gcp_dind": gcpDind, "vswitch_id": vswitchID, "sec_gp_id": secGpID,
		"large_disk": largeDisk,
	}
	maps.Copy(vars, varsAppend)
	workingPath := vars["repo_name"] + "-" + vars["runer_id"]
	if vars["runner_type"] == "org" {
		workingPath = vars["org_name"] + "-" + vars["runer_id"]
	}
	if applyMarker(workingDir+workingPath, false) {
		vars["runer_id"] += "-sls-comp"
		workingPath += "-sls-comp"
	}
	return tfController(workingPath, workingDir+workingPath, vars, store,
		common.Ternary(len(strings.TrimSpace(cloudPr)) > 0, strings.TrimSpace(cloudPr), strings.TrimSpace(pr)).(string),
		workingDir, gitAgent)
}

func DestroyController(filePath, workingDir string, store common.Store, pr string, disDir string, gitAgent agent.IGit) ITfController {
	gcpDind, _ := store.GetGcpDind()
	key, _ := store.GetKey()
	cloudPr, _ := store.GetCloudProvider()
	if len(key) <= 0 {
		logrus.Warnf("org and repo token not exist. please run the 'register serverless runner' workflow first.")
		return nil
	}
	vars := map[string]string{
		"key": key, "gcp_dind": gcpDind,
	}
	return tfController(filePath, workingDir, vars, store,
		common.Ternary(len(strings.TrimSpace(cloudPr)) > 0, strings.TrimSpace(cloudPr), strings.TrimSpace(pr)).(string),
		disDir, gitAgent)
}

func tfController(filePath, workingDir string, vars map[string]string, store common.Store, pr string, disDir string, gitAgent agent.IGit) ITfController {
	return &TfController{false, "1.10.5", "/usr/local/bin/terraform", filePath, workingDir, nil,
		"20m", vars, "ubuntu_runner.tfvars", disDir + "tf_plugin_cache",
		map[string]string{"ali": "alicloud_eci_container_group", "azure": "azurerm_container_group",
			"gcp": "google_cloud_run", "gcp_dind": "gcp_runner_batch_job_module"}, pr, false, false,
		store, "/.terraform.lock.hcl", "/var/lock/sls-", ".lock.", 300, 60, disDir, gitAgent}
}

func applyMarker(workingDir string, init bool) bool {
	_, err := os.Stat(workingDir)
	if err != nil {
		logrus.Errorf("apply marker, runner directory dose not exists: %s", workingDir)
		return false
	}
	markerFile := workingDir + "/.marker"
	if init {
		markerFile, err := os.Create(markerFile)
		if err == nil {
			logrus.Warnf("marker initialized for %s", workingDir)
			markerClose := func() {
				if err := markerFile.Close(); err != nil {
					logrus.Errorf("fail to close marker %v", err)
				}
			}
			defer markerClose()
		} else {
			logrus.Errorf("fail to create apply marker for %s", workingDir)
		}
		return false
	} else {
		fDes, err := os.Stat(markerFile)
		if err != nil {
			logrus.Errorf("marker file dose not exists: %s", workingDir)
			return false
		} else {
			curTime := time.Now()
			fTime := fDes.ModTime()
			logrus.Warnf("marker %s, curTime %v, fTime %v, sub %v mins", workingDir, curTime, fTime, curTime.Sub(fTime).Minutes())
			return curTime.Sub(fTime).Minutes() > 5
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
	if tk, _ := jctl.ExchangeAPIKey(); len(tk) == 0 {
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
		srcFile, err := os.Open(src)
		if err == nil {
			destFile, err := os.Create(dest)
			if err == nil {
				fileClose := func() {
					if err := destFile.Close(); err != nil {
						logrus.Errorf("cpTfConfig, fail to close dest file: %s", err)
					}
					if err := srcFile.Close(); err != nil {
						logrus.Errorf("cpTfConfig, fail to close src file: %s", err)
					}
				}
				defer fileClose()
				lens, err := io.Copy(destFile, srcFile)
				if err == nil {
					logrus.Infof("success cp %v size file from %s to %s", lens, src, dest)
				}
			}
		}
	}
	return err
}

func (tfc *TfController) setupCache() {
	if len(tfc.cacheDir) > 0 {
		if err := tfc.tf.SetEnv(map[string]string{
			"TF_PLUGIN_cacheDir": tfc.cacheDir,
			// it will cause unmatch issue in gcp pr
			"TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE": common.Ternary(tfc.pr == "ali" || tfc.pr == "azure", "true", "").(string),
		}); err != nil {
			logrus.Errorf("fail to set tf env: %s", err)
		}
		_, err := os.Stat(tfc.cacheDir)
		if err != nil {
			if err := os.MkdirAll(tfc.cacheDir, os.ModePerm); err != nil {
				logrus.Errorf("fail to create cache dir: %s", err)
			}
		} else {
			logrus.Infof("cache dir already exists.")
		}
	}
}

func (tfc *TfController) preBuildHcl() {
	cacheHclPath := tfc.cacheDir + tfc.hclFilepath
	initOps := []tfexec.InitOption{
		tfexec.Upgrade(false),
	}

	runnerDir := tfc.disDir + "runner/" + tfc.pr
	tf, err := tfexec.NewTerraform(runnerDir, tfc.execPath)
	if err != nil {
		logrus.Errorf(runnerDir+" fail to new tf to create cache: %s", err)
		return
	}
	err = tf.Init(context.Background(), initOps...)
	if err != nil {
		logrus.Errorf(tfc.cacheDir+" fail to run init to create cache: %s", err)
	}
	err = tfc.cpTfConfig(runnerDir+tfc.hclFilepath, cacheHclPath)
	if err != nil {
		logrus.Errorf(cacheHclPath+" can't be setup match. %s", err)
	}
	if err = os.Remove(runnerDir + tfc.hclFilepath); err != nil {
		logrus.Errorf(runnerDir+" fail to remove hcl file path: %v", err)
	}
	if err = os.RemoveAll(runnerDir + "/.terraform"); err != nil {
		logrus.Errorf(runnerDir+" fail to remove tf dir: %v", err)
	}
}

func (tfc TfController) CleanHCL() error {
	return os.Remove(tfc.workingDir + tfc.hclFilepath)
}

func (tfc *TfController) preMatchHcl() {
	tfc.setupCache()
	if tfc.prebuildCache {
		tfc.preBuildHcl()
	}
	if tfc.matchHclConf {
		tfc.matchHcl()
	}
}

func (tfc *TfController) postMatchHcl() {
	if tfc.matchHclConf {
		tfc.matchHcl()
	}
}

// match hcl can reduce the pr version not match caused delay
func (tfc *TfController) matchHcl() {
	runnerHclPath := tfc.workingDir + tfc.hclFilepath
	cacheHclPath := tfc.cacheDir + tfc.hclFilepath
	_, errCacheLock := os.Stat(cacheHclPath)
	_, errRunnerLock := os.Stat(runnerHclPath)
	if errCacheLock == nil && errRunnerLock != nil {
		logrus.Infof("match hcl from cache to runner %s", tfc.workingDir)
		err := tfc.cpTfConfig(cacheHclPath, runnerHclPath)
		if err != nil {
			logrus.Errorf("fail to match hcl from cache to runner: %s", err)
		}
	} else if errCacheLock != nil && errRunnerLock == nil {
		logrus.Infof("match hcl from runner to cache %s", tfc.workingDir)
		err := tfc.cpTfConfig(runnerHclPath, cacheHclPath)
		if err != nil {
			logrus.Errorf("fail to init cache hcl from runner to cache: %s", err)
		}
	} else if errCacheLock != nil && errRunnerLock != nil {
		logrus.Infof("%s, %s, hcl dose not exists for cache and runner. welcome first runner."+tfc.workingDir,
			errCacheLock, errRunnerLock)
	}

}

func (tfc *TfController) buildVars(op string) ([]string, error) {
	switch op {
	case "init":
		return []string{}, nil
	case "apply":
		return tfc.buildApplyVars()
	case "destroy":
		return tfc.buildDestroyVars()
	default:
		return []string{}, fmt.Errorf("invalid build var operation, should be init/apply/destroy")
	}
}

func (tfc *TfController) buildDestroyVars() ([]string, error) {
	vars := []string{}
	if tfc.pr == "azure" {
		armSubnetID, _ := tfc.store.GetArmSubnetID()
		armLogAnaWorkspaceID, _ := tfc.store.GetArmLogAnalyticsWorkspaceID()
		armLogAnaWorkspaceKey, _ := tfc.store.GetArmLogAnalyticsWorkspaceKey()
		aciLocation, _ := tfc.store.GetAciLocation()
		vars = []string{"subnet_ids=" + armSubnetID, "workspace_id=" + armLogAnaWorkspaceID,
			"workspace_key=" + armLogAnaWorkspaceKey, "aci_location=" + aciLocation}
	}
	return vars, nil
}

func (tfc *TfController) buildRegToken(regTokenStr string) string {
	marshalFreshToken := func(newToken agent.GitRegToken) string {
		logrus.Warnf("marshalFreshToken reg token refreshed, new %s, %s", newToken.Token, newToken.Exp)
		if len(newToken.Token) > 0 {
			if bToken, err := json.Marshal(newToken); err == nil && len(string(bToken)) > 0 {
				regTokenStr = string(bToken)
				logrus.Warnf("marshalFreshToken updated new token, %s", regTokenStr)
				tfc.store.UpdateRepoRegToken(regTokenStr)
				return regTokenStr
			} else {
				logrus.Errorf("marshalFreshToken fail to marshalFreshToken newToken, %v", err)
			}
		} else {
			logrus.Errorf("marshalFreshToken newToken is empty")
		}
		return "none"
	}
	if len(regTokenStr) > 0 {
		regToken := agent.GitRegToken{}
		if err := json.Unmarshal([]byte(regTokenStr), &regToken); err != nil {
			logrus.Errorf("buildRunnerRegToken, fail to unmarshal regTokenStr: %s, %v", regTokenStr, err)
		} else {
			logrus.Warnf("buildRunnerRegToken, reg token check, old %s, %s", regToken.Token, regToken.Exp)
			if newToken, valid := tfc.gitAgent.GetRegistrationToken(regToken); !valid {
				return marshalFreshToken(newToken)
			} else {
				return marshalFreshToken(regToken)
			}
		}
	} else {
		tk, _ := tfc.gitAgent.GetRegistrationToken(
			agent.GitRegToken{IsOrg: common.Ternary(strings.ToLower(tfc.vars["runner_type"]) == "repo", false, true).(bool),
				Repo: tfc.vars["repo_name"], URL: strings.ReplaceAll(tfc.vars["reg_url"], "/"+tfc.vars["repo_name"], ""),
				Token: "", Exp: ""})
		return marshalFreshToken(tk)
	}
	return "none"
}

func (tfc *TfController) buildApplyVars() ([]string, error) {
	runnerOrg := "none"
	chargeLabels, _ := tfc.store.GetChargeLabels()
	runnerGroup, _ := tfc.store.GetRunnerGroup()
	armResourceGroupName, _ := tfc.store.GetArmResourceGroupName()
	armSubnetID, _ := tfc.store.GetArmSubnetID()
	armLogAnaWorkspaceID, _ := tfc.store.GetArmLogAnalyticsWorkspaceID()
	armLogAnaWorkspaceKey, _ := tfc.store.GetArmLogAnalyticsWorkspaceKey()
	gcpProject, _ := tfc.store.GetGcpProject()
	gcpRegion, _ := tfc.store.GetGcpRegion()
	gcpSa, _ := tfc.store.GetGcpSA()
	gcpVpc, _ := tfc.store.GetGcpVpc()
	gcpSubnet, _ := tfc.store.GetGcpSubnet()
	aciLocation, _ := tfc.store.GetAciLocation()
	aciSku, _ := tfc.store.GetAciSku()
	aciNetworkType, _ := tfc.store.GetAciNetworkType()
	if tfc.vars["runner_type"] == "org" {
		runnerOrg = tfc.vars["org_name"]
	}
	regTokenStr, _ := tfc.store.GetRepoRegToken()
	regTokenStr = tfc.buildRegToken(regTokenStr)

	vars := []string{"runner_id=" + tfc.vars["runer_id"], "runner_repname=" + tfc.vars["repo_name"], "runner_orgowner=" + tfc.vars["owner_name"],
		"runner_action=" + tfc.vars["act"], "runner_repurl=" + tfc.vars["reg_url"], "runner_token=" + tfc.vars["pat"], "image_ver=" + tfc.vars["image_ver"],
		"security_group_id=" + tfc.vars["sec_gp_id"], "vswitch_id=" + tfc.vars["vswitch_id"], "container_type=" + tfc.vars["runner_type"],
		"runner_orgname=" + runnerOrg, "network_mode=fixed", "runner_cpu=" + tfc.vars["cpu"], "runner_memory=" + tfc.vars["mem"],
		"runner_labels=" + common.Ternary(tfc.vars["specific_labels"] == "", "none", strings.ReplaceAll(tfc.vars["specific_labels"], " ", "")).(string),
		"charge_labels=" + common.Ternary(chargeLabels == "", "none", strings.ReplaceAll(chargeLabels, " ", "")).(string),
		"runner_group=" + common.Ternary(runnerGroup == "", "default", strings.ReplaceAll(runnerGroup, " ", "")).(string),
		"oss_mount=" + tfc.vars["large_disk"], "cloud_pr=" + tfc.pr, "dis_ip=" + tfc.vars["dis_ip"],
		"ctx_log_level=" + tfc.vars["ctx_log_level"], "subnet_ids=" + armSubnetID, "resource_group_name=" + armResourceGroupName,
		"workspace_id=" + armLogAnaWorkspaceID, "workspace_key=" + armLogAnaWorkspaceKey,
		"aci_location=" + aciLocation, "aci_sku=" + aciSku, "aci_network_type=" + aciNetworkType,
		"gcp_project=" + gcpProject, "gcp_region=" + gcpRegion, "gcp_project_sa_email=" + gcpSa, "gcp_project_apikey=" + tfc.vars["gcp_project_apikey"],
		"gcp_runner_dind=" + tfc.vars["gcp_dind"], "gcp_vpc=" + gcpVpc, "gcp_subnet=" + gcpSubnet, "repo_reg_tk=" + regTokenStr,
		"IMAGE_RETRIEVE_SERVER=" + common.Ternary(tfc.pr == "azure", os.Getenv("TF_VAR_AZURE_ACR_SERVER"), os.Getenv("TF_VAR_IMAGE_RETRIEVE_SERVER")).(string),
		"IMAGE_RETRIEVE_USERNAME=" + common.Ternary(tfc.pr == "azure", os.Getenv("TF_VAR_AZURE_ACR_USRNAME"), os.Getenv("TF_VAR_IMAGE_RETRIEVE_USERNAME")).(string),
		"IMAGE_RETRIEVE_PWD=" + common.Ternary(tfc.pr == "azure", os.Getenv("TF_VAR_AZURE_ACR_PWD"), os.Getenv("TF_VAR_IMAGE_RETRIEVE_PWD")).(string),
		"CTX_USERNAME=" + os.Getenv("TF_VAR_CTX_USERNAME"), "CTX_PWD=" + os.Getenv("TF_VAR_CTX_PWD"), "SLS_ENC_KEY=" + os.Getenv("TF_VAR_SLS_ENC_KEY"),
	}
	return vars, nil
}

func (tfc *TfController) buildEnvs() {
	sec, _ := tfc.store.GetSecret()
	region, _ := tfc.store.GetRegion()
	armClientID, _ := tfc.store.GetArmClientID()
	armClientSecret, _ := tfc.store.GetArmClientSecret()
	armSubscriptionID, _ := tfc.store.GetArmSubscriptionID()
	armTenantID, _ := tfc.store.GetArmTenantID()
	armEnvironment, _ := tfc.store.GetArmEnvironment()
	armRpRegistration, _ := tfc.store.GetArmRPRegistration()
	gcpCredentials, _ := tfc.store.GetGcpCredential()
	gcpProject, _ := tfc.store.GetGcpProject()
	gcpRegion, _ := tfc.store.GetGcpRegion()
	gcpSa, _ := tfc.store.GetGcpSA()
	gcpApikey, _ := tfc.store.GetGcpApikey()

	switch tfc.pr {
	case "gcp", "gcp_dind":
		tfc.Envs(map[string]string{
			"GOOGLE_PROJECT":          gcpProject,
			"GOOGLE_REGION":           gcpRegion,
			"GOOGLE_PROJECT_SA_EMAIL": gcpSa,
			"GOOGLE_RUNNER_DIND":      tfc.vars["gcp_dind"],
		})
		tfc.EnvsBase64(map[string]string{
			"GOOGLE_CREDENTIALS":            gcpCredentials,
			"GOOGLE_CREDENTIALS_PRIVATEKEY": gcpApikey,
		})
	case "azure":
		tfc.Envs(map[string]string{
			"ARM_CLIENT_ID":                       armClientID,
			"ARM_CLIENT_SECRET":                   armClientSecret,
			"ARM_SUBSCRIPTION_ID":                 armSubscriptionID,
			"ARM_TENANT_ID":                       armTenantID,
			"ARM_ENVIRONMENT":                     armEnvironment,
			"ARM_RESOURCE_PROVIDER_REGISTRATIONS": armRpRegistration,
		})
	case "ali":
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
		fBytesCode, err := os.ReadFile(tfc.workingDir + "/http_response_code.log")
		if err == nil {
			apdlogs += string(fBytesCode)
		} else {
			logrus.Errorf("fail to get the log %s, %s", tfc.workingDir+"/http_response_code.log", err)
		}

		fBytesBody, err := os.ReadFile(tfc.workingDir + "/http_response_body.log")
		if err == nil {
			apdlogs += string(fBytesBody)
		} else {
			logrus.Errorf("fail to get the log %s, %s", tfc.workingDir+"/http_response_body.log", err)
		}
		return apdlogs
	} else {
		// job log
		return ""
	}
}

func (tfc TfController) TfFilePath() string {
	return tfc.filePath
}

func (tfc *TfController) EnvsBase64(envs map[string]string) {
	if tfc.tf != nil {
		decodeEnvs := map[string]string{}
		for key, envEncode := range envs {
			decode, err := base64.StdEncoding.DecodeString(envEncode)
			if err == nil {
				decodeEnvs[key] = string(decode)
			} else {
				logrus.Errorf("fail to decode base64 env %s, %v", key, err)
			}
		}
		if len(decodeEnvs["GOOGLE_CREDENTIALS"]) > 0 && len(decodeEnvs["GOOGLE_CREDENTIALS_PRIVATEKEY"]) > 0 {
			decodeEnvs["GOOGLE_PROJECT_APIKEY"] = tfc.genAccessToken(decodeEnvs["GOOGLE_CREDENTIALS"],
				decodeEnvs["GOOGLE_CREDENTIALS_PRIVATEKEY"])
			if len(decodeEnvs["GOOGLE_PROJECT_APIKEY"]) > 0 {
				tfc.vars["gcp_project_apikey"] = decodeEnvs["GOOGLE_PROJECT_APIKEY"]
			} else {
				logrus.Errorf("can't get the access key to create gcp services: %v", fmt.Errorf("fail to retrieve GOOGLE_PROJECT_APIKEY"))
			}
		}
		// for idx, e := range decodeEnvs {
		// 	logrus.Debugf(tfc.workingDir+" DEBUGING checking decode envs %s = %s ,", idx, e)
		// }
		if err := tfc.tf.SetEnv(decodeEnvs); err != nil {
			logrus.Errorf("env base64, fail to set env, %v", err)
		}
	} else {
		logrus.Warnf("fail to set envs. Please run InitTerraform to initialize terraform first")
	}
}

func (tfc *TfController) Envs(envs map[string]string) {
	if tfc.tf != nil {
		// for idx, e := range envs {
		// 	logrus.Debugf(tfc.workingDir+" DEBUGING checking normal envs %s = %s ,", idx, e)
		// }
		if err := tfc.tf.SetEnv(envs); err != nil {
			logrus.Errorf("fail to set env, %v", err)
		}
	} else {
		logrus.Warnf("fail to set envs. Please run InitTerraform to initialize terraform first")
	}
}

func (tfc *TfController) Finished(op string) (bool, error) {
	if tflocksPool[tfc.filePath+op] == nil {
		logrus.Warnf("new flock for %s", tfc.flockPrefix+tfc.filePath+tfc.flockSuffix+op)
		tflocksPool[tfc.filePath+op] = flock.New(tfc.flockPrefix + tfc.filePath + tfc.flockSuffix + op)
	} else if tflocksPool[tfc.filePath+op].Locked() {
		// TODO: tflocksPool thread safe and is RWMutex, not R. TryLock still return true when it is in locked state.
		logrus.Warnf("lock in use: %s", tfc.flockPrefix+tfc.filePath+tfc.flockSuffix+op)
		return false, fmt.Errorf("%s, the lock %s is in use", op, tfc.filePath)
	}
	logrus.Warnf(tfc.filePath+op+", before try flock: p %s, path %s, info %v, state %s",
		tflocksPool[tfc.filePath+op], tflocksPool[tfc.filePath+op].Path(),
		tflocksPool[tfc.filePath+op].String(), tflocksPool[tfc.filePath+op].Locked())
	locked, err := tflocksPool[tfc.filePath+op].TryLock()
	if err != nil {
		logrus.Warnf(tfc.filePath+op+", fail to get flock: %s", err)
		return false, err
	}
	if locked {
		logrus.Warnf(tfc.filePath+op+", success flock: p %s, path %s, info %v, state %s",
			tflocksPool[tfc.filePath+op], tflocksPool[tfc.filePath+op].Path(),
			tflocksPool[tfc.filePath+op].String(), tflocksPool[tfc.filePath+op].Locked())
		return true, nil
	} else {
		logrus.Warnf("%s %s, can't lock %s", tfc.filePath, op, tfc.filePath)
		return false, fmt.Errorf("%s, can't lock %s, in use", op, tfc.filePath)
	}
}

func (tfc *TfController) MarkAsFinish(op string, delay bool) error {
	if delay {
		if sysDelay[tfc.filePath+op] <= 0 {
			sysDelay[tfc.filePath+op] = tfc.delayInit
		}
		logrus.Warnf("%s, %s, mark as finish, delay %d seconds start", tfc.filePath, op, sysDelay[tfc.filePath+op])
		time.Sleep(time.Duration(sysDelay[tfc.filePath+op]) * time.Second)
		sysDelay[tfc.filePath+op] = sysDelay[tfc.filePath+op] + tfc.delayDelta
		logrus.Warnf("%s, %s, mark as finish, delay %d end", tfc.filePath, op, sysDelay[tfc.filePath+op])
		// TODO: if delayed several time during stress test, it is better to unload all pr
	}
	if tflocksPool[tfc.filePath+op] != nil {
		logrus.Infof("unlock %s", tfc.flockPrefix+tfc.filePath+tfc.flockSuffix+op)
		err := tflocksPool[tfc.filePath+op].Unlock()
		if err != nil {
			logrus.Warnf(tfc.filePath+op+", fail to unlock: %s", err)
			return err
		}
		logrus.Infof(tfc.filePath+op+", success unlock: path %s, info %v, state %s",
			tflocksPool[tfc.filePath+op].Path(), tflocksPool[tfc.filePath+op].String(),
			tflocksPool[tfc.filePath+op].Locked())
	} else {
		logrus.Warn(tfc.filePath + op + ", nil flock handler")
		return fmt.Errorf("%s, nil flock handler for %s", op, tfc.filePath)
	}
	if err := os.Remove(tfc.flockPrefix + tfc.filePath + tfc.flockSuffix + op); err != nil {
		logrus.Warnf("can not remove the lock file: %s, err: %s",
			tfc.flockPrefix+tfc.filePath+tfc.flockSuffix+op, err)
	}
	return nil
}

func (tfc *TfController) TfConfigsExists() bool {
	_, err := os.Stat(tfc.workingDir)
	if err == nil {
		logrus.Infof("runner directory exists %s", tfc.workingDir)
		return true
	}
	logrus.Infof(tfc.workingDir+" runner directory dose not exists. info %s", err)
	return false
}

func (tfc *TfController) GenTfConfigs(srcDir string) (err error) {
	src := srcDir + tfc.pr
	if tfc.TfConfigsExists() {
		return nil
	}

	err = cp.Copy(src, tfc.workingDir)
	if err != nil {
		logrus.Errorf("fail to generate terraform configs: %s", err)
	}
	return err
}

func (tfc *TfController) InitTerraform() (err error) {
	if tfc.install {
		installer := &releases.ExactVersion{
			Product: product.Terraform,
			Version: version.Must(version.NewVersion(tfc.tfVersion)),
		}
		tfc.execPath, err = installer.Install(context.Background())
		if err != nil {
			logrus.Errorf(tfc.workingDir+" fail to installing Terraform: %s", err)
		}
	}
	if tfc.pr == "gcp" && tfc.vars["gcp_dind"] == "true" {
		tfc.pr = "gcp_dind"
	}
	if !tfc.TfConfigsExists() {
		logrus.Warnf("%s already destroyed. skip it", tfc.workingDir)
		return fmt.Errorf("%s dose not exists", tfc.workingDir)
	}
	tfc.tf, err = tfexec.NewTerraform(tfc.workingDir, tfc.execPath)
	if err != nil {
		logrus.Errorf(tfc.workingDir+" fail to running NewTerraform: %s", err)
		return err
	}
	tfc.buildEnvs()
	return err
}

func (tfc *TfController) Init() (err error) {
	tfc.preMatchHcl()
	if err != nil {
		logrus.Warnf(tfc.workingDir+" fail to run Plan: %s", err)
	}
	initOps := []tfexec.InitOption{
		tfexec.Upgrade(false),
	}
	err = tfc.tf.Init(context.Background(), initOps...)
	if err != nil {
		logrus.Errorf(tfc.workingDir+" fail to run Init: %s", err)
	}
	tfc.postMatchHcl()
	return err
}

func (tfc *TfController) Plan() (planDiff bool, err error) {
	planOptions := []tfexec.PlanOption{
		tfexec.Out(tfc.workingDir),
	}
	planDiff, err = tfc.tf.Plan(context.Background(), planOptions...)
	if err != nil {
		logrus.Errorf(tfc.workingDir+" fail to run Plan: %s", err)
	}
	return planDiff, err
}

func (tfc *TfController) Apply() (err error) {
	logwriter := logrus.StandardLogger().Writer()
	tfc.tf.SetStdout(logwriter)
	tfc.tf.SetStderr(logwriter)
	writerClose := func() {
		if err := logwriter.Close(); err != nil {
			logrus.Errorf("apply, fail to close log writer, %v ", err)
		}
	}
	defer writerClose()
	tfc.buildEnvs()
	vars, err := tfc.buildVars("apply")
	if err != nil {
		logrus.Errorf("fail to build apply vars %s", tfc.workingDir)
		return err
	}
	applyOpt := []tfexec.ApplyOption{
		tfexec.Lock(true),
		tfexec.LockTimeout(tfc.locktimeout),
		tfexec.VarFile(tfc.varFile),
	}
	for _, v := range vars {
		// logrus.Debugf(tfc.workingDir + " DEBUGING checking vars start: " + v)
		applyOpt = append(applyOpt, tfexec.Var(v))
	}

	err = tfc.tf.Apply(context.Background(), applyOpt...)
	if err != nil {
		logrus.Errorf(tfc.workingDir+" fail to run Apply: %s, output %s", err, tfc.appendApplyLog())
		return err
	}
	logrus.Infof(tfc.workingDir+"success apply resources: %s", tfc.appendApplyLog())
	return nil
}

func (tfc *TfController) Destroy() (err error) {
	destroyOpt := []tfexec.DestroyOption{
		tfexec.Lock(true),
		tfexec.LockTimeout(tfc.locktimeout),
		tfexec.VarFile(tfc.varFile),
	}
	vars, err := tfc.buildVars("destroy")
	if err != nil {
		logrus.Errorf("fail to build destroy vars %s", tfc.workingDir)
		return err
	}
	for _, v := range vars {
		destroyOpt = append(destroyOpt, tfexec.Var(v))
	}
	if err = tfc.tf.Destroy(context.Background(), destroyOpt...); err != nil {
		logrus.Errorf(tfc.workingDir+" fail to run Destroy: %s", err)
		return err
	}
	if err = os.RemoveAll(tfc.workingDir); err != nil {
		logrus.Errorf("fail to remove dir "+tfc.workingDir+": %s", err)
		return err
	}
	return nil
}

func (tfc *TfController) CleanTfAndLock() (err error) {
	if err = os.RemoveAll(tfc.workingDir); err != nil {
		logrus.Errorf("CleanTfAndLock fail to remove dir "+tfc.workingDir+": %v", err)
		return err
	}
	if err := tfc.MarkAsFinish("gen", false); err != nil {
		logrus.Errorf("CleanTfAndLock fail to remove gen lock "+tfc.workingDir+": %v", err)
	}
	if err := tfc.MarkAsFinish("del", false); err != nil {
		logrus.Errorf("CleanTfAndLock fail to remove del lock "+tfc.workingDir+": %v", err)
	}
	return nil
}

func (tfc *TfController) FileState(filepath string) (bool, string) {
	f := common.Ternary(filepath == "", tfc.workingDir+"/terraform.tfstate", filepath+"/terraform.tfstate").(string)
	fBytesCode, err := os.ReadFile(f)
	if err == nil {
		if strings.Contains(string(fBytesCode), tfc.prServices[tfc.pr]) {
			msg := f + " service recorded in terraform.tfstate"
			logrus.Warn(msg)
			return true, msg
		} else {
			msg := f + " service not recorded in terraform.tfstate"
			logrus.Warn(msg)
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
		msg := tfc.workingDir + " the tf state not exists."
		logrus.Errorf(msg+" error running Show: %s", err)
		return tfc.FileState("")
	}
	if err := state.Validate(); err != nil {
		msg := tfc.workingDir + " the tf state validate fail"
		logrus.Errorf(msg+": %s", err)
		return false, msg
	}
	if state.Values == nil || state.Values.RootModule == nil {
		msg := tfc.workingDir + " the tf state dose not contains values, please init first"
		logrus.Warnf(msg+": %s", err)
		return false, msg
	} else {
		isRsExists := func(resources []*tfjson.StateResource, k string) bool {
			for _, r := range resources {
				logrus.Infof(tfc.workingDir+", resource address:%s, type: %s, key: %s", r.Address, r.Type, k)
				if r.Type == k || strings.Contains(r.Address, k) {
					logrus.Infof(tfc.workingDir+", state exists: %s", k)
					return true
				}
			}
			return false
		}
		if exist := isRsExists(state.Values.RootModule.Resources, tfc.prServices[tfc.pr]); exist {
			if marker {
				applyMarker(tfc.workingDir, true)
			}
			return true, "service created"
		}
		if state.Values.RootModule.ChildModules != nil {
			for _, child := range state.Values.RootModule.ChildModules {
				if exist := isRsExists(child.Resources, tfc.prServices[tfc.pr]); exist {
					if marker {
						applyMarker(tfc.workingDir, true)
					}
					return true, "child service created"
				}
			}
		}
	}
	return false, "service not created"
}
