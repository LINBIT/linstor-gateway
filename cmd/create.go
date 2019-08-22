package cmd

import (
	"net"
	"strconv"

	"github.com/LINBIT/linstor-iscsi/pkg/crmcontrol"
	"github.com/LINBIT/linstor-iscsi/pkg/iscsi"
	"github.com/LINBIT/linstor-iscsi/pkg/linstorcontrol"
	"github.com/rck/unit"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var ip net.IP
var username, password, size, portals, group string
var sizeKiB uint64
var ipChanged bool

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates an iSCSI target",
	Long: `Creates a highly available iSCSI target based on LINSTOR and Pacemaker.
At first it creates a new resouce within the linstor system, using the
specified resource group. The name of the linstor resources is derived
from the iqn and the lun number.
After that it creates resource primitives in the Pacemaker cluster including
all necessary order and location constraints. The Pacemaker primites are
prefixed with p_, contain the name and a resource type postfix.

For example:
linstor-iscsi create --iqn=iqn.2019-08.com.linbit:example --ip=192.168.122.181 \
 -username=foo --lun=0 --password=bar --resource_group=ssd_thin_2way --size=2G

Creates linstor resources example_lu0 and
pacemaker primitives p_iscsi_example_ip, p_iscsi_example, p_iscsi_example_lu0`,

	Args: cobra.NoArgs,
	PreRun: func(cmd *cobra.Command, args []string) {
		// TODO directly use for size, unit fulfills the flag interface.
		units := unit.DefaultUnits
		units["KiB"] = units["K"]
		units["MiB"] = units["M"]
		units["GiB"] = units["G"]
		units["TiB"] = units["T"]
		units["PiB"] = units["P"]
		units["EiB"] = units["E"]
		u := unit.MustNewUnit(units)

		v, err := u.ValueFromString(size)
		if err != nil {
			log.Fatal(err)
		}
		if v.Value < 0 {
			log.Fatal("Negative sizes are not allowed")
		}
		sizeKiB = uint64(v.Value / unit.DefaultUnits["K"])

		if portals == "" {
			portals = ip.String() + ":" + strconv.Itoa(iscsi.DFLT_ISCSI_PORTAL_PORT)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		if !cmd.Flags().Changed("controller") {
			foundIP, err := crmcontrol.FindLinstorController()
			if err == nil { // it might be ok to not find it...
				controller = foundIP
			}
		}
		linstorCfg := linstorcontrol.Linstor{
			Loglevel:          log.GetLevel().String(),
			ControllerIP:      controller,
			ResourceGroupName: group,
		}

		targetCfg := iscsi.TargetConfig{
			LUNs:      []*iscsi.LUN{&iscsi.LUN{ID: uint8(lun), SizeKiB: sizeKiB}},
			IQN:       iqn,
			ServiceIP: ip,
			Username:  username,
			Password:  password,
			Portals:   portals,
		}
		target := iscsi.NewTargetMust(targetCfg)
		iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: target}
		err := iscsiCfg.CreateResource()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().IPVar(&ip, "ip", net.IPv4(127, 0, 0, 1), "Set the service IP of the target (required)")
	createCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")
	createCmd.Flags().StringVar(&portals, "portals", "", "Set up portals, if unset, the service ip and default port")
	createCmd.Flags().StringVarP(&username, "username", "u", "", "Set the username (required)")
	createCmd.Flags().StringVarP(&password, "password", "p", "", "Set the password (required)")
	createCmd.Flags().StringVar(&size, "size", "1G", "Set the size (required)")
	createCmd.Flags().StringVarP(&group, "resource-group", "g", "default", "Set the LINSTOR resource-group")

	createCmd.MarkFlagRequired("ip")
	createCmd.MarkFlagRequired("username")
	createCmd.MarkFlagRequired("password")
	createCmd.MarkFlagRequired("nodes")
	createCmd.MarkFlagRequired("size")

	createCmd.MarkPersistentFlagRequired("iqn")
	createCmd.MarkPersistentFlagRequired("lun")
}
