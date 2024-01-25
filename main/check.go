package main

import (
	"github.com/anhk/mtun/pkg/log"
	"github.com/spf13/cobra"
	"os"
	"os/user"
)

// 检查当前用户是否以Root权限运行
func checkIsRoot(cmd *cobra.Command, args []string) {
	u, _ := user.Current()
	if u == nil || u.Username != "root" {
		log.Error("please run as root")
		os.Exit(0)
	}
}
