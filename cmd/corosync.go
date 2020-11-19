package cmd

import (
	"net"

	corosync "github.com/LINBIT/gocorosync"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func corosyncCommand() *cobra.Command {
	var nodeIPs []net.IP
	var clusterName string

	var corosyncCmd = &cobra.Command{
		Use:   "corosync",
		Short: "Generates a corosync config",
		Long: `Generates a corosync config

For example:
linstor-iscsi corosync --ips="192.168.1.1,192.168.1.2"`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(nodeIPs) == 0 {
				log.Fatal("IP list is empty")
			}
			corosync.GenerateConfig(nodeIPs, clusterName)
		},
	}

	corosyncCmd.ResetCommands()
	corosyncCmd.Flags().IPSliceVar(&nodeIPs, "ips", []net.IP{net.IPv4(127, 0, 0, 1)}, "comma seprated list of IPs (e.g., 1.2.3.4,1.2.3.5)")
	corosyncCmd.Flags().StringVar(&clusterName, "cluster-name", "mycluster", "name of the cluster")

	corosyncCmd.MarkFlagRequired("iqn")
	corosyncCmd.MarkFlagRequired("lun")
	corosyncCmd.DisableAutoGenTag = true

	return corosyncCmd
}
