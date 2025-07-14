# 1254676370219887 Tenant MACK
# gcp project
WF_SLS_REG_COMMON_110 := "\"Pat\":\"\",\"Key\":\"${ALICLOUD_ACCESS_KEY_110}\",\"Secret\":\"${ALICLOUD_SECRET_KEY_110}\",\"Region\":\"${WF_SLS_REGION}\",\"SecGpId\":\"${DISPATCHER_SG_ID_110}\",\"VSwitchId\":\"${DISPATCHER_VSWITCH_ID_110}\",\"PullInterval\":\"${WF_SLS_PULL_INTERVAL_Sec}\""
WF_SLS_DOMAIN := git.build.ingka.ikea.com
WF_SLS_ORG_NAME := gcp-task-force
WF_SLS_REPOS_NAME := "ccoe-macp,ccoe-quattro"
WF_SLS_TYPE := "Org"  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "1.0"
WF_SLS_SIZE_MEMORY := "2.0"
WF_SLS_CHARGE_LABELS := "gcp-task-force" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM3 := "{${WF_SLS_REG_COMMON_110}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG3 := $(WF_SLS_REG_ITEM3)

# macp project
WF_SLS_DOMAIN := github.com
WF_SLS_ORG_NAME := ingka-group-digital
WF_SLS_REPOS_NAME := "mmacp-deployment-examples,macp-pulumi,macp-fastapi-kong-demo"
WF_SLS_TYPE := "Org"  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "1.0"
WF_SLS_SIZE_MEMORY := "2.0"
WF_SLS_CHARGE_LABELS := "ingka-group-digital" #label should not contains blank
WF_SLS_LABELS := "macp-deployment-examples,macp-pulumi,macp-fastapi-kong-demo" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM4 := "{${WF_SLS_REG_COMMON_MACP}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG4 := $(WF_SLS_REG_ITEM4)

# apim project
WF_SLS_DOMAIN := github.com
WF_SLS_ORG_NAME := ingka-group-digital
WF_SLS_REPOS_NAME := "apim-kong-gw-docker,apim-opentelemetry-infra"
WF_SLS_TYPE := "Org"  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "1.0"
WF_SLS_SIZE_MEMORY := "2.0"
WF_SLS_CHARGE_LABELS := "apim-runner" #label should not contains blank
WF_SLS_LABELS := "apim-kong-gw-docker,apim-opentelemetry-infra" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM5 := "{${WF_SLS_REG_COMMON_MACP}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG5 := $(WF_SLS_REG_ITEM5)

# upptacka 
WF_SLS_DOMAIN := github.com
WF_SLS_ORG_NAME := ingka-group-digital
WF_SLS_REPOS_NAME := "upptacka-api-find"
WF_SLS_TYPE := "Org"  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "1.0"
WF_SLS_SIZE_MEMORY := "2.0"
WF_SLS_CHARGE_LABELS := "upptacka-api-find" #label should not contains blank
WF_SLS_LABELS := "upptacka-api-find" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM7 := "{${WF_SLS_REG_COMMON_110}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG7 := $(WF_SLS_REG_ITEM7)
WF_SLS_DOMAIN := github.com
WF_SLS_ORG_NAME := ingka-group-digital
WF_SLS_REPOS_NAME := "upptacka-monitoring"
WF_SLS_TYPE := "Org"  
WF_SLS_SIZE := "1"
WF_SLS_SIZE_CPU := "1.0"
WF_SLS_SIZE_MEMORY := "2.0"
WF_SLS_CHARGE_LABELS := "upptacka-monitoring" #label should not contains blank
WF_SLS_LABELS := "upptacka-monitoring" #label should not contains blank
WF_SLS_ORG_URL := https://${WF_SLS_DOMAIN}/${WF_SLS_ORG_NAME}
WF_SLS_REG_ITEM8 := "{${WF_SLS_REG_COMMON_110}, \"Type\":\"${WF_SLS_TYPE}\",\"Name\":\"${WF_SLS_ORG_NAME}\",\"Url\":\"${WF_SLS_ORG_URL}\",\"Size\":\"${WF_SLS_SIZE}\",\"Cpu\":\"${WF_SLS_SIZE_CPU}\",\"Memory\":\"${WF_SLS_SIZE_MEMORY}\",\"Repos\":\"${WF_SLS_REPOS_NAME}\",\"Labels\":\"${WF_SLS_LABELS}\",\"ChargeLabels\":\"${WF_SLS_CHARGE_LABELS}\"}"
REG8 := $(WF_SLS_REG_ITEM8)

WF_SLS_REGS := "[${REG3},${REG4},${REG5},${REG7},${REG8}]"