package cmd

import (
	"github.com/fatih/color"

	"github.com/LINBIT/linstor-gateway/pkg/common"
)

var (
	colorHeader   = color.New(color.FgBlue, color.Bold).SprintFunc()
	colorOk       = color.New(color.FgGreen).SprintFunc()
	colorDegraded = color.New(color.FgYellow).SprintFunc()
	colorBad      = color.New(color.FgRed, color.Bold).SprintFunc()
)

func ColorServiceState(state common.ServiceState, s string) string {
	switch state {
	case common.ServiceStateStarted:
		return colorOk(s)
	case common.ServiceStateStopped:
		return colorBad(s)
	default:
		return colorDegraded(s)
	}
}

func ColorResourceState(state common.ResourceState, s string) string {
	switch state {
	case common.ResourceStateOK:
		return colorOk(s)
	case common.ResourceStateDegraded:
		return colorDegraded(s)
	case common.ResourceStateBad:
		return colorBad(s)
	default:
		return colorDegraded(s)
	}
}
