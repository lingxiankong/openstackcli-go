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
	"github.com/gophercloud/gophercloud"
)

// OpenStackConfig defines OpenStack credentials configuration.
type OpenStackConfig struct {
	Username    string
	Password    string
	ProjectName string `mapstructure:"project_name"`
	AuthURL     string `mapstructure:"auth_url"`
	Region      string
}

// ToAuthOptions gets openstack auth options
func (cfg OpenStackConfig) ToAuthOptions() gophercloud.AuthOptions {
	return gophercloud.AuthOptions{
		IdentityEndpoint: cfg.AuthURL,
		Username:         cfg.Username,
		Password:         cfg.Password,
		TenantName:       cfg.ProjectName,
		DomainName:       "default",
		AllowReauth:      true,
	}
}
