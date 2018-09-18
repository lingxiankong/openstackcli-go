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
	"os"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	myOpenstack "github.com/lingxiankong/openstackcli-go/pkg/openstack"
)

var (
	cfgFile string
	conf    myOpenstack.OpenStackConfig
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "osctl",
	Short: "A simple command line tool written in Go",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.runit.yaml)")
	rootCmd.PersistentFlags().StringVarP(&conf.Username, "user-name", "u", os.Getenv("OS_USERNAME"), "user name")
	rootCmd.PersistentFlags().StringVarP(&conf.Password, "password", "p", os.Getenv("OS_PASSWORD"), "user password")
	rootCmd.PersistentFlags().StringVarP(&conf.ProjectName, "project-name", "", os.Getenv("OS_PROJECT_NAME"), "project name")
	rootCmd.PersistentFlags().StringVarP(&conf.Region, "region", "r", os.Getenv("OS_REGION_NAME"), "region name")
	rootCmd.PersistentFlags().StringVarP(&conf.AuthURL, "authurl", "a", os.Getenv("OS_AUTH_URL"), "auth url")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".runit" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".runit")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to read config file")
	}

	log.WithFields(log.Fields{"file": viper.ConfigFileUsed()}).Debug("Using config file")

	if err := viper.Unmarshal(&conf); err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Unable to parse the configuration")
	}
}
