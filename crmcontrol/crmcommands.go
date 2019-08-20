package crmcontrol

// CRM (Pacemaker) commands

type crmCommand struct {
	executable string
	arguments  []string
}

const (
	crmUtility = "cibadmin"
)

// crmCreateCommand is the command for creating new resources.
var crmCreateCommand = crmCommand{crmUtility, []string{"--modify", "--allow-create", "--xml-pipe"}}

// crmUpdateCommand is the command for updating existing resources.
//
// Also used for deleting existing resources.
var crmUpdateCommand = crmCommand{crmUtility, []string{"--replace", "--xml-pipe"}}

// crmListCommand is the command for reading the CIB
var crmListCommand = crmCommand{crmUtility, []string{"--query"}}
