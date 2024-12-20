## 005 Tenant Testig
WF_SLS_TEAM := "005-testing"
WF_SLS_DOMAIN := github.com
WF_SLS_ORG_NAME := ingka-group-digital
WF_SLS_REPOS_NAME := "cdh-br-ark-ei,app-monitor-agent,serverless-hosted-runner"
WF_SLS_TYPE := "Org" #  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "2.0"
WF_SLS_SIZE_MEMORY := "8.0"
WF_SLS_RUNNER_GROUP := "default"
WF_SLS_LABELS := "serverless-hosted-runner-restore,app-monitor-agent,serverless-hosted-runner-gcp" #label should not contains blank
WF_SLS_CHARGE_LABELS := "ingka-group-digital-app-monitor-agent" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_ARM_ENV := "china"
WF_SLS_ARM_RP_REG := "none"
WF_SLS_ARM_SUB_ID := "b5937c02-df3d-4846-849b-6fe858a84d0e"
WF_SLS_ARM_RG_NAME := "sls-runner"
WF_SLS_ARM_SUBNET_ID := "/subscriptions/b5937c02-df3d-4846-849b-6fe858a84d0e/resourceGroups/rg-auto-cn-north3-test/providers/Microsoft.Network/virtualNetworks/vnet-auto-cn-north3-test/subnets/sls-runner-acl-subnet"
WF_SLS_AZURE := "\"ArmEnvironment\":\"${WF_SLS_ARM_ENV}\",\"ArmRPRegistration\":\"${WF_SLS_ARM_RP_REG}\",\"ArmSubscriptionId\":\"${WF_SLS_ARM_SUB_ID}\",\"ArmResourceGroupName\":\"${WF_SLS_ARM_RG_NAME}\",\"ArmSubnetId\":\"${WF_SLS_ARM_SUBNET_ID}\""
# WF_SLS_REG_ITEM1 := "{${WF_SLS_REG_COMMON_AZURE}, ${WF_SLS_AZURE}, ${WF_SLS_REG_COMMON}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\",\"RunnerGroup\":\"${WF_SLS_RUNNER_GROUP}\"}"
WF_SLS_REG_ITEM1 := "{${WF_SLS_REG_COMMON_GCP}, ${WF_SLS_REG_COMMON}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\",\"RunnerGroup\":\"${WF_SLS_RUNNER_GROUP}\"}"
REG1 := $(WF_SLS_REG_ITEM1)
WF_SLS_REGS := "[${REG1}]"
## 005 Tenant Testig
WF_SLS_TEAM := "005-testing"
WF_SLS_DOMAIN := git.build.ingka.ikea.com
WF_SLS_ORG_NAME := china-digital-hub
WF_SLS_REPOS_NAME := "serverless-runner-testing"
WF_SLS_TYPE := "Org" #  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "2.0"
WF_SLS_SIZE_MEMORY := "8.0"
WF_SLS_RUNNER_GROUP := "default"
WF_SLS_LABELS := "serverless-runner-testing" #label should not contains blank
WF_SLS_CHARGE_LABELS := "china-digital-hub-serverless-runner-testing" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM2 := "{${WF_SLS_REG_COMMON}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\",\"RunnerGroup\":\"${WF_SLS_RUNNER_GROUP}\"}"
REG2 := $(WF_SLS_REG_ITEM2)
WF_SLS_REGS := "[${REG1},${REG2}]"
## 005 Tenant Testing
# WF_SLS_DOMAIN := git.build.ingka.ikea.com
# WF_SLS_ORG_NAME := labrador
# WF_SLS_REPOS_NAME := "sentry-exporter"
# WF_SLS_TYPE := "Org"  
# WF_SLS_SIZE := "1"
# WF_SLS_SIZE_CPU := "0.5"
# WF_SLS_SIZE_MEMORY := "1.0"
# WF_SLS_LABELS := "cn-runner" #label should not contains blank
# WF_SLS_CHARGE_LABELS := "sentry-group-charged" #label should not contains blank
# WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
# WF_SLS_REG_ITEM2 := "{${WF_SLS_REG_COMMON}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
# REG2 := $(WF_SLS_REG_ITEM2)
# WF_SLS_REGS := "[${REG1},${REG2}]"