/*
Package upgrade migrates existing resources to the latest version.
We will always try to upgrade to the most modern version of the drbd-reactor configuration.

In practice, this means the following DRBD options:

options {
   auto-promote no;
   quorum majority;
   on-suspended-primary-outdated force-secondary;
   on-no-quorum io-error;
}

And the following drbd-reactor settings:

[[promoter]]
  id = "iscsi-target1"
  [promoter.resources]
    [promoter.resources.target1]
      on-drbd-demote-failure = "reboot-immediate"
      runner = "systemd"
      start = [
        "ocf:heartbeat:Filesystem fs_cluster_private device=/dev/drbd/by-res/target1/0 directory=/srv/ha/internal/target1 fstype=ext4 run_fsck=no",
        "ocf:heartbeat:portblock pblock0 action=block ip=192.168.122.222 portno=3260 protocol=tcp",
        "ocf:heartbeat:IPaddr2 service_ip0 cidr_netmask=24 ip=192.168.122.222",
        "ocf:heartbeat:iSCSITarget target allowed_initiators='' incoming_password='' incoming_username='' iqn=iqn.2020-01.com.linbit:target1 portals=192.168.122.222:3260",
        "ocf:heartbeat:iSCSILogicalUnit lu1 lun=1 path=/dev/drbd/by-res/target1/1 product_id='LINSTOR iSCSI' scsi_sn=6391f15d target_iqn=iqn.2020-01.com.linbit:target1",
        "ocf:heartbeat:portblock portunblock0 action=unblock ip=192.168.122.222 portno=3260 protocol=tcp tickle_dir=/srv/ha/internal/target1",
      ]
      stop-services-on-exit = true
      target-as = "Requires"
*/
package upgrade

import (
	"bufio"
	"context"
	"fmt"
	"github.com/LINBIT/golinstor/client"
	"github.com/LINBIT/linstor-gateway/pkg/reactor"
	"github.com/google/go-cmp/cmp"
	"github.com/pelletier/go-toml"
	"github.com/sergi/go-diff/diffmatchpatch"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// confirm displays a prompt `s` to the user and returns a bool indicating yes / no
// If the lowercased, trimmed input begins with anything other than 'y', it returns false
func confirm(s string) bool {
	r := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [y/N]: ", s)

	res, err := r.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(strings.TrimSpace(res)) == "y"
}

func encode(cfg *reactor.PromoterConfig) (string, error) {
	buffer := strings.Builder{}
	encoder := toml.NewEncoder(&buffer).ArraysWithOneElementPerLine(true)

	err := encoder.Encode(&reactor.Config{Promoter: []reactor.PromoterConfig{*cfg}})
	if err != nil {
		return "", fmt.Errorf("error encoding toml: %w", err)
	}
	return buffer.String(), nil
}

func parseExistingConfig(ctx context.Context, linstor *client.Client, path string) (*reactor.PromoterConfig, *client.ResourceDefinition, []client.VolumeDefinition, []client.ResourceWithVolumes, error) {
	file, err := linstor.Controller.GetExternalFile(ctx, path)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch promoter config: %w", err)
	}

	var fullCfg reactor.Config
	err = toml.Unmarshal(file.Content, &fullCfg)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to decode promoter config: %w", err)
	}
	cfg := &fullCfg.Promoter[0]

	resourceDefinition, _, volumeDefinitions, resources, err := cfg.DeployedResources(ctx, linstor)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to fetch deployed resources: %w", err)
	}
	return cfg, resourceDefinition, volumeDefinitions, resources, nil
}

func maybeWriteNewConfig(ctx context.Context, linstor *client.Client, oldConfig *reactor.PromoterConfig, newConfig *reactor.PromoterConfig, forceYes, dryRun bool) (bool, error) {
	if cmp.Equal(oldConfig, newConfig) {
		// nothing to do
		return false, nil
	}
	oldToml, err := encode(oldConfig)
	if err != nil {
		return false, fmt.Errorf("failed to encode old promoter config: %w", err)
	}
	newToml, err := encode(newConfig)
	if err != nil {
		return false, fmt.Errorf("failed to marshal new config to toml: %w", err)
	}
	fmt.Println("The following configuration changes are necessary:")
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldToml, newToml, false)
	fmt.Println(dmp.DiffPrettyText(diffs))
	fmt.Println()
	if dryRun {
		return true, nil
	}
	if !forceYes {
		yes := confirm("Apply this configuration now?")
		if !yes {
			// abort
			return false, fmt.Errorf("aborted")
		}
	}
	err = reactor.EnsureConfig(ctx, linstor, newConfig)
	if err != nil {
		return true, fmt.Errorf("failed to install config: %w", err)
	}
	return true, nil
}
