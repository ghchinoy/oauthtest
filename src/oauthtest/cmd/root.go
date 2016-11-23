// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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
	"oauthtest/helper"
	"oauthtest/oauthtestlib"
	"os"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	debug   bool
	profile string
	threads int
	max     int
	config  helper.Configuration
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "oauthtest",
	Short: "Calls an endpoint protected with OAuth 2 Client Credentials",
	Long: `A testing cli tool that calls an endpoint protected with OAuth 2 Client Credentials.
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if debug {
			fmt.Printf("%s %v clients\n", config.URI, len(config.Clients))
			fmt.Println("Threads", config.Threads)
			fmt.Println("Max", config.Max)
		}
		config = oauthtestlib.ObtainOAuthTokens(config, debug)
		oauthtestlib.MakeAPICalls(config, debug)
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// global flags
	// --config
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.akana/oauthtest.json)")
	// --debug
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "provide debug output")
	// --profile: default to default
	RootCmd.PersistentFlags().StringVar(&profile, "profile", "default", "set profile, must be in config")

	// set bash completion
	validConfigFilenames := []string{"json", "js"}
	RootCmd.PersistentFlags().SetAnnotation("config", cobra.BashCompFilenameExt, validConfigFilenames)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// For the command calld toggle
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	// --threads threads will override config file
	RootCmd.Flags().IntVar(&threads, "threads", 2, "threads to use")
	// --max max will override config file
	RootCmd.Flags().IntVar(&max, "max", 20, "max calls to make")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.SetConfigName("oauthtest")    // name of config file (without extension)
	viper.AddConfigPath("$HOME/.akana") // adding home directory as first search path
	viper.AutomaticEnv()                // read in environment variables that match

	if cfgFile != "" { // enable ability to specify config file via flag
		fmt.Println("configfile: ", cfgFile)
		viper.SetConfigFile(cfgFile)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if debug {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}

		profile, err := RootCmd.Flags().GetString("profile")
		if err != nil {
			fmt.Println("No profile chosen.")
			os.Exit(1)
		}
		if debug {
			fmt.Println("Using profile:", profile)
		}
		cfgMap := viper.GetStringMap(profile)
		if len(cfgMap) == 0 {
			fmt.Println("Profile doesn't exist:", profile)
			os.Exit(1)
		}

		err = mapstructure.Decode(cfgMap, &config)
		if err != nil {
			fmt.Println("Can't convert into config structure.")
			os.Exit(1)
		}

		max, err := RootCmd.Flags().GetInt("max")
		if err != nil {
			fmt.Println("Unable to determine max")
		}
		config.Max = max

		threads, err := RootCmd.Flags().GetInt("threads")
		if err != nil {
			fmt.Println("Unable to determine thread count")
		}
		config.Threads = threads

	} else {
		fmt.Println("Configuration file not found:", cfgFile)
		os.Exit(1)
	}
}
