// Copyright © 2019 Lucas dos Santos Abreu <lucas.s.abreu@gmail.com>
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
	"strings"

	"github.com/lucassabreu/clockify-cli/cmd/completion"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	version   string
	commit    string
	buildDate string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "clockify-cli",
	Short:         "Allow to integrate with Clockify through terminal",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v, c, d string) {
	version = v
	commit = c
	buildDate = d

	if err := rootCmd.Execute(); err != nil {
		if viper.GetBool("debug") {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		os.Exit(1)
	}
}

const USER_ID_FLAG = "user-id"
const ENV_PREFIX = "CLOCKIFY"

func init() {
	cobra.OnInitialize(initConfig)

	viper.SetEnvPrefix(ENV_PREFIX)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.clockify-cli.yaml)")

	rootCmd.PersistentFlags().StringP(TOKEN, "t", "",
		"clockify's token (defaults to env $"+ENV_PREFIX+"_TOKEN)\n"+
			"\tCan be generated here: https://clockify.me/user/settings#generateApiKeyBtn",
	)
	_ = viper.BindPFlag(TOKEN, rootCmd.PersistentFlags().Lookup(TOKEN))

	rootCmd.PersistentFlags().StringP(WORKSPACE, "w", "", "workspace to be used (defaults to env $"+ENV_PREFIX+"_WORKSPACE)")
	_ = viper.BindPFlag(WORKSPACE, rootCmd.PersistentFlags().Lookup(WORKSPACE))
	_ = completion.AddSuggestionsToFlag(rootCmd, WORKSPACE, suggestWithClientAPI(suggestWorkspaces))

	rootCmd.PersistentFlags().StringP(USER_ID_FLAG, "u", "", "user id from the token (defaults to env $"+ENV_PREFIX+"_USER_ID)")
	_ = viper.BindPFlag(USER_ID, rootCmd.PersistentFlags().Lookup(USER_ID_FLAG))
	_ = completion.AddSuggestionsToFlag(rootCmd, USER_ID, suggestWithClientAPI(suggestUsers))

	rootCmd.PersistentFlags().Bool("debug", false, "show debug log (defaults to env $"+ENV_PREFIX+"_DEBUG)")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	rootCmd.PersistentFlags().BoolP(INTERACTIVE, "i", false,
		"will prompt you to confirm/complement commands input before executing the action (defaults to env $"+ENV_PREFIX+"_INTERACTIVE).\n"+
			"\tYou can be disable it temporally by setting it to 0 (-i=0 or "+ENV_PREFIX+"_INTERACTIVE=0)",
	)
	_ = viper.BindPFlag(INTERACTIVE, rootCmd.PersistentFlags().Lookup(INTERACTIVE))

	viper.RegisterAlias(ALLOW_NAME_FOR_ID, "allow-project-name")

	rootCmd.PersistentFlags().BoolP("allow-project-name", "", false, "allow use of project name when id is asked (defaults to env $"+ENV_PREFIX+"_ALLOW_PROJECT_NAME)")
	_ = rootCmd.Flags().MarkDeprecated("allow-project-name", "use `allow-name-for-id` flag instead")
	_ = viper.BindPFlag("allow-project-name", rootCmd.PersistentFlags().Lookup("allow-project-name"))

	rootCmd.PersistentFlags().BoolP(ALLOW_NAME_FOR_ID, "", false, "allow use of project/tag's name when id is asked (defaults to env $"+ENV_PREFIX+"_ALLOW_NAME_FOR_ID)")
	_ = viper.BindPFlag(ALLOW_NAME_FOR_ID, rootCmd.PersistentFlags().Lookup(ALLOW_NAME_FOR_ID))

	_ = viper.BindEnv(
		ALLOW_NAME_FOR_ID,
		ENV_PREFIX+"_ALLOW_NAME_FOR_ID",
		ENV_PREFIX+"_ALLOW_PROJECT_NAME",
	)

	_ = rootCmd.MarkFlagRequired(TOKEN)

	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		cmd.Println(err)
		cmd.Println(cmd.UsageString())
		return nil
	})
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

		// Search config in home directory with name ".clockify-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".clockify-cli")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.SetEnvPrefix("clockify")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			return
		default:
			fmt.Print(err)
			return
		}
	}
}
