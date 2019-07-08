package run

import (
	"github.com/buexplain/go-watch/cmd"
	"github.com/buexplain/go-watch/cmd/run/executor"
	"github.com/buexplain/go-watch/logger"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

var Run *cobra.Command

var info *Info

func init() {
	info = NewInfo()
}

func init() {
	Run = &cobra.Command{
		Use: "run",
		Run: func(cmd *cobra.Command, args []string) {
			//检查参数
			if info.Filter() == false {
				os.Exit(1)
			}

			//打印参数信息
			logger.InfoF("参数信息: \n%s\n", info.String())

			//初始化执行器
			e := executor.NewExecutor(executor.Info{Cmd: info.Cmd, Signal: info.Signal, Timeout: info.Timeout, AutoRestart: info.AutoRestart}).Init()
			//启动子程序
			e.Start()

			//监听信号
			go func() {
				var signalCH chan os.Signal = make(chan os.Signal, 1)
				//监听kill默认信号
				signal.Notify(signalCH, syscall.SIGTERM)
				//收到信号
				<-signalCH
				//结束子进程
				<-e.Kill()
				//结束父进程
				os.Exit(0)
			}()

			//初始化文件监听器
			m := NewMonitor(e)

			//启动监听器
			if m.Run() {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
		},
	}

	//绑定参数
	Run.Flags().StringVar(&info.Cmd, "cmd", "", "启动命令")
	Run.Flags().StringSliceVar(&info.Folder, "folder", nil, "监听的文件夹")
	Run.Flags().StringSliceVar(&info.Ext, "ext", nil, "监听的文件的扩展")
	Run.Flags().UintVar(&info.Delay, "delay", info.Delay, "命令延迟执行秒数")
	Run.Flags().IntVar(&info.Signal, "signal", info.Signal, "子进程关闭信号")
	Run.Flags().IntVar(&info.Timeout, "timeout", info.Timeout, "等待子进程关闭超时秒数")
	Run.Flags().BoolVar(&info.AutoRestart, "autoRestart", info.AutoRestart, "是否自动重启子进程，子进程非守护类型不建议自动重启")
	Run.Flags()

	cmd.Root.AddCommand(Run)
}
