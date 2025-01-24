package cmd

import (
	"fmt"
	"os"

	"github.com/danielunderwood/portctl/cmd/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "portctl",
	Short: "A tool to manage port forwarding",
	Long:  "A tool to manage port forwarding.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to portctl!")
		fmt.Printf("Config: %v\n", config.C)
	},
}

var configFile string

func init() {
	cobra.OnInitialize(func() {
		if configFile != "" {
			viper.SetConfigFile(configFile)
			// TODO: This shouldn't only be if there's no config file -- there
			// should be defaults. But some commands (e.g., one-shot forwarding)
			// don't need this and shouldn't error
			config.Init()
		}
	})
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file")
	// rootCmd.Flags().AddFlag(&pflag.Flag{
	// 	Name:      "config",
	// 	Shorthand: "c",
	// 	Usage:     "config file",
	// })

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(forwardCmd)
}

func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
