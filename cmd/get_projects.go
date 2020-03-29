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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	myOpenstack "github.com/lingxiankong/openstackcli-go/pkg/openstack"
)

var getProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Get all projects ID and name(admin only).",
	Run: func(cmd *cobra.Command, args []string) {
		osClient, err := myOpenstack.NewOpenStack(conf)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to initialize openstack client")
		}

		projects, err := osClient.GetProjects()
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Fatal("Failed to get projects")
		}

		for _, p := range projects {
			fmt.Printf("ID: %s, Name: %s\n", p.ID, p.Name)
		}
	},
}

func init() {
	getCmd.AddCommand(getProjectsCmd)
}
