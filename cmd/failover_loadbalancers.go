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
	"context"
	"errors"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	myOpenstack "github.com/lingxiankong/openstackcli-go/pkg/openstack"
	"github.com/lingxiankong/openstackcli-go/pkg/util"
)

var (
	parallelism int
	includeLBs  []string
	excludeLBs  []string
	timeout     int
)

var failoverLoadBalancersCmd = &cobra.Command{
	Use:   "loadbalancers",
	Short: "Failover load balancers in Octavia service(admin required)",
	Long: `Advantages of using this commmand over the "openstack loadbalancer failover":
	- Fail over the selected load balancers automatically.
	- Support different filtering conditions for load balancer selection e.g. by project, by inclusion or exclusion.
	- Support concurrency running.
	- Automatic skip if the load balancer has already been upgraded.
	- Automatic skip load balancers in invalid status and print information in warning log level.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if parallelism < 1 || parallelism > 6 {
			return errors.New("invalid --parallelism specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		osClient, err := myOpenstack.NewOpenStack(conf)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to initialize openstack client")
		}

		// Get the latest amphora image
		imageID, err := osClient.GetAmphoraImage()
		if err != nil {
			log.Fatalf("Failed to get latest amphora image: %v", err)
		}

		lbs, err := osClient.GetLoadbalancers(projectID)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to get load balancers.")
		}

		var validLBs []string

		// Find LBs that can be failed over, for LBs in invalid status, show the updated timestamp and skip.
		for _, lb := range lbs {
			if strings.Contains(lb.Name, "tempest") {
				log.WithFields(log.Fields{"loadbalancer": lb.ID}).Info("Created by tempest, skip")
				continue
			}

			if len(excludeLBs) > 0 && util.FindString(lb.ID, excludeLBs) {
				log.WithFields(log.Fields{"loadbalancer": lb.ID}).Info("excluded")
				continue
			}

			if len(includeLBs) == 0 || util.FindString(lb.ID, includeLBs) {
				if lb.ProvisioningStatus != "ACTIVE" && lb.ProvisioningStatus != "ERROR" {
					log.WithFields(log.Fields{"loadbalancer": lb.ID}).Warnf("Load balancer %s not in ACTIVE or ERROR, updated at %s, skipped", lb.Name, lb.UpdatedAt)
				} else {
					validLBs = append(validLBs, lb.ID)
				}
			}
		}

		if len(validLBs) == 0 {
			log.Info("No load balancers need to failover.")
			return
		}

		log.WithFields(log.Fields{"loadbalancers": validLBs}).Info("Will failover the load balancers.")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lbsCh := make(chan string)
		failCh := make(chan bool, parallelism)
		var waitgroup sync.WaitGroup

		// Catch signals
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			cancel()
			log.Info("Waiting for the existing operation to be finished...")
		}()

		// Fill the lbs need to failover into a channel
		go func(ch chan string, lbs []string) {
			for _, lb := range lbs {
				ch <- lb
			}
			close(ch)
		}(lbsCh, validLBs)

		// Create parallelism goroutines to handle all the lbs. If any of the goroutines fails, the whole process will stop.
		for i := 0; i < parallelism; i++ {
			waitgroup.Add(1)
			go func(ctx context.Context, ch <-chan string, failCh chan<- bool) {
				defer waitgroup.Done()

				for {
					select {
					case lbID, ok := <-ch:
						if !ok {
							return
						}

						log.WithFields(log.Fields{"loadbalancer": lbID}).Info("Starting failover load balancer")

						if err := osClient.FailoverLoadBalancer(lbID, imageID, timeout); err != nil {
							log.WithFields(log.Fields{"loadbalancer": lbID}).Errorf("Failed to failover load balancer: %v", err)
							failCh <- true
							return
						} else {
							log.WithFields(log.Fields{"loadbalancer": lbID}).Info("Finished to failover load balancer")
						}
					case <-ctx.Done():
						return
					}
				}
			}(ctx, lbsCh, failCh)
		}

		go func(failCh chan bool) {
			if <-failCh {
				cancel()
			}
		}(failCh)

		waitgroup.Wait()
	},
}

func init() {
	failoverLoadBalancersCmd.Flags().IntVar(&parallelism, "parallelism", 2, "Specifies the maximum desired number(1-5) of failover processes at any given time.")
	failoverLoadBalancersCmd.Flags().StringVar(&projectID, "project", "", "Only do failover for the load balancers belonging to the given project.")
	failoverLoadBalancersCmd.Flags().StringArrayVarP(&excludeLBs, "exclude-loadbalancers", "e", nil, "Load balancer IDs to ignore.")
	failoverLoadBalancersCmd.Flags().StringArrayVarP(&includeLBs, "include-loadbalancers", "i", nil, "Load balancer IDs to include.")
	failoverLoadBalancersCmd.Flags().IntVarP(&timeout, "timeout", "t", 600, "Timeout in seconds for the failover process.")

	failoverCmd.AddCommand(failoverLoadBalancersCmd)
}
