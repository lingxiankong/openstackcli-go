// Copyright © 2020 Lingxian Kong <anlin.kong@gmail.com>
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
	"github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
)

// GetAmphoraImage gets latest amphora image ID.
func (os *OpenStack) GetAmphoraImage() (string, error) {
	listOpts := images.ListOpts{
		Limit: 1,
		Tags:  []string{"amphora"},
		Sort:  "created_at:desc",
	}
	allPages, err := images.List(os.Glance, listOpts).AllPages()
	if err != nil {
		return "", err
	}

	allImages, err := images.ExtractImages(allPages)
	if err != nil {
		return "", err
	}

	if len(allImages) == 0 {
		return "", fmt.Errorf("cannot find amphora image")
	}

	return allImages[0].ID, nil
}
