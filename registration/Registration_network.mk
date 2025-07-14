# network team network-automation-dev org
WF_SLS_DOMAIN := git.build.ingka.ikea.com
WF_SLS_ORG_NAME := network-automation-dev
WF_SLS_REPOS_NAME := "cn-ansible-fortigate,cn-ansible-huawei-nce,cn-awx-dev,cn-monitoring-dev,cn-public-cloud-network,cn-oob"
WF_SLS_TYPE := "Org"  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "2.0"
WF_SLS_SIZE_MEMORY := "4.0"
WF_SLS_CHARGE_LABELS := "network-cc5333" #label should not contains blank
WF_SLS_LABELS := "self-hosted,dev,git-en-runner,devnet,network" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM3 := "{${WF_SLS_REG_COMMON}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG3 := $(WF_SLS_REG_ITEM3)
# network team network-automation-prod org
WF_SLS_DOMAIN := git.build.ingka.ikea.com
WF_SLS_ORG_NAME := network-automation-prod
WF_SLS_REPOS_NAME := "cn-ansible-fortigate-prd,cn-awx-prod,cn-monitoring-prod"
WF_SLS_TYPE := "Org"
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "2.0"
WF_SLS_SIZE_MEMORY := "4.0"
WF_SLS_LABELS := "self-hosted,prod,git-en-runner,devnet,network" #label should not contains blank
WF_SLS_CHARGE_LABELS := "network-cc5333" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM4 := "{${WF_SLS_REG_COMMON}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG4 := $(WF_SLS_REG_ITEM4)
WF_SLS_REGS := "[${REG3},${REG4}]"