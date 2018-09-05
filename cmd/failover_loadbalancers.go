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
	"sync"

	"os"
	"os/signal"
	"syscall"

	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	myOpenstack "github.com/lingxiankong/openstackcli-go/pkg/openstack"
	"github.com/lingxiankong/openstackcli-go/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	parallelism int
	excludeLBs  []string
)

var failoverLoadBalancersCmd = &cobra.Command{
	Use:   "loadbalancers",
	Short: "Failover load balancers in Octavia service(admin required)",
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

		lbs, err := osClient.GetLoadbalancers(projectID)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to get load balancers.")
		}

		var notActiveLBs []string
		var validLBs []loadbalancers.LoadBalancer
		for _, lb := range lbs {
			if lb.ProvisioningStatus != "ACTIVE" {
				notActiveLBs = append(notActiveLBs, lb.ID)
			}
			if !util.FindString(lb.ID, excludeLBs) {
				validLBs = append(validLBs, lb)
			}
		}
		if len(notActiveLBs) != 0 {
			log.WithFields(log.Fields{"loadbalancers": notActiveLBs}).Fatal("Not all the load balancers are ACTIVE.")
		}
		if len(validLBs) == 0 {
			log.Info("No load balancers need to failover.")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lbsCh := make(chan loadbalancers.LoadBalancer)
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

		// Fill the lbs need to failover up to a channel
		go func(ch chan loadbalancers.LoadBalancer, lbs []loadbalancers.LoadBalancer) {
			for _, lb := range lbs {
				ch <- lb
			}
			close(ch)
		}(lbsCh, validLBs)

		// Create parallelism goroutines to handle all the lbs. If any of the goroutines fails, the whole process will stop.
		for i := 0; i < parallelism; i++ {
			waitgroup.Add(1)
			go func(ctx context.Context, ch <-chan loadbalancers.LoadBalancer, failCh chan<- bool) {
				defer waitgroup.Done()

				for {
					select {
					case lb, ok := <-ch:
						if !ok {
							return
						}

						log.WithFields(log.Fields{"loadbalancer": lb.ID}).Info("Start to failover load balancer")
						if err := osClient.FailoverLoadBalancer(lb.ID); err != nil {
							log.WithFields(log.Fields{"loadbalancer": lb.ID}).Error("Failed to failover load balancer")
							failCh <- true
							return
						} else {
							log.WithFields(log.Fields{"loadbalancer": lb.ID}).Info("Finished to failover load balancer")
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
	failoverLoadBalancersCmd.Flags().StringArrayVarP(&excludeLBs, "exclude-loadbalancers", "e", nil, "Ignore the load balancers.")

	failoverCmd.AddCommand(failoverLoadBalancersCmd)
}
