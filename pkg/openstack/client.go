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
	"github.com/gophercloud/gophercloud/openstack"
	log "github.com/sirupsen/logrus"
)

// OpenStack is an implementation of cloud provider Interface for OpenStack.
type OpenStack struct {
	keystone *gophercloud.ServiceClient
	Octavia  *gophercloud.ServiceClient
	Nova     *gophercloud.ServiceClient
	Neutron  *gophercloud.ServiceClient
	config   OpenStackConfig
}

// NewOpenStack gets openstack struct
func NewOpenStack(cfg OpenStackConfig) (*OpenStack, error) {
	provider, err := openstack.NewClient(cfg.AuthURL)
	if err != nil {
		return nil, err
	}

	if err = openstack.Authenticate(provider, cfg.ToAuthOptions()); err != nil {
		return nil, err
	}

	// get keystone admin client
	var keystone *gophercloud.ServiceClient
	keystone, err = openstack.NewIdentityV3(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to get keystone client: %v", err)
	}

	// get octavia service client
	var lb *gophercloud.ServiceClient
	lb, err = openstack.NewLoadBalancerV2(provider, gophercloud.EndpointOpts{
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find octavia endpoint for region %s: %v", cfg.Region, err)
	}

	// get neutron service client
	var network *gophercloud.ServiceClient
	network, err = openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find neutron endpoint for region %s: %v", cfg.Region, err)
	}

	// get nova service client
	var compute *gophercloud.ServiceClient
	compute, err = openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find compute v2 endpoint for region %s: %v", cfg.Region, err)
	}

	os := OpenStack{
		keystone: keystone,
		Octavia:  lb,
		Nova:     compute,
		Neutron:  network,
		config:   cfg,
	}

	log.Debug("openstack client initialized")

	return &os, nil
}
