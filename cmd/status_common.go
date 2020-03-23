package cmd

import (
	"github.com/LINBIT/gopacemaker/cib"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/logrusorgru/aurora"
)

const (
	statusOk       = "✓"
	statusBad      = "✗"
	statusUnknown  = "?"
	statusDegraded = "!"
)

func stateToLongStatus(state cib.LrmRunState) string {
	return stateToColor(state)(state).String()
}

func linstorStateToLongStatus(state linstorcontrol.ResourceState) string {
	str := state.String()
	return linstorStateToColor(state)(str).String()
}

func stateToColor(state cib.LrmRunState) func(interface{}) aurora.Value {
	switch state {
	case cib.Running:
		return aurora.Green
	case cib.Stopped:
		return aurora.Red
	default:
		return aurora.Yellow
	}
}

func linstorStateToColor(state linstorcontrol.ResourceState) func(interface{}) aurora.Value {
	switch state {
	case linstorcontrol.OK:
		return aurora.Green
	case linstorcontrol.Degraded:
		return aurora.Yellow
	case linstorcontrol.Bad:
		return aurora.Red
	default:
		return aurora.Yellow
	}
}

func stateToStatus(state cib.LrmRunState) string {
	var str string
	switch state {
	case cib.Running:
		str = statusOk
	case cib.Stopped:
		str = statusBad
	default:
		str = statusUnknown
	}

	return stateToColor(state)(str).String()
}

func linstorStateToStatus(state linstorcontrol.ResourceState) string {
	var str string
	switch state {
	case linstorcontrol.OK:
		str = statusOk
	case linstorcontrol.Degraded:
		str = statusDegraded
	case linstorcontrol.Bad:
		str = statusBad
	default:
		str = statusUnknown
	}
	return linstorStateToColor(state)(str).String()
}

