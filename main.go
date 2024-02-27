package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ohdat/inviteDemo/src/invite"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands

var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	// Uncomment the following line if your bare application
	// has an action associated with it:

	Run: func(cmd *cobra.Command, args []string) {
		invite.Invite() // Call the "invite" function
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

var CfgFile string

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	if CfgFile != "" {
		viper.SetConfigFile(CfgFile)
	} else {
		// Search config in home directory with name .config" (without extension).
		// 获取当前文件的路径
		currentPath, err := os.Executable()
		if err != nil {
			log.Fatalln(err)
		}
		currentDir := filepath.Dir(currentPath)
		viper.AddConfigPath(currentDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml") //设置配置文件类型，可选
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println("Using config file err:", err)
	}
}

func init() {
	cobra.OnInitialize(InitConfig)
	//The default should be the production environment.
	viper.SetDefault("environment", "production")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&CfgFile, "config", "", "config file (default is config.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
