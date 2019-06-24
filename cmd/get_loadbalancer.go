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

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/pagination"
	myOpenstack "github.com/lingxiankong/openstackcli-go/pkg/openstack"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var lbID string

var getLoadBalancerCmd = &cobra.Command{
	Use:   "loadbalancer",
	Short: "Get all the underlying resources related to the load balancer(admin only)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lbID = args[0]
		osClient, err := myOpenstack.NewOpenStack(conf)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to initialize openstack client")
		}

		// vip
		lb, err := loadbalancers.Get(osClient.Octavia, lbID).Extract()
		if err != nil {
			log.WithFields(log.Fields{"error": err, "lbID": lbID}).Fatal("Failed to get the loadbalancer info")
		}
		fmt.Println(fmt.Sprintf("vip port: %s, IP: %s", lb.VipPortID, lb.VipAddress))

		// vip sg
		vipSgs, err := osClient.GetPortSecurityGroups(lb.VipPortID)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "portID": lb.VipPortID}).Fatal("Failed to get vip port security groups")
		}
		fmt.Println(fmt.Sprintf("\tsecurity groups: %s", vipSgs))

		// server group
		expectedName := fmt.Sprintf("octavia-lb-%s", lb.Name)
		err = servergroups.List(osClient.Nova).EachPage(func(page pagination.Page) (bool, error) {
			actual, err := servergroups.ExtractServerGroups(page)
			if err != nil {
				return false, err
			}

			for _, sg := range actual {
				if sg.Name == expectedName {
					fmt.Println(fmt.Sprintf("server group: %s", sg.ID))
					// return false to stop iteration.
					return false, nil
				}
			}

			return true, nil
		})
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to query server group")
		}

		// amphorae
		ams, err := osClient.GetLoadBalancerAmphorae(lbID)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "lbID": lbID}).Fatal("Failed to get amphorae")
		}

		fmt.Println("amphorae:")
		for _, am := range ams {
			fmt.Println(fmt.Sprintf("\t%s", am.ComputeID))
			fmt.Println(fmt.Sprintf("\t\tvrrp port: %s", am.VRRPPortID))

			// vrrp port sg
			sgs, err := osClient.GetPortSecurityGroups(am.VRRPPortID)
			if err != nil {
				log.WithFields(log.Fields{"error": err, "portID": am.VRRPPortID}).Fatal("Failed to get vrrp port security groups")
			}
			fmt.Println(fmt.Sprintf("\t\t\tsecurity groups: %s", sgs))
		}
	},
}

func init() {
	getCmd.AddCommand(getLoadBalancerCmd)
}
