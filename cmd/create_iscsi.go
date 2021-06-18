package cmd

import (
	"net"
	"strconv"

	"github.com/LINBIT/gopacemaker/cib"
	"github.com/LINBIT/linstor-gateway/pkg/crmcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/iscsi"
	"github.com/LINBIT/linstor-gateway/pkg/linstorcontrol"
	"github.com/LINBIT/linstor-gateway/pkg/targetutil"
	"github.com/rck/unit"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func createISCSICommand() *cobra.Command {
	var ipCIDR string
	var controller net.IP
	var username, password, portals, group string
	var iqn string
	var lun int

	var sz *unit.Value
	var sizeKiB uint64
	var ip net.IP
	var ipnet *net.IPNet

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
linstor-iscsi create --iqn=iqn.2019-08.com.linbit:example --ip=192.168.122.181/24 \
 -username=foo --lun=1 --password=bar --resource-group=ssd_thin_2way --size=2G

Creates linstor resources example_lu0 and
pacemaker primitives p_iscsi_example_ip, p_iscsi_example, p_iscsi_example_lu0`,

		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			sizeKiB = uint64(sz.Value / unit.DefaultUnits["K"])

			var err error
			ip, ipnet, err = net.ParseCIDR(ipCIDR)
			if err != nil {
				log.Fatalf("Invalid service IP: %s", err.Error())
			}

			if portals == "" {
				portals = ip.String() + ":" + strconv.Itoa(iscsi.DFLT_ISCSI_PORTAL_PORT)
			}
		},

		Run: func(cmd *cobra.Command, args []string) {
			if !cmd.Flags().Changed("controller") {
				foundIP, err := crmcontrol.FindLinstorController()
				if err == nil { // it might be ok to not find it...
					controller = foundIP
				} else if err == cib.ErrCibFailed {
					log.Fatal(err)
				}
			}
			linstorCfg := linstorcontrol.Linstor{
				ControllerIP:      controller,
				ResourceGroupName: group,
				SizeKiB:           sizeKiB,
			}

			ones, _ := ipnet.Mask.Size()

			targetCfg := targetutil.TargetConfig{
				LUNs:             []*targetutil.LUN{&targetutil.LUN{ID: uint8(lun)}},
				IQN:              iqn,
				ServiceIP:        ip,
				ServiceIPNetmask: ones,
				Username:         username,
				Password:         password,
				Portals:          portals,
			}
			target := cliNewTargetMust(cmd, targetCfg)
			iscsiCfg := &iscsi.ISCSI{Linstor: linstorCfg, Target: target}
			err := iscsiCfg.CreateResource()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	createCmd.Flags().StringVarP(&iqn, "iqn", "i", "", "Set the iSCSI Qualified Name (e.g., iqn.2019-08.com.linbit:unique) (required)")
	createCmd.Flags().IntVarP(&lun, "lun", "l", 1, "Set the LUN Number (required)")
	createCmd.Flags().StringVar(&ipCIDR, "ip", "127.0.0.1/8", "Set the service IP and netmask of the target (required)")
	createCmd.Flags().IPVarP(&controller, "controller", "c", net.IPv4(127, 0, 0, 1), "Set the IP of the linstor controller node")
	createCmd.Flags().StringVar(&portals, "portals", "", "Set up portals, if unset, the service ip and default port")
	createCmd.Flags().StringVarP(&username, "username", "u", "", "Set the username (required)")
	createCmd.Flags().StringVarP(&password, "password", "p", "", "Set the password (required)")

	units := unit.DefaultUnits
	units["KiB"] = units["K"]
	units["MiB"] = units["M"]
	units["GiB"] = units["G"]
	units["TiB"] = units["T"]
	units["PiB"] = units["P"]
	units["EiB"] = units["E"]
	u := unit.MustNewUnit(units)
	sz = u.MustNewValue(1*units["G"], unit.None)
	createCmd.Flags().Var(sz, "size", "Set a size (e.g, 1TiB)")

	createCmd.Flags().StringVarP(&group, "resource-group", "g", "DfltRscGrp", "Set the LINSTOR resource-group")

	createCmd.MarkFlagRequired("ip")
	createCmd.MarkFlagRequired("username")
	createCmd.MarkFlagRequired("password")
	createCmd.MarkFlagRequired("iqn")
	createCmd.MarkFlagRequired("lun")

	return createCmd
}
