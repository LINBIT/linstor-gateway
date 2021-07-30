package cmd

import (
	"github.com/olekukonko/tablewriter"

	"github.com/LINBIT/linstor-gateway/pkg/common"
)

var (
	tableColorHeader   = tablewriter.Colors{tablewriter.FgBlueColor, tablewriter.Bold}
	tableColorOk       = tablewriter.Colors{tablewriter.FgGreenColor}
	tableColorDegraded = tablewriter.Colors{tablewriter.FgYellowColor}
	tableColorBad      = tablewriter.Colors{tablewriter.FgRedColor, tablewriter.Bold}
)

func ServiceStateColor(state common.ServiceState) tablewriter.Colors {
	switch state {
	case common.ServiceStateStarted:
		return tableColorOk
	case common.ServiceStateStopped:
		return tableColorBad
	default:
		return tableColorDegraded
	}
}

func ResourceStateColor(state common.ResourceState) tablewriter.Colors {
	switch state {
	case common.ResourceStateOK:
		return tableColorOk
	case common.ResourceStateDegraded:
		return tableColorDegraded
	case common.ResourceStateBad:
		return tableColorBad
	default:
		return tableColorDegraded
	}
}
