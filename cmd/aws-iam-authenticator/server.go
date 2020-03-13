/*
Copyright 2017 by the contributors.

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
	"flag"
	"fmt"
	"strings"

	"k8s.io/sample-controller/pkg/signals"
	"sigs.k8s.io/aws-iam-authenticator/pkg/mapper"
	"sigs.k8s.io/aws-iam-authenticator/pkg/server"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// DefaultPort is the default localhost port (chosen randomly).
	DefaultPort = 21362
	// Default Ec2 TPS Variables
	DefaultEC2DescribeInstancesQps   = 15
	DefaultEC2DescribeInstancesBurst = 5
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run a webhook validation server suitable that validates tokens using AWS IAM",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		stopCh := signals.SetupSignalHandler()

		cfg, err := getConfig()
		if err != nil {
			logrus.Fatalf("%s", err)
		}

		mappers := server.BuildMapperChain(cfg)
		for _, m := range mappers {
			logrus.Infof("starting mapper %q", m.Name())
			if err := m.Start(stopCh); err != nil {
				logrus.Fatalf("start mapper %q failed", m.Name())
			}
		}

		httpServer := server.New(cfg, mappers)
		httpServer.Run(stopCh)
	},
}

func init() {
	serverCmd.Flags().String("generate-kubeconfig",
		"/etc/kubernetes/aws-iam-authenticator/kubeconfig.yaml",
		"Output `path` where a generated webhook kubeconfig (for `--authentication-token-webhook-config-file`) will be stored (should be a hostPath mount).")
	viper.BindPFlag("server.generateKubeconfig", serverCmd.Flags().Lookup("generate-kubeconfig"))

	serverCmd.Flags().Bool("kubeconfig-pregenerated",
		false,
		"set to `true` when a webhook kubeconfig is pre-generated by running the `init` command, and therefore the `server` shouldn't unnecessarily re-generate a new one.")
	viper.BindPFlag("server.kubeconfigPregenerated", serverCmd.Flags().Lookup("kubeconfig-pregenerated"))

	serverCmd.Flags().String("state-dir",
		"/var/aws-iam-authenticator",
		"State `directory` for generated certificate and private key (should be a hostPath mount).")
	viper.BindPFlag("server.stateDir", serverCmd.Flags().Lookup("state-dir"))

	serverCmd.Flags().String("kubeconfig",
		"",
		"kubeconfig file path for using a local kubeconfig to configure the client to talk to the API server for the IAMIdentityMappings.")
	viper.BindPFlag("server.kubeconfig", serverCmd.Flags().Lookup("kubeconfig"))
	serverCmd.Flags().String("master",
		"",
		"master is the URL to the api server")
	viper.BindPFlag("server.master", serverCmd.Flags().Lookup("master"))

	serverCmd.Flags().String("address",
		"127.0.0.1",
		"IP Address to bind the server to listen to. (should be a 127.0.0.1 or 0.0.0.0)")
	viper.BindPFlag("server.address", serverCmd.Flags().Lookup("address"))

	serverCmd.Flags().StringSlice("backend-mode",
		[]string{mapper.ModeFile},
		fmt.Sprintf("Ordered list of backends to get mappings from. The first one that returns a matching mapping wins. Comma-delimited list of: %s", strings.Join(mapper.BackendModeChoices, ",")))
	viper.BindPFlag("server.backendMode", serverCmd.Flags().Lookup("backend-mode"))

	serverCmd.Flags().Int(
		"port",
		DefaultPort,
		"Port to bind the server to listen to")
	viper.BindPFlag("server.port", serverCmd.Flags().Lookup("port"))

	serverCmd.Flags().Int(
		"ec2-describeInstances-qps",
		DefaultEC2DescribeInstancesQps,
		"AWS EC2 rate limiting with qps")
	viper.BindPFlag("server.ec2DescribeInstancesQps", serverCmd.Flags().Lookup("ec2-describeInstances-qps"))

	serverCmd.Flags().Int(
		"ec2-describeInstances-burst",
		DefaultEC2DescribeInstancesBurst,
		"AWS EC2 rate Limiting with burst")
	viper.BindPFlag("server.ec2DescribeInstancesBurst", serverCmd.Flags().Lookup("ec2-describeInstances-burst"))

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	_ = fs.Parse([]string{})
	flag.CommandLine = fs

	rootCmd.AddCommand(serverCmd)
}
