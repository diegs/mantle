// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package systemd

import (
	"fmt"
	"strings"
	"time"

	"github.com/coreos/coreos-cloudinit/config"

	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/util"
)

var (
	gatewayconf = config.CloudConfig{
		CoreOS: config.CoreOS{
			Units: []config.Unit{
				config.Unit{
					Name:    "systemd-journal-gatewayd.socket",
					Command: "start",
				},
			},
		},
		Hostname: "gateway",
	}
)

func init() {
	register.Register(&register.Test{
		Run:         journalRemote,
		ClusterSize: 0,
		Name:        "systemd.journal.remote",
		UserData:    `#cloud-config`,
	})
}

// JournalRemote tests that systemd-journal-remote can read log entries from
// a systemd-journal-gatewayd server.
func journalRemote(c cluster.TestCluster) {
	// start gatewayd and log a message
	gateway, err := c.NewMachine(gatewayconf.String())
	if err != nil {
		c.Fatalf("Cluster.NewMachine: %s", err)
	}
	defer gateway.Destroy()

	// log a unique message on gatewayd machine
	msg := "supercalifragilisticexpialidocious"
	out, err := gateway.SSH("logger " + msg)
	if err != nil {
		c.Fatalf("logger: %v: %v", out, err)
	}

	// spawn a machine to read from gatewayd
	collector, err := c.NewMachine("#cloud-config")
	if err != nil {
		c.Fatalf("Cluster.NewMachine: %s", err)
	}
	defer collector.Destroy()

	// collect logs from gatewayd machine
	cmd := fmt.Sprintf("sudo systemd-run --unit systemd-journal-remote-client /usr/lib/systemd/systemd-journal-remote --url http://%s:19531", gateway.PrivateIP())
	out, err = collector.SSH(cmd)
	if err != nil {
		c.Fatalf("failed to start systemd-journal-remote: %v: %v", out, err)
	}

	// find the message on the collector
	journalReader := func() error {
		cmd = fmt.Sprintf("sudo journalctl _HOSTNAME=%s -t core --file /var/log/journal/remote/remote-%s.journal", gatewayconf.Hostname, gateway.PrivateIP())
		out, err = collector.SSH(cmd)
		if err != nil {
			return fmt.Errorf("journalctl: %v: %v", out, err)
		}

		if !strings.Contains(string(out), msg) {
			return fmt.Errorf("journal missing entry: expected %q got %q", msg, out)
		}

		return nil
	}

	if err := util.Retry(5, 2*time.Second, journalReader); err != nil {
		c.Fatal(err)
	}
}
