package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yhy0/ChYing/conf/file"
	"github.com/yhy0/ChYing/pkg/desktop"
)

var certCmd = &cobra.Command{
	Use:   "cert",
	Short: "MITM CA 证书",
}

var certInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "把当前 Proxify CA 写入本机用户/系统信任库",
	RunE: func(cmd *cobra.Command, args []string) error {
		file.New()
		certPath := filepath.Join(file.ChyingDir, "proxify_data", "cacert.pem")
		result := desktop.InstallCACertificate(certPath)
		printStatus("%s", result.Message)
		if !result.Installed {
			return fmt.Errorf("%s", result.Message)
		}
		return nil
	},
}

func init() {
	certCmd.AddCommand(certInstallCmd)
}
