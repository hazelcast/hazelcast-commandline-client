package commands

import (
	"fmt"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const DefaultConfigFile = ".hzc.yaml"

var (
	cfgFile     string
	addresses   string
	clusterName string
	cloudToken  string
	rootCmd     = &cobra.Command{
		Use:   "hz-cli",
		Short: "Hazelcast command-line client",
		Long:  "Hazelcast command-line client connects your command-line to a Hazelcast cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is $HOME/%s)", DefaultConfigFile))
	rootCmd.PersistentFlags().StringVar(&addresses, "addr", "", "Addresses of the instances in the cluster.")
	rootCmd.PersistentFlags().StringVar(&clusterName, "cluster-group-name", "", "Name of the cluster that contains the instances.")
	rootCmd.PersistentFlags().StringVar(&cloudToken, "cloud-token", "", "Your Hazelcast Cloud token.")
	rootCmd.AddCommand(mapCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigName(DefaultConfigFile)
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
