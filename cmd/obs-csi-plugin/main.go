/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/config"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/obs"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils/metadatas"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils/mounts"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/version"
)

var (
	endpoint    string
	cloudConfig string
)

//nolint:errcheck
func main() {
	flag.CommandLine.Parse([]string{})

	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: fmt.Sprintf("HuaweiCloud OBS CSI plugin %s", version.Version),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Glog requires this otherwise it complains.
			flag.CommandLine.Parse(nil)

			// This is a temporary hack to enable proper logging until upstream dependencies
			// are migrated to fully utilize klog instead of glog.
			klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
			klog.InitFlags(klogFlags)

			// Sync the glog and klog flags.
			cmd.Flags().VisitAll(func(f1 *pflag.Flag) {
				f2 := klogFlags.Lookup(f1.Name)
				if f2 != nil {
					value := f1.Value.String()
					f2.Value.Set(value)
				}
			})
		},
		Run: func(cmd *cobra.Command, args []string) {
			cloud, err := config.LoadConfig(cloudConfig)
			if err != nil {
				klog.V(3).Infof("Failed to load cloud config: %v", err)
			}

			klog.V(3).Infof("run obs csi driver")
			d := obs.NewDriver(endpoint, cloud)
			mount := mounts.GetMountProvider()
			metadata := metadatas.GetMetadataProvider(metadatas.MetadataID)
			mountClient := http.Client{
				Transport: &http.Transport{
					DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
						return net.Dial("unix", obs.SocketPath)
					},
				},
			}
			d.SetupDriver(mount, metadata, mountClient)
			d.Run()
		},
	}

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	cmd.PersistentFlags().StringVar(&endpoint, "endpoint", "unix://tmp/csi.sock", "CSI endpoint")

	cmd.PersistentFlags().StringVar(&cloudConfig, "cloud-config", "", "CSI driver cloud config")
	cmd.MarkPersistentFlagRequired("cloud-config")

	logs.InitLogs()
	defer logs.FlushLogs()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
