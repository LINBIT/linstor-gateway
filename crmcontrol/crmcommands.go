package crmcontrol

// CRM (Pacemaker) commands

type CrmCommand struct {
	executable string
	arguments  []string
}

const (
	CRM_UTILITY = "cibadmin"
)

// Command for creating new resources
var CRM_CREATE_COMMAND = CrmCommand{CRM_UTILITY, []string{"--modify", "--allow-create", "--xml-pipe"}}

// Command for updating existing resources; also used for deleting existing resources
var CRM_UPDATE_COMMAND = CrmCommand{CRM_UTILITY, []string{"--replace", "--xml-pipe"}}

// Command for reading the CIB
var CRM_LIST_COMMAND = CrmCommand{CRM_UTILITY, []string{"--query"}}
