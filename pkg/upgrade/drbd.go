package upgrade

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	log "github.com/sirupsen/logrus"

	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/prompt"
)

func checkDrbdOptions(resDef client.ResourceDefinition) map[string][2]string {
	overrides := make(map[string][2]string)
	for key, targetValue := range linstorcontrol.DefaultResourceProps() {
		if resDef.Props[key] == targetValue {
			log.WithFields(log.Fields{
				"key":       key,
				"fromValue": resDef.Props[key],
			}).Debugf("DRBD option already correctly set")
			continue
		}
		fromValue := resDef.Props[key]
		log.WithFields(log.Fields{
			"key":       key,
			"fromValue": fromValue,
			"toValue":   targetValue,
		}).Debugf("Changing DRBD option")
		overrides[key] = [2]string{fromValue, targetValue}
	}
	return overrides
}

// upgradeDrbdOptions checks if the options of the given resource are current,
// and changes them if necessary. It returns a boolean indicating whether any
// changes were made, and an error, if any.
func upgradeDrbdOptions(ctx context.Context, linstor *client.Client, resource string, forceYes bool, dryRun bool) (bool, error) {
	resDef, err := linstor.ResourceDefinitions.Get(ctx, resource)
	if err != nil {
		return false, fmt.Errorf("failed to get resource definition: %w", err)
	}
	replaceOptions := checkDrbdOptions(resDef)
	if len(replaceOptions) == 0 {
		// nothing to do
		return false, nil
	}
	fmt.Println("The following resource options need to be changed:")
	colorCfg := renderer.ColorizedConfig{
		Column: renderer.Tint{
			Columns: []renderer.Tint{
				{},                                    // Property: no color
				{FG: renderer.Colors{color.FgRed}},   // Old Value: red
				{FG: renderer.Colors{color.FgGreen}}, // New Value: green
			},
		},
	}
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewColorized(colorCfg)),
	)
	table.Header("Property", "Old Value", "New Value")

	overrides := make(map[string]string, len(replaceOptions))
	for k, v := range replaceOptions {
		_ = table.Append([]string{k, v[0], v[1]})
		overrides[k] = v[1]
	}
	_ = table.Render()
	fmt.Println()
	if dryRun {
		return true, nil
	}
	if !forceYes {
		yes := prompt.Confirm("Change these options now?")
		if !yes {
			// abort
			return false, fmt.Errorf("aborted")
		}
	}
	err = linstor.ResourceDefinitions.Modify(ctx, resource, client.GenericPropsModify{
		OverrideProps: overrides,
	})
	if err != nil {
		return false, fmt.Errorf("failed to modify resource definition: %w", err)
	}
	return true, nil
}
