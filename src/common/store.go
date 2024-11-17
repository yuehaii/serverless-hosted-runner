package common

import (
	"os"
	"strings"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

type Store interface {
	Save()
	GetKey() (string, string)
	GetSecret() (string, string)
	GetRegion() (string, string)
	GetSecGpId() (string, string)
	GetVSwitchId() (string, string)
	GetPat() (string, string)
	GetUrl() (string, string)
	GetSize() (string, string)
	GetCpu() (string, string)
	GetMemory() (string, string)
	GetLabels() (string, string)
	GetChargeLabels() (string, string)
	GetRunnerGroup() (string, string)
	GetPreSize() string
	AnyChange() bool
	GetPreKey() string
	GetPreSecret() string
	GetPreRegion() string
	GetPreSecGpId() string
	GetPreVSwitchId() string
	GetPrePat() string
	GetPreUrl() string
	GetAPIEntTk() string
	GetAPIGitTk() string
	GetGcpCredential() (string, string)
	GetGcpProject() (string, string)
	GetGcpRegion() (string, string)
	GetArmClientId() (string, string)
	GetArmClientSecret() (string, string)
	GetArmSubscriptionId() (string, string)
	GetArmTenantId() (string, string)
	GetArmEnvironment() (string, string)
	GetArmRPRegistration() (string, string)
	GetArmResourceGroupName() (string, string)
	GetArmSubnetID() (string, string)
	GetArmLogAnalyticsWorkspaceID() (string, string)
	GetArmLogAnalyticsWorkspaceKey() (string, string)
	GetPreGcpCredential() string
	GetPreGcpProject() string
	GetPreGcpRegion() string
	GetPreArmClientId() string
	GetPreArmClientSecret() string
	GetPreArmSubscriptionId() string
	GetPreArmTenantId() string
	GetPreArmEnvironment() string
	GetPreArmRPRegistration() string
	GetPreArmResourceGroupName() string
	GetPreArmSubnetID() string
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

func (store MsgStore) get(key string) string {
	return os.Getenv(key)
}
func (store MsgStore) set(key string, val string) {
	os.Setenv(key, val)
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
func (store MsgStore) setMsgUrl() {
	pre, _ := store.msgItem("Url")
	store.set(store.msgKey("UrlPre"), pre)
	store.set(store.msgKey("Url"), store.msg.Url)
}
func (store MsgStore) setMsgSize() {
	pre, _ := store.msgItem("Size")
	store.set(store.msgKey("SizePre"), pre)
	store.set(store.msgKey("Size"), store.msg.Size)
}
func (store MsgStore) setMsgCpu() {
	pre, _ := store.msgItem("Cpu")
	store.set(store.msgKey("CpuPre"), pre)
	store.set(store.msgKey("Cpu"), store.msg.Cpu)
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
func (store MsgStore) setMsgSecGpId() {
	pre, _ := store.msgItem("SecGpId")
	store.set(store.msgKey("SecGpIdPre"), pre)
	store.set(store.msgKey("SecGpId"), store.msg.SecGpId)
}
func (store MsgStore) setMsgVSwitchId() {
	pre, _ := store.msgItem("VSwitchId")
	store.set(store.msgKey("VSwitchIdPre"), pre)
	store.set(store.msgKey("VSwitchId"), store.msg.VSwitchId)
}

func (store MsgStore) setMsgGcpCredential() {
	pre, _ := store.msgItem("GcpCredential")
	store.set(store.msgKey("GcpCredentialPre"), pre)
	store.set(store.msgKey("GcpCredential"), store.msg.GcpCredential)
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
func (store MsgStore) setMsgArmClientId() {
	pre, _ := store.msgItem("ArmClientId")
	store.set(store.msgKey("ArmClientIdPre"), pre)
	store.set(store.msgKey("ArmClientId"), store.msg.ArmClientId)
}
func (store MsgStore) setMsgArmClientSecret() {
	pre, _ := store.msgItem("ArmClientSecret")
	store.set(store.msgKey("ArmClientSecretPre"), pre)
	store.set(store.msgKey("ArmClientSecret"), store.msg.ArmClientSecret)
}
func (store MsgStore) setMsgArmSubscriptionId() {
	pre, _ := store.msgItem("ArmSubscriptionId")
	store.set(store.msgKey("ArmSubscriptionIdPre"), pre)
	store.set(store.msgKey("ArmSubscriptionId"), store.msg.ArmSubscriptionId)
}
func (store MsgStore) setMsgArmTenantId() {
	pre, _ := store.msgItem("ArmTenantId")
	store.set(store.msgKey("ArmTenantIdPre"), pre)
	store.set(store.msgKey("ArmTenantId"), store.msg.ArmTenantId)
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
	store.set(store.msgKey("ArmSubnetID"), store.msg.ArmSubnetId)
}
func (store MsgStore) setMsgArmLogAnalyticsWorkspaceID() {
	pre, _ := store.msgItem("ArmLogAnalyticsWorkspaceID")
	store.set(store.msgKey("ArmLogAnalyticsWorkspaceIDPre"), pre)
	store.set(store.msgKey("ArmLogAnalyticsWorkspaceID"), store.msg.ArmLogAnaWorkspaceId)
}
func (store MsgStore) setMsgArmLogAnalyticsWorkspaceKey() {
	pre, _ := store.msgItem("ArmLogAnalyticsWorkspaceKey")
	store.set(store.msgKey("ArmLogAnalyticsWorkspaceKeyPre"), pre)
	store.set(store.msgKey("ArmLogAnalyticsWorkspaceKey"), store.msg.ArmLogAnaWorkspaceKey)
}

func (store MsgStore) Save() {
	store.setMsgName()
	store.setMsgPat()
	store.setMsgUrl()
	store.setMsgSize()
	store.setMsgKey()
	store.setMsgSecret()
	store.setMsgRegion()
	store.setMsgSecGpId()
	store.setMsgVSwitchId()
	store.setMsgCpu()
	store.setMsgMemory()
	store.setMsgLabels()
	store.setMsgChargeLabels()
	store.setMsgRunnerGroup()
	store.setMsgGcpCredential()
	store.setMsgGcpProject()
	store.setMsgGcpRegion()
	store.setMsgArmClientId()
	store.setMsgArmClientSecret()
	store.setMsgArmSubscriptionId()
	store.setMsgArmTenantId()
	store.setMsgArmEnvironment()
	store.setMsgArmRPRegistration()
	store.setMsgArmResourceGroupName()
	store.setMsgArmSubnetID()
	store.setMsgArmLogAnalyticsWorkspaceID()
	store.setMsgArmLogAnalyticsWorkspaceKey()
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
func (store MsgStore) GetUrl() (string, string) {
	return store.msgItem("Url")
}
func (store MsgStore) GetSize() (string, string) {
	return store.msgItem("Size")
}
func (store MsgStore) GetCpu() (string, string) {
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
func (store MsgStore) GetSecGpId() (string, string) {
	return store.msgItem("SecGpId")
}
func (store MsgStore) GetVSwitchId() (string, string) {
	return store.msgItem("VSwitchId")
}
func (store MsgStore) GetGcpCredential() (string, string) {
	return store.msgItem("GcpCredential")
}
func (store MsgStore) GetGcpProject() (string, string) {
	return store.msgItem("GcpProject")
}
func (store MsgStore) GetGcpRegion() (string, string) {
	return store.msgItem("GcpRegion")
}
func (store MsgStore) GetArmClientId() (string, string) {
	return store.msgItem("ArmClientId")
}
func (store MsgStore) GetArmClientSecret() (string, string) {
	return store.msgItem("ArmClientSecret")
}
func (store MsgStore) GetArmSubscriptionId() (string, string) {
	return store.msgItem("ArmSubscriptionId")
}
func (store MsgStore) GetArmTenantId() (string, string) {
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
func (store MsgStore) GetPreCpu() string {
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
func (store MsgStore) GetPreSecGpId() string {
	item, _ := store.msgItem("SecGpIdPre")
	return item
}
func (store MsgStore) GetPreVSwitchId() string {
	item, _ := store.msgItem("VSwitchIdPre")
	return item
}
func (store MsgStore) GetPrePat() string {
	item, _ := store.msgItem("PatPre")
	return item
}
func (store MsgStore) GetPreUrl() string {
	item, _ := store.msgItem("UrlPre")
	return item
}
func (store MsgStore) GetPreGcpCredential() string {
	item, _ := store.msgItem("GcpCredentialPre")
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
func (store MsgStore) GetPreArmClientId() string {
	item, _ := store.msgItem("ArmClientIdPre")
	return item
}
func (store MsgStore) GetPreArmClientSecret() string {
	item, _ := store.msgItem("ArmClientSecretPre")
	return item
}
func (store MsgStore) GetPreArmSubscriptionId() string {
	item, _ := store.msgItem("ArmSubscriptionIdPre")
	return item
}
func (store MsgStore) GetPreArmTenantId() string {
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
	url, _ := store.GetUrl()
	pat, _ := store.GetPat()
	sg, _ := store.GetSecGpId()
	sw, _ := store.GetVSwitchId()
	region, _ := store.GetRegion()
	secret, _ := store.GetSecret()
	key, _ := store.GetKey()
	size, _ := store.GetSize()
	cpu, _ := store.GetCpu()
	memory, _ := store.GetMemory()
	labels, _ := store.GetLabels()
	chargelabels, _ := store.GetChargeLabels()
	runnergroup, _ := store.GetRunnerGroup()
	gcp_credentials, _ := store.GetGcpCredential()
	gcp_project, _ := store.GetGcpProject()
	gcp_region, _ := store.GetGcpRegion()
	arm_client_id, _ := store.GetArmClientId()
	arm_client_secret, _ := store.GetArmClientSecret()
	arm_subscription_id, _ := store.GetArmSubscriptionId()
	arm_tenant_id, _ := store.GetArmTenantId()
	arm_environment, _ := store.GetArmEnvironment()
	arm_rp_registration, _ := store.GetArmRPRegistration()
	arm_resource_group_name, _ := store.GetArmResourceGroupName()
	arm_subnet_id, _ := store.GetArmSubnetID()
	return !(store.GetPreUrl() == url && store.GetPrePat() == pat &&
		store.GetPreSecGpId() == sg && store.GetPreVSwitchId() == sw &&
		store.GetPreRegion() == region && store.GetPreSecret() == secret &&
		store.GetPreKey() == key && store.GetPreSize() == size &&
		store.GetPreCpu() == cpu && store.GetPreMemory() == memory &&
		store.GetPreLabels() == labels && store.GetPreChargeLabels() == chargelabels &&
		store.GetPreRunnerGroup() == runnergroup &&
		gcp_credentials == store.GetPreGcpCredential() &&
		gcp_project == store.GetPreGcpProject() &&
		gcp_region == store.GetPreGcpRegion() &&
		arm_client_id == store.GetPreArmClientId() &&
		arm_client_secret == store.GetPreArmClientSecret() &&
		arm_subscription_id == store.GetPreArmSubscriptionId() &&
		arm_tenant_id == store.GetPreArmTenantId() &&
		arm_environment == store.GetPreArmEnvironment() &&
		arm_rp_registration == store.GetPreArmRPRegistration() &&
		arm_resource_group_name == store.GetPreArmResourceGroupName() &&
		arm_subnet_id == store.GetPreArmSubnetID())
}
