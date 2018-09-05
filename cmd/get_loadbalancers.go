// Copyright © 2018 Lingxian Kong <anlin.kong@gmail.com>
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

package cmd

import (
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
	myOpenstack "github.com/lingxiankong/openstackcli-go/pkg/openstack"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var projectID string

var getLoadBalancersCmd = &cobra.Command{
	Use:   "loadbalancers",
	Short: "Get all the load balancers and the sub-resources(listeners, pools, members, etc.).",
	Run: func(cmd *cobra.Command, args []string) {
		osClient, err := myOpenstack.NewOpenStack(conf)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to initialize openstack client")
		}

		lbs, err := osClient.GetLoadbalancers(projectID)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to get load balancers.")
		}

		for _, lb := range lbs {
			var lbInfoList []string
			lbInfoList = append(lbInfoList, fmt.Sprintf("- LoadBalancer: %s", lb.ID), fmt.Sprintf("status: %s", lb.ProvisioningStatus), fmt.Sprintf("vip: %s", lb.VipAddress))
			if lb.Name != "" {
				lbInfoList = append(lbInfoList, fmt.Sprintf("name: %s", lb.Name))
			}
			fmt.Println(strings.Join(lbInfoList, ", "))

			for _, listener := range lb.Listeners {
				listenerInfo, err := listeners.Get(osClient.Octavia, listener.ID).Extract()
				if err != nil {
					log.WithFields(log.Fields{"error": err, "loadbalancer": lb.ID, "listener": listener.ID}).Fatal("Failed to get listener.")
				}

				listenerLine := fmt.Sprintf("\t- Listener: %s, protocol: %s, port: %d", listenerInfo.ID, listenerInfo.Protocol, listenerInfo.ProtocolPort)
				if listenerInfo.Name != "" {
					listenerLine += fmt.Sprintf(", name: %s", listenerInfo.Name)
				}
				fmt.Println(listenerLine)

				// Get listener pools, pools can only be retrieved by loadbalancer rather than listener.
				listenerPools, err := osClient.GetPools(lb.ID, false, listener.ID)
				if err != nil {
					log.WithFields(log.Fields{"error": err, "loadbalancer": lb.ID, "listener": listener.ID}).Fatal("Failed to get listener pools.")
				}

				for _, pool := range listenerPools {
					fmt.Printf("\t\t- Pool: %s, protocol: %s\n", pool.ID, pool.Protocol)

					// Get pool members
					members, err := osClient.GetMembers(pool.ID)
					if err != nil {
						log.WithFields(log.Fields{"error": err, "loadbalancer": lb.ID, "pool": pool.ID}).Fatal("Failed to get members.")
					}

					for _, m := range members {
						fmt.Printf("\t\t\t- Member: %s, address: %s, port: %d\n", m.ID, m.Address, m.ProtocolPort)
					}
				}
			}

			// Get shared pools
			sharedPools, err := osClient.GetPools(lb.ID, true, "")
			if err != nil {
				log.WithFields(log.Fields{"error": err, "loadbalancer": lb.ID}).Fatal("Failed to get shared pools.")
			}

			for _, pool := range sharedPools {
				fmt.Printf("\t- Pool: %s, protocol: %s\n", pool.ID, pool.Protocol)

				// Get pool members
				members, err := osClient.GetMembers(pool.ID)
				if err != nil {
					log.WithFields(log.Fields{"error": err, "loadbalancer": lb.ID, "pool": pool.ID}).Fatal("Failed to get members.")
				}

				for _, m := range members {
					fmt.Printf("\t\t- Member: %s, address: %s, port: %d\n", m.ID, m.Address, m.ProtocolPort)
				}
			}
		}
	},
}

func init() {
	getLoadBalancersCmd.Flags().StringVar(&projectID, "project", "", "Only get loadbalancer resources for the given project(admin required).")
	getCmd.AddCommand(getLoadBalancersCmd)
}
