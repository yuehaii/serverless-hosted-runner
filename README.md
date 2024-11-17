# Serverless-hosted-runner

## Introduction
This application is a kind of Github Action [Self-hosted Runner](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/about-self-hosted-runners).
It is light-weight and has enterprize, organization and repo level runners. It also supports the JIT, ephemeral runner and runner pool. 
This application contains 'dispatcher' and 'runer' micro-services. Each service can be deployed in same or different tenant. So it can be deployed as distributed dispatcher mode (every tenant has its event dispatcher and runner) and centralized dispatcher mode (which has only one centralized dispatcher for different tenant). 

## Deploy Mode 
### Distributed dispatcher mode 
The application's dipatcher and runner are deployed under same tenant. Different team/tenant has isolated dispacher/runner. Each team can control its dispatcher/runner behavior. It will be easy to calculate the cloud fee under such mode. It will also reduce the workload of the dispacher since each tenant has its own dispatcher.
#### Installation
We can add below configuraton and run 'make lazy_install' to deploy the dispacher/runner into same tenant. Then assign 'l-ccoecn-a-itcnshg' as Admin of your repo/org and change the workflow run on label as "serverless-hosted-runner" or your [customized lable](https://none).
```bash
# add your env.sh
export ALICLOUD_REGION=cn-shanghai
export ALICLOUD_ACCESS_KEY=<AliRAMAccessKey>
export ALICLOUD_SECRET_KEY=<AliRAMAccessPassword>
export TF_VAR_IMAGE_RETRIEVE_USERNAME=<DockerImageRetrieveUserName>
export TF_VAR_IMAGE_RETRIEVE_PWD=<DockerImageRetrievePassword>
export TF_VAR_IMAGE_RETRIEVE_SERVER=<DockerImageServerHostName>, e.g, artifactory.cloud.ingka-system.cn
export DISPATCHER_SG_ID=<DispacherAliSecurityGroupId>
export DISPATCHER_VSWITCH_ID=<DispatcherVSwitchId> 
export SLS_ENC_KEY="xxxxx"
export SLS_GITHUB_TK="ghp_xxxxx"
export SLS_GITENT_TK="ghp_xxxxx"
 
# add a Registration_xxx.mk under ./registration
# and include this registration in Makefile
include ./Registration.mk
include ./registration/Registration_test.mk

# sample ./Registratation_test.mk 
WF_SLS_DOMAIN := git.build.ingka.ikea.com
WF_SLS_ORG_NAME := labrador
WF_SLS_REPOS_NAME := "sentry-exporter"
WF_SLS_TYPE := "Org"  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "2.0"
WF_SLS_SIZE_MEMORY := "4.0"
WF_SLS_LABELS := "cn-runner" #label should not contains blank
WF_SLS_CHARGE_LABELS := "sentry-group-charged" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM2 := "{${WF_SLS_REG_COMMON}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG2 := $(WF_SLS_REG_ITEM2)
WF_SLS_REGS := "[${REG2}]"
```
If we want to add/change orgnaziation or repos configuration, please add/change the configuration under ./registration/Registration_<registration_alias>.mk and run 'make lazy_install ralias=<registration_alias>' again. 

### Centralized dispacther mode 
Under such mode, it has only one dispatcher. Different tenants can register their runner and let the dispatcher creating the runner into specific tenant. The centralized dispacher cost need to be shared with multiple teams. 
#### Allen portal registration
Please follow the serverless runner [onboarding process](https://none) to register your runner on Allen portal.
#### Lazy registration 
We can still use 'make lazy_install' to install the centralized dispacther. Just make sure the repo/org's runner registration's "Key" and "Secret" in Registration.mk are different with dispacher's
```bash
WF_SLS_REG_COMMON := "\"Pat\":\"\",\"Key\":\"${YouRunnerTenatnAliAccessKey}\",\"Secret\":\"${YouRunnerTenatnAliAccessSecret}\",\"Region\":\"${WF_SLS_REGION}\",\"SecGpId\":\"${WF_SLS_SECGROUP_ID}\",\"VSwitchId\":\"${WF_SLS_VSWITCH_ID}\",\"PullInterval\":\"${WF_SLS_PULL_INTERVAL_Sec}\""
```

## Preparation
1. Please make sure to have the security group and vswitch configured to host runner. 
2. Make sure to assign 'l-ccoecn-a-itcnshg' as Admin of your repo/org.
3. Then workflow run-on labels match the runner default label or customized label.

## Build and Deployment
1. Configure the env.sh and ./registrations (lazy_install mode).
2. Run below command to build the images.
```bash
make image
```
3. Run below command to install the image with allen portal registrtion
```bash 
make install
```
or run below command for lazy registration
```bash 
make lazy_install ralias=<registration_alias>
```

## CCOECN centralized dispatcher
If you don't want to deploy dispacher. You can use CCOECN team's dispacher instead. Below are steps:
1. Assign 'l-ccoecn-a-itcnshg' as Admin of your Repo/Org. 
2. Prepare your run-on labels, charge labels, and the repos names.
4. Ping hayue2 on teams for application. 

## Dynamic runner size
Runner CPU and memory size can be configured in lazy configuration or Allen configuration. We also provide a workflow level dynamic runner size. Please use it with customized label or default label. E.g:
```yaml
    runs-on: 
      labels: [serverless-hosted-runner, cpu-0.5, memory-1.0]
```
This config priority is higher than lazy/allen configuration. It is designed for some special workflow requirement.

## Multiple Cloud
Serverless runner can be deployed on multiple cloud. It can be used for different teams with Ali, Azure, GCP, AWS cloud.
### Azure cloud configuration. 
For Azure cloud, please follow below configuration. 
1. Please add the [arm environment registration](https://github.com/ingka-group-digital/serverless-hosted-runner/blob/feat-multi-cloud/Registration.mk#L9C31-L9C42) and [aci registration](https://github.com/ingka-group-digital/serverless-hosted-runner/blob/feat-multi-cloud/registration/Registration_test.mk#L19).
2. Running the registration workflow and select "cloud provider" as "azure".
3. Please notice that DinD need privilege which only available in "Confidential" tier. This tier is not available in China Azure Cloud. Please use global Azure Cloud instead. 

## Known issue
CI building may raise "[signal kill](https://github.com/beego/wetalk/issues/32)" error if the runner memory is not enough. You can add label as below to increase the memory size. 
```yaml
    runs-on: 
      labels: [serverless-hosted-runner, cpu-2.0, memory-4.0]
```