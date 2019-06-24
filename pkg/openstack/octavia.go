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

package openstack

import (
	"fmt"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/amphorae"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/loadbalancers"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/pools"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/lingxiankong/openstackcli-go/pkg/util"
)

// GetLoadbalancers get all the lbs.
func (os *OpenStack) GetLoadbalancers(project string) ([]loadbalancers.LoadBalancer, error) {
	opts := loadbalancers.ListOpts{}
	if project != "" {
		opts = loadbalancers.ListOpts{ProjectID: project}
	}

	allPages, err := loadbalancers.List(os.Octavia, opts).AllPages()
	if err != nil {
		return nil, err
	}

	allLoadbalancers, err := loadbalancers.ExtractLoadBalancers(allPages)
	if err != nil {
		return nil, err
	}

	return allLoadbalancers, nil
}

// GetPools retrives the pools belong to the loadbalancer. If isOrphan is true, only return shared pools in the
// loadbalancer. If listenerID is specified, return pools belong to that listener.
func (os *OpenStack) GetPools(lbID string, isOrphan bool, listenerID string) ([]pools.Pool, error) {
	if isOrphan {
		listenerID = ""
	}

	var lbPools []pools.Pool

	opts := pools.ListOpts{
		LoadbalancerID: lbID,
	}
	err := pools.List(os.Octavia, opts).EachPage(func(page pagination.Page) (bool, error) {
		ps, err := pools.ExtractPools(page)
		if err != nil {
			return false, err
		}
		for _, p := range ps {
			if isOrphan && len(p.Listeners) != 0 {
				continue
			}

			if listenerID != "" {
				var listenerIDs []string
				for _, l := range p.Listeners {
					listenerIDs = append(listenerIDs, l.ID)
				}

				if !util.FindString(listenerID, listenerIDs) {
					continue
				}
			}

			lbPools = append(lbPools, p)
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return lbPools, nil
}

// GetMembers retrieve all the members of the specified pool
func (os *OpenStack) GetMembers(poolID string) ([]pools.Member, error) {
	var members []pools.Member

	opts := pools.ListMembersOpts{}
	err := pools.ListMembers(os.Octavia, poolID, opts).EachPage(func(page pagination.Page) (bool, error) {
		v, err := pools.ExtractMembers(page)
		if err != nil {
			return false, err
		}
		members = append(members, v...)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return members, nil
}

// FailoverLoadBalancer fails over the specified load balancer and wait for the load balancer to be ACTIVE
func (os *OpenStack) FailoverLoadBalancer(id string, timeout int) error {
	if res := loadbalancers.Failover(os.Octavia, id); res.Err != nil {
		return res.Err
	}

	if err := os.WaitForLoadBalancerState(id, "ACTIVE", timeout); err != nil {
		return err
	}

	return nil
}

// WaitForLoadBalancerState will wait until a loadbalancer reaches a given state or ERROR.
func (os *OpenStack) WaitForLoadBalancerState(lbID, status string, secs int) error {
	return gophercloud.WaitFor(secs, func() (bool, error) {
		current, err := loadbalancers.Get(os.Octavia, lbID).Extract()
		if err != nil {
			if httpStatus, ok := err.(gophercloud.ErrDefault404); ok {
				if httpStatus.Actual == 404 {
					if status == "DELETED" {
						return true, nil
					}
				}
			}
			return false, err
		}

		if current.ProvisioningStatus == "ERROR" {
			return false, fmt.Errorf("loadbalancer %s goes to ERROR", lbID)
		}

		if current.ProvisioningStatus == status {
			return true, nil
		}

		return false, nil
	})
}

// GetLoadBalancerAmphorae return all the amphorae for a load balancer.
func (os *OpenStack) GetLoadBalancerAmphorae(id string) ([]amphorae.Amphora, error) {
	listOpts := amphorae.ListOpts{
		LoadbalancerID: id,
	}
	allPages, err := amphorae.List(os.Octavia, listOpts).AllPages()
	if err != nil {
		return nil, err
	}
	allAmphorae, err := amphorae.ExtractAmphorae(allPages)
	if err != nil {
		return nil, err
	}

	return allAmphorae, nil
}
