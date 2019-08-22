package crmtemplate

const CRM_OBJ_NAMES = `p_iscsi_{{.CRM_TARGET_NAME}}_ip
p_pblock_{{.CRM_TARGET_NAME}}
p_iscsi_{{.CRM_TARGET_NAME}}
{{range .CRM_LUS}}
p_iscsi_{{$.CRM_TARGET_NAME}}_{{.}}
{{end}}
p_punblock_{{.CRM_TARGET_NAME}}`
