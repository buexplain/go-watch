package cmd

import (
	"github.com/buexplain/go-watch/logger"
	"github.com/spf13/cobra"
	"os"
)

var Root *cobra.Command

func init() {
	Root = &cobra.Command{
		Use: "",
		Run: func(cmd *cobra.Command, args []string) {
			logger.InfoF("监控你的程序文件变化并自动重启服务器，使用帮助: %s --help", os.Args[0])
		},
	}
}
