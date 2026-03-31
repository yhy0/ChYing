package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yhy0/ChYing/conf"
)

var rootCmd = &cobra.Command{
	Use:   "chying-cli",
	Short: "ChYing CLI - 被动扫描服务",
	Long:  "ChYing 被动扫描服务，提供 HTTP 代理 + 被动扫描 + MCP 接口。",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ChYing CLI v%s\n", conf.Version)
	},
}

func main() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
