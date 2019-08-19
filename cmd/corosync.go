package cmd

import (
	"net"
	"os"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var nodeIPs []net.IP
var clusterName string

const corotmpl = `totem {
 version: 2
 cluster_name: {{.Name}}
 secauth: off
 transport: udpu
}

nodelist {{"{"}}{{range $i, $v := .IPs}}
  node {
    ring0_addr: {{$v}}
    nodeid: {{inc $i}}
  }{{end}}
}

quorum {
  provider: corosync_votequorum
}

logging {
  to_logfile: yes
  logfile: /var/log/cluster/corosync.log
  to_syslog: yes
}`

// corosyncCmd represents the corosync command
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
		funcMap := template.FuncMap{
			"inc": func(i int) int {
				return i + 1
			},
		}
		t := template.Must(template.New("").Funcs(funcMap).Parse(corotmpl))
		type data struct {
			IPs  []net.IP
			Name string
		}

		t.Execute(os.Stdout, data{IPs: nodeIPs, Name: clusterName})
	},
}

func init() {
	rootCmd.AddCommand(corosyncCmd)

	corosyncCmd.ResetCommands()
	corosyncCmd.Flags().IPSliceVar(&nodeIPs, "ips", []net.IP{net.IPv4(127, 0, 0, 1)}, "comma seprated list of IPs (e.g., 1.2.3.4,1.2.3.5)")
	corosyncCmd.Flags().StringVar(&clusterName, "cluster-name", "mycluster", "name of the cluster")

	corosyncCmd.MarkPersistentFlagRequired("iqn")
	corosyncCmd.MarkPersistentFlagRequired("lun")
}
