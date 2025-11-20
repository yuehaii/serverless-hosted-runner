package common

import (
	"os"
	"strings"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

var RegTokenReuse = map[string]string{}

type Store interface {
	Save()
	GetKey() (string, string)
	GetSecret() (string, string)
	GetRegion() (string, string)
	GetSecGpID() (string, string)
	GetVSwitchID() (string, string)
	GetCloudProvider() (string, string)
	GetPat() (string, string)
	GetURL() (string, string)
	GetSize() (string, string)
	GetCPU() (string, string)
	GetMemory() (string, string)
	GetLabels() (string, string)
	GetChargeLabels() (string, string)
	GetRunnerGroup() (string, string)
	GetPreSize() string
	AnyChange() bool
	GetPreKey() string
	GetPreSecret() string
	GetPreRegion() string
	GetPreSecGpID() string
	GetPreVSwitchID() string
	GetPreCloudProvider() string
	GetPrePat() string
	GetPreURL() string
	GetAPIEntTk() string
	GetAPIGitTk() string
	GetGcpCredential() (string, string)
	GetRepoRegToken() (string, string)
	GetGcpProject() (string, string)
	GetGcpRegion() (string, string)
	GetGcpSA() (string, string)
	GetGcpApikey() (string, string)
	GetGcpDind() (string, string)
	GetGcpVpc() (string, string)
	GetGcpSubnet() (string, string)
	GetImageVersion() (string, string)
	GetArmClientID() (string, string)
	GetAciLocation() (string, string)
	GetAciSku() (string, string)
	GetAciNetworkType() (string, string)
	GetArmClientSecret() (string, string)
	GetArmSubscriptionID() (string, string)
	GetArmTenantID() (string, string)
	GetArmEnvironment() (string, string)
	GetArmRPRegistration() (string, string)
	GetArmResourceGroupName() (string, string)
	GetArmSubnetID() (string, string)
	GetArmLogAnalyticsWorkspaceID() (string, string)
	GetArmLogAnalyticsWorkspaceKey() (string, string)
	GetPreGcpCredential() string
	GetPreRepoRegToken() string
	GetPreGcpProject() string
	GetPreGcpRegion() string
	GetPreGcpSA() string
	GetPreGcpApikey() string
	GetPreGcpDind() string
	GetPreGcpVpc() string
	GetPreGcpSubnet() string
	GetPreImageVersion() string
	GetPreArmClientID() string
	GetPreAciLocation() string
	GetPreAciSku() string
	GetPreAciNetworkType() string
	GetPreArmClientSecret() string
	GetPreArmSubscriptionID() string
	GetPreArmTenantID() string
	GetPreArmEnvironment() string
	GetPreArmRPRegistration() string
	GetPreArmResourceGroupName() string
	GetPreArmSubnetID() string
	IsDestory(string) bool
	MarkDestory(string)
	ResetDestory(string)
	UpdateRepoRegToken(string)
}

func EnvStore(msg *PoolMsg, org string, repo string) Store {
	return MsgStore{msg, org, repo, "SLS_GITHUB_TK", "SLS_GITENT_TK",
		"github.com", "git.build.ingka.ikea.com", "SLS_ENC_KEY"}
}

func RedisStore() Store {
	return nil
}

type MsgStore struct {
	msg     *PoolMsg
	org     string
	repo    string
	gittk   string
	entk    string
	gitfqdn string
	enfqdn  string
	enckey  string
}

func (store MsgStore) IsDestory(workflow string) bool {
	return os.Getenv(workflow) == workflow
}

func (store MsgStore) MarkDestory(workflow string) {
	if err := os.Setenv(workflow, workflow); err != nil {
		logrus.Errorf("MarkDestory Setenv failure: %v", err)
	}
}

func (store MsgStore) ResetDestory(workflow string) {
	if err := os.Setenv(workflow, ""); err != nil {
		logrus.Errorf("ResetDestory Setenv failure: %v", err)
	}
}

func (store MsgStore) get(key string) string {
	return os.Getenv(key)
}
func (store MsgStore) set(key string, val string) {
	if err := os.Setenv(key, val); err != nil {
		logrus.Errorf("set Setenv failure: %v", err)
	}
}
func (store MsgStore) msgKey(k string) string {
	prefix := store.msg.Type + "_" + store.msg.Name + "_"
	return prefix + k
}
func (store MsgStore) repoKey(k string) string {
	prefix := "Repo_" + store.repo + "_"
	return prefix + k
}
func (store MsgStore) orgKey(k string) string {
	prefix := "Org_" + store.org + "_"
	return prefix + k
}
func (store MsgStore) poolKey(k string) string {
	prefix := "Pool_" + store.org + "_"
	return prefix + k
}
func (store MsgStore) msgItem(k string) (string, string) {
	if len(store.get(store.orgKey(k))) > 0 {
		return store.get(store.orgKey(k)), "org"
	} else if len(store.get(store.repoKey(k))) > 0 {
		return store.get(store.repoKey(k)), "repo"
	} else if len(store.get(store.poolKey(k))) > 0 {
		return store.get(store.poolKey(k)), "pool"
	} else {
		return "", ""
	}
}
func (store MsgStore) setMsgName() {
	pre, _ := store.msgItem("Name")
	store.set(store.msgKey("NamePre"), pre)
	store.set(store.msgKey("Name"), store.msg.Name)
}
func (store MsgStore) setMsgPat() {
	pre, _ := store.msgItem("Pat")
	store.set(store.msgKey("PatPre"), pre)
	store.set(store.msgKey("Pat"), store.msg.Pat)
}
func (store MsgStore) setMsgURL() {
	pre, _ := store.msgItem("Url")
	store.set(store.msgKey("UrlPre"), pre)
	store.set(store.msgKey("Url"), store.msg.URL)
}
func (store MsgStore) setMsgSize() {
	pre, _ := store.msgItem("Size")
	store.set(store.msgKey("SizePre"), pre)
	store.set(store.msgKey("Size"), store.msg.Size)
}
func (store MsgStore) setMsgCPU() {
	pre, _ := store.msgItem("Cpu")
	store.set(store.msgKey("CpuPre"), pre)
	store.set(store.msgKey("Cpu"), store.msg.CPU)
}
func (store MsgStore) setMsgMemory() {
	pre, _ := store.msgItem("Memory")
	store.set(store.msgKey("MemoryPre"), pre)
	store.set(store.msgKey("Memory"), store.msg.Memory)
}
func (store MsgStore) setMsgLabels() {
	pre, _ := store.msgItem("Labels")
	store.set(store.msgKey("LabelsPre"), pre)
	store.set(store.msgKey("Labels"), store.msg.Labels)
}
func (store MsgStore) setMsgChargeLabels() {
	pre, _ := store.msgItem("ChargeLabels")
	store.set(store.msgKey("ChargeLabelsPre"), pre)
	store.set(store.msgKey("ChargeLabels"), store.msg.ChargeLabels)
}
func (store MsgStore) setMsgRunnerGroup() {
	pre, _ := store.msgItem("RunnerGroup")
	store.set(store.msgKey("RunnerGroupPre"), pre)
	store.set(store.msgKey("RunnerGroup"), store.msg.RunnerGroup)
}
func (store MsgStore) setMsgKey() {
	pre, _ := store.msgItem("Key")
	store.set(store.msgKey("KeyPre"), pre)
	store.set(store.msgKey("Key"), store.msg.Key)
}
func (store MsgStore) setMsgSecret() {
	pre, _ := store.msgItem("Secret")
	store.set(store.msgKey("SecretPre"), pre)
	store.set(store.msgKey("Secret"), store.msg.Secret)
}
func (store MsgStore) setMsgRegion() {
	pre, _ := store.msgItem("Region")
	store.set(store.msgKey("RegionPre"), pre)
	store.set(store.msgKey("Region"), store.msg.Region)
}
func (store MsgStore) setMsgSecGpID() {
	pre, _ := store.msgItem("SecGpId")
	store.set(store.msgKey("SecGpIdPre"), pre)
	store.set(store.msgKey("SecGpId"), store.msg.SecGpID)
}
func (store MsgStore) setMsgVSwitchID() {
	pre, _ := store.msgItem("VSwitchId")
	store.set(store.msgKey("VSwitchIdPre"), pre)
	store.set(store.msgKey("VSwitchId"), store.msg.VSwitchID)
}
func (store MsgStore) setMsgCloudProvider() {
	pre, _ := store.msgItem("CloudProvider")
	store.set(store.msgKey("CloudProviderPre"), pre)
	store.set(store.msgKey("CloudProvider"), store.msg.CloudProvider)
}
func (store MsgStore) setMsgGcpCredential() {
	pre, _ := store.msgItem("GcpCredential")
	store.set(store.msgKey("GcpCredentialPre"), pre)
	store.set(store.msgKey("GcpCredential"), store.msg.GcpCredential)
}
func (store MsgStore) setMsgRepoRegToken() {
	pre, _ := store.msgItem("RepoRegToken")
	store.set(store.msgKey("RepoRegTokenPre"), pre)
	store.set(store.msgKey("RepoRegToken"), store.msg.RepoRegToken)
}
func (store MsgStore) setMsgGcpProject() {
	pre, _ := store.msgItem("GcpProject")
	store.set(store.msgKey("GcpProjectPre"), pre)
	store.set(store.msgKey("GcpProject"), store.msg.GcpProject)
}
func (store MsgStore) setMsgGcpRegion() {
	pre, _ := store.msgItem("GcpRegion")
	store.set(store.msgKey("GcpRegionPre"), pre)
	store.set(store.msgKey("GcpRegion"), store.msg.GcpRegion)
}
func (store MsgStore) setMsgGcpSA() {
	pre, _ := store.msgItem("GcpSA")
	store.set(store.msgKey("GcpSAPre"), pre)
	store.set(store.msgKey("GcpSA"), store.msg.GcpSA)
}
func (store MsgStore) setMsgGcpApikey() {
	pre, _ := store.msgItem("GcpApikey")
	store.set(store.msgKey("GcpApikeyPre"), pre)
	store.set(store.msgKey("GcpApikey"), store.msg.GcpApikey)
}
func (store MsgStore) setMsgGcpDind() {
	pre, _ := store.msgItem("GcpDind")
	store.set(store.msgKey("GcpDindPre"), pre)
	store.set(store.msgKey("GcpDind"), store.msg.GcpDind)
}
func (store MsgStore) setMsgGcpVpc() {
	pre, _ := store.msgItem("GcpVpc")
	store.set(store.msgKey("GcpVpcPre"), pre)
	store.set(store.msgKey("GcpVpc"), store.msg.GcpVpc)
}
func (store MsgStore) setMsgGcpSubnet() {
	pre, _ := store.msgItem("GcpSubnet")
	store.set(store.msgKey("GcpSubnetPre"), pre)
	store.set(store.msgKey("GcpSubnet"), store.msg.GcpSubnet)
}
func (store MsgStore) setMsgImageVersion() {
	pre, _ := store.msgItem("ImageVersion")
	store.set(store.msgKey("ImageVersionPre"), pre)
	store.set(store.msgKey("ImageVersion"), store.msg.ImageVersion)
}
func (store MsgStore) setMsgArmClientID() {
	pre, _ := store.msgItem("ArmClientId")
	store.set(store.msgKey("ArmClientIdPre"), pre)
	store.set(store.msgKey("ArmClientId"), store.msg.ArmClientID)
}
func (store MsgStore) setMsgAciLocation() {
	pre, _ := store.msgItem("AciLocation")
	store.set(store.msgKey("AciLocationPre"), pre)
	store.set(store.msgKey("AciLocation"), store.msg.AciLocation)
}
func (store MsgStore) setMsgAciSku() {
	pre, _ := store.msgItem("AciSku")
	store.set(store.msgKey("AciSkuPre"), pre)
	store.set(store.msgKey("AciSku"), store.msg.AciSku)
}
func (store MsgStore) setMsgAciNetworkType() {
	pre, _ := store.msgItem("AciNetworkType")
	store.set(store.msgKey("AciNetworkTypePre"), pre)
	store.set(store.msgKey("AciNetworkType"), store.msg.AciNetworkType)
}
func (store MsgStore) setMsgArmClientSecret() {
	pre, _ := store.msgItem("ArmClientSecret")
	store.set(store.msgKey("ArmClientSecretPre"), pre)
	store.set(store.msgKey("ArmClientSecret"), store.msg.ArmClientSecret)
}
func (store MsgStore) setMsgArmSubscriptionID() {
	pre, _ := store.msgItem("ArmSubscriptionId")
	store.set(store.msgKey("ArmSubscriptionIdPre"), pre)
	store.set(store.msgKey("ArmSubscriptionId"), store.msg.ArmSubscriptionID)
}
func (store MsgStore) setMsgArmTenantID() {
	pre, _ := store.msgItem("ArmTenantId")
	store.set(store.msgKey("ArmTenantIdPre"), pre)
	store.set(store.msgKey("ArmTenantId"), store.msg.ArmTenantID)
}
func (store MsgStore) setMsgArmEnvironment() {
	pre, _ := store.msgItem("ArmEnvironment")
	store.set(store.msgKey("ArmEnvironmentPre"), pre)
	store.set(store.msgKey("ArmEnvironment"), store.msg.ArmEnvironment)
}
func (store MsgStore) setMsgArmRPRegistration() {
	pre, _ := store.msgItem("ArmRPRegistration")
	store.set(store.msgKey("ArmRPRegistrationPre"), pre)
	store.set(store.msgKey("ArmRPRegistration"), store.msg.ArmRPRegistration)
}
func (store MsgStore) setMsgArmResourceGroupName() {
	pre, _ := store.msgItem("ArmResourceGroupName")
	store.set(store.msgKey("ArmResourceGroupNamePre"), pre)
	store.set(store.msgKey("ArmResourceGroupName"), store.msg.ArmResourceGroupName)
}
func (store MsgStore) setMsgArmSubnetID() {
	pre, _ := store.msgItem("ArmSubnetID")
	store.set(store.msgKey("ArmSubnetIDPre"), pre)
	store.set(store.msgKey("ArmSubnetID"), store.msg.ArmSubnetID)
}
func (store MsgStore) setMsgArmLogAnalyticsWorkspaceID() {
	pre, _ := store.msgItem("ArmLogAnalyticsWorkspaceID")
	store.set(store.msgKey("ArmLogAnalyticsWorkspaceIDPre"), pre)
	store.set(store.msgKey("ArmLogAnalyticsWorkspaceID"), store.msg.ArmLogAnaWorkspaceID)
}
func (store MsgStore) setMsgArmLogAnalyticsWorkspaceKey() {
	pre, _ := store.msgItem("ArmLogAnalyticsWorkspaceKey")
	store.set(store.msgKey("ArmLogAnalyticsWorkspaceKeyPre"), pre)
	store.set(store.msgKey("ArmLogAnalyticsWorkspaceKey"), store.msg.ArmLogAnaWorkspaceKey)
}

func (store MsgStore) Save() {
	store.setMsgName()
	store.setMsgPat()
	store.setMsgURL()
	store.setMsgSize()
	store.setMsgKey()
	store.setMsgSecret()
	store.setMsgRegion()
	store.setMsgSecGpID()
	store.setMsgVSwitchID()
	store.setMsgCloudProvider()
	store.setMsgCPU()
	store.setMsgMemory()
	store.setMsgLabels()
	store.setMsgChargeLabels()
	store.setMsgRunnerGroup()
	store.setMsgGcpCredential()
	store.setMsgRepoRegToken()
	store.setMsgGcpProject()
	store.setMsgGcpRegion()
	store.setMsgGcpSA()
	store.setMsgGcpApikey()
	store.setMsgGcpDind()
	store.setMsgGcpVpc()
	store.setMsgGcpSubnet()
	store.setMsgImageVersion()
	store.setMsgArmClientID()
	store.setMsgAciLocation()
	store.setMsgAciSku()
	store.setMsgAciNetworkType()
	store.setMsgArmClientSecret()
	store.setMsgArmSubscriptionID()
	store.setMsgArmTenantID()
	store.setMsgArmEnvironment()
	store.setMsgArmRPRegistration()
	store.setMsgArmResourceGroupName()
	store.setMsgArmSubnetID()
	store.setMsgArmLogAnalyticsWorkspaceID()
	store.setMsgArmLogAnalyticsWorkspaceKey()
}

func (store MsgStore) UpdateRepoRegToken(token string) {
	findKey := func(k string) string {
		n, t := store.GetName()
		return strings.Title(t) + "_" + n + "_" + k
	}

	pre, _ := store.msgItem("RepoRegToken")
	store.set(findKey("RepoRegTokenPre"), pre)
	store.set(findKey("RepoRegToken"), token)
	RegTokenReuse[findKey("RepoRegTokenPre")] = pre
	RegTokenReuse[findKey("RepoRegToken")] = token
}

func (store MsgStore) GetName() (string, string) {
	return store.msgItem("Name")
}
func (store MsgStore) GetAPIEntTk() string {
	return store.get(store.entk)
}
func (store MsgStore) GetAPIGitTk() string {
	return store.get(store.gittk)
}
func (store MsgStore) DefaultPatToRepo(pat string, url string, ut string) (string, string) {
	crypto := DefaultCryptography(store.get(store.enckey))
	if (pat == "null" || len(pat) == 0) && len(url) > 0 {
		logrus.Infof("GetPat gitfqdn: %s, enfqdn: %s", store.gitfqdn, store.enfqdn)
		if strings.Contains(url, store.gitfqdn) {
			logrus.Infof("GetPat use default git tk")
			return crypto.EncryptMsg(store.get(store.gittk)), "repo"
		} else if strings.Contains(url, store.enfqdn) {
			logrus.Infof("GetPat use default en tk")
			return crypto.EncryptMsg(store.get(store.entk)), "repo"
		}
	}
	return crypto.EncryptMsg(pat), ut
}
func (store MsgStore) GetPat() (string, string) {
	pat, t := store.msgItem("Pat")
	url, _ := store.msgItem("Url")
	logrus.Infof("GetPat url: %s, ut: %s", url, t)
	pat, pt := store.DefaultPatToRepo(pat, url, t)
	return pat, pt
}
func (store MsgStore) GetURL() (string, string) {
	return store.msgItem("Url")
}
func (store MsgStore) GetSize() (string, string) {
	return store.msgItem("Size")
}
func (store MsgStore) GetCPU() (string, string) {
	return store.msgItem("Cpu")
}
func (store MsgStore) GetMemory() (string, string) {
	return store.msgItem("Memory")
}
func (store MsgStore) GetLabels() (string, string) {
	return store.msgItem("Labels")
}
func (store MsgStore) GetChargeLabels() (string, string) {
	return store.msgItem("ChargeLabels")
}
func (store MsgStore) GetRunnerGroup() (string, string) {
	return store.msgItem("RunnerGroup")
}
func (store MsgStore) GetKey() (string, string) {
	return store.msgItem("Key")
}
func (store MsgStore) GetSecret() (string, string) {
	return store.msgItem("Secret")
}
func (store MsgStore) GetRegion() (string, string) {
	return store.msgItem("Region")
}
func (store MsgStore) GetSecGpID() (string, string) {
	return store.msgItem("SecGpId")
}
func (store MsgStore) GetVSwitchID() (string, string) {
	return store.msgItem("VSwitchId")
}
func (store MsgStore) GetCloudProvider() (string, string) {
	return store.msgItem("CloudProvider")
}
func (store MsgStore) GetGcpCredential() (string, string) {
	return store.msgItem("GcpCredential")
}
func (store MsgStore) GetRepoRegToken() (string, string) {
	findKey := func(k string) string {
		n, t := store.GetName()
		return strings.Title(t) + "_" + n + "_" + k
	}
	val, t := store.msgItem("RepoRegToken")
	if tk, ok := RegTokenReuse[findKey("RepoRegToken")]; ok {
		logrus.Infof("RepoRegToken %s %s", tk, t)
		return tk, t
	} else {
		logrus.Infof("RepoRegToken ori %s %v", val, ok)
		return val, t
	}
}
func (store MsgStore) GetGcpProject() (string, string) {
	return store.msgItem("GcpProject")
}
func (store MsgStore) GetGcpRegion() (string, string) {
	return store.msgItem("GcpRegion")
}
func (store MsgStore) GetGcpSA() (string, string) {
	return store.msgItem("GcpSA")
}
func (store MsgStore) GetGcpApikey() (string, string) {
	return store.msgItem("GcpApikey")
}
func (store MsgStore) GetGcpDind() (string, string) {
	return store.msgItem("GcpDind")
}
func (store MsgStore) GetGcpVpc() (string, string) {
	return store.msgItem("GcpVpc")
}
func (store MsgStore) GetGcpSubnet() (string, string) {
	return store.msgItem("GcpSubnet")
}
func (store MsgStore) GetImageVersion() (string, string) {
	return store.msgItem("ImageVersion")
}
func (store MsgStore) GetArmClientID() (string, string) {
	return store.msgItem("ArmClientId")
}
func (store MsgStore) GetAciLocation() (string, string) {
	return store.msgItem("AciLocation")
}
func (store MsgStore) GetAciSku() (string, string) {
	return store.msgItem("AciSku")
}
func (store MsgStore) GetAciNetworkType() (string, string) {
	return store.msgItem("AciNetworkType")
}
func (store MsgStore) GetArmClientSecret() (string, string) {
	return store.msgItem("ArmClientSecret")
}
func (store MsgStore) GetArmSubscriptionID() (string, string) {
	return store.msgItem("ArmSubscriptionId")
}
func (store MsgStore) GetArmTenantID() (string, string) {
	return store.msgItem("ArmTenantId")
}
func (store MsgStore) GetArmEnvironment() (string, string) {
	return store.msgItem("ArmEnvironment")
}
func (store MsgStore) GetArmRPRegistration() (string, string) {
	return store.msgItem("ArmRPRegistration")
}
func (store MsgStore) GetArmResourceGroupName() (string, string) {
	return store.msgItem("ArmResourceGroupName")
}
func (store MsgStore) GetArmSubnetID() (string, string) {
	return store.msgItem("ArmSubnetID")
}
func (store MsgStore) GetArmLogAnalyticsWorkspaceID() (string, string) {
	return store.msgItem("ArmLogAnalyticsWorkspaceID")
}
func (store MsgStore) GetArmLogAnalyticsWorkspaceKey() (string, string) {
	return store.msgItem("ArmLogAnalyticsWorkspaceKey")
}
func (store MsgStore) GetPreSize() string {
	item, _ := store.msgItem("SizePre")
	return item
}
func (store MsgStore) GetPreCPU() string {
	item, _ := store.msgItem("CpuPre")
	return item
}
func (store MsgStore) GetPreMemory() string {
	item, _ := store.msgItem("MemoryPre")
	return item
}
func (store MsgStore) GetPreLabels() string {
	item, _ := store.msgItem("LabelsPre")
	return item
}
func (store MsgStore) GetPreChargeLabels() string {
	item, _ := store.msgItem("ChargeLabelsPre")
	return item
}
func (store MsgStore) GetPreRunnerGroup() string {
	item, _ := store.msgItem("RunnerGroupPre")
	return item
}
func (store MsgStore) GetPreKey() string {
	item, _ := store.msgItem("KeyPre")
	return item
}
func (store MsgStore) GetPreSecret() string {
	item, _ := store.msgItem("SecretPre")
	return item
}
func (store MsgStore) GetPreRegion() string {
	item, _ := store.msgItem("RegionPre")
	return item
}
func (store MsgStore) GetPreSecGpID() string {
	item, _ := store.msgItem("SecGpIdPre")
	return item
}
func (store MsgStore) GetPreVSwitchID() string {
	item, _ := store.msgItem("VSwitchIdPre")
	return item
}
func (store MsgStore) GetPreCloudProvider() string {
	item, _ := store.msgItem("CloudProviderPre")
	return item
}
func (store MsgStore) GetPrePat() string {
	item, _ := store.msgItem("PatPre")
	return item
}
func (store MsgStore) GetPreURL() string {
	item, _ := store.msgItem("UrlPre")
	return item
}
func (store MsgStore) GetPreGcpCredential() string {
	item, _ := store.msgItem("GcpCredentialPre")
	return item
}
func (store MsgStore) GetPreRepoRegToken() string {
	item, _ := store.msgItem("RepoRegTokenPre")
	return item
}
func (store MsgStore) GetPreGcpProject() string {
	item, _ := store.msgItem("GcpProjectPre")
	return item
}
func (store MsgStore) GetPreGcpRegion() string {
	item, _ := store.msgItem("GcpRegionPre")
	return item
}
func (store MsgStore) GetPreGcpSA() string {
	item, _ := store.msgItem("GcpSAPre")
	return item
}
func (store MsgStore) GetPreGcpApikey() string {
	item, _ := store.msgItem("GcpApikeyPre")
	return item
}
func (store MsgStore) GetPreGcpDind() string {
	item, _ := store.msgItem("GcpDindPre")
	return item
}
func (store MsgStore) GetPreGcpVpc() string {
	item, _ := store.msgItem("GcpVpcPre")
	return item
}
func (store MsgStore) GetPreGcpSubnet() string {
	item, _ := store.msgItem("GcpSubnetPre")
	return item
}
func (store MsgStore) GetPreImageVersion() string {
	item, _ := store.msgItem("ImageVersionPre")
	return item
}
func (store MsgStore) GetPreArmClientID() string {
	item, _ := store.msgItem("ArmClientIdPre")
	return item
}
func (store MsgStore) GetPreAciLocation() string {
	item, _ := store.msgItem("AciLocationPre")
	return item
}
func (store MsgStore) GetPreAciSku() string {
	item, _ := store.msgItem("AciSkuPre")
	return item
}
func (store MsgStore) GetPreAciNetworkType() string {
	item, _ := store.msgItem("AciNetworkTypePre")
	return item
}
func (store MsgStore) GetPreArmClientSecret() string {
	item, _ := store.msgItem("ArmClientSecretPre")
	return item
}
func (store MsgStore) GetPreArmSubscriptionID() string {
	item, _ := store.msgItem("ArmSubscriptionIdPre")
	return item
}
func (store MsgStore) GetPreArmTenantID() string {
	item, _ := store.msgItem("ArmTenantIdPre")
	return item
}
func (store MsgStore) GetPreArmEnvironment() string {
	item, _ := store.msgItem("ArmEnvironmentPre")
	return item
}
func (store MsgStore) GetPreArmRPRegistration() string {
	item, _ := store.msgItem("ArmRPRegistrationPre")
	return item
}
func (store MsgStore) GetPreArmResourceGroupName() string {
	item, _ := store.msgItem("ArmResourceGroupNamePre")
	return item
}
func (store MsgStore) GetPreArmSubnetID() string {
	item, _ := store.msgItem("ArmSubnetIDPre")
	return item
}

func (store MsgStore) AnyChange() bool {
	url, _ := store.GetURL()
	pat, _ := store.GetPat()
	sg, _ := store.GetSecGpID()
	sw, _ := store.GetVSwitchID()
	pr, _ := store.GetCloudProvider()
	region, _ := store.GetRegion()
	secret, _ := store.GetSecret()
	key, _ := store.GetKey()
	size, _ := store.GetSize()
	cpu, _ := store.GetCPU()
	memory, _ := store.GetMemory()
	labels, _ := store.GetLabels()
	chargelabels, _ := store.GetChargeLabels()
	runnergroup, _ := store.GetRunnerGroup()
	gcpCredentials, _ := store.GetGcpCredential()
	repoRegToken, _ := store.GetRepoRegToken()
	gcpProject, _ := store.GetGcpProject()
	gcpRegion, _ := store.GetGcpRegion()
	gcpSa, _ := store.GetGcpSA()
	gcpApikey, _ := store.GetGcpApikey()
	gcpDind, _ := store.GetGcpDind()
	gcpVpc, _ := store.GetGcpVpc()
	gcpSubnet, _ := store.GetGcpSubnet()
	imageVer, _ := store.GetImageVersion()
	armClientID, _ := store.GetArmClientID()
	aciLocation, _ := store.GetAciLocation()
	aciSku, _ := store.GetAciSku()
	aciNetworkType, _ := store.GetAciNetworkType()
	armClientSecret, _ := store.GetArmClientSecret()
	armSubscriptionID, _ := store.GetArmSubscriptionID()
	armTenantID, _ := store.GetArmTenantID()
	armEnvironment, _ := store.GetArmEnvironment()
	armRpRegistration, _ := store.GetArmRPRegistration()
	armResourceGroupName, _ := store.GetArmResourceGroupName()
	armSubnetID, _ := store.GetArmSubnetID()
	sameConfig := store.GetPreURL() == url && store.GetPrePat() == pat &&
		store.GetPreSecGpID() == sg && store.GetPreVSwitchID() == sw &&
		store.GetPreRegion() == region && store.GetPreSecret() == secret &&
		store.GetPreKey() == key && store.GetPreSize() == size &&
		store.GetPreCPU() == cpu && store.GetPreMemory() == memory &&
		store.GetPreLabels() == labels && store.GetPreChargeLabels() == chargelabels &&
		store.GetPreRunnerGroup() == runnergroup && store.GetPreCloudProvider() == pr &&
		gcpCredentials == store.GetPreGcpCredential() &&
		repoRegToken == store.GetPreRepoRegToken() &&
		gcpProject == store.GetPreGcpProject() &&
		gcpRegion == store.GetPreGcpRegion() &&
		gcpSa == store.GetPreGcpSA() &&
		gcpApikey == store.GetPreGcpApikey() &&
		gcpDind == store.GetPreGcpDind() &&
		gcpVpc == store.GetPreGcpVpc() &&
		gcpSubnet == store.GetPreGcpSubnet() &&
		imageVer == store.GetPreImageVersion() &&
		armClientID == store.GetPreArmClientID() &&
		aciLocation == store.GetPreAciLocation() &&
		aciSku == store.GetPreAciSku() &&
		aciNetworkType == store.GetPreAciNetworkType() &&
		armClientSecret == store.GetPreArmClientSecret() &&
		armSubscriptionID == store.GetPreArmSubscriptionID() &&
		armTenantID == store.GetPreArmTenantID() &&
		armEnvironment == store.GetPreArmEnvironment() &&
		armRpRegistration == store.GetPreArmRPRegistration() &&
		armResourceGroupName == store.GetPreArmResourceGroupName() &&
		armSubnetID == store.GetPreArmSubnetID()
	return !sameConfig
}
