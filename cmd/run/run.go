package run

import (
	"github.com/buexplain/go-watch/cmd"
	"github.com/buexplain/go-watch/cmd/run/executor"
	"github.com/buexplain/go-watch/cmd/run/monitor"
	"github.com/buexplain/go-watch/cmd/run/monitor/factory"
	"github.com/buexplain/go-watch/logger"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
	"time"
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
			logger.InfoF("参数信息: \n%s\n\n", info.String())

			//初始化执行器
			eInfo := executor.Info{
				Cmd:           info.Cmd,
				Signal:        info.Signal,
				Timeout:       info.Timeout,
				AutoRestart:   info.AutoRestart,
				PreCmd:        info.PreCmd,
				PreCmdTimeout: info.PreCmdTimeout,
				PreCmdIgnoreError:info.PreCmdIgnoreError,
			}
			eInfo.Args = make([]string, len(info.Args))
			copy(eInfo.Args, info.Args)
			e := executor.NewExecutor(eInfo).Init()

			//监视信号
			go func() {
				var signalCH chan os.Signal = make(chan os.Signal, 1)
				//监视kill默认信号 和 Ctrl+C 发出的信号
				signal.Notify(signalCH, syscall.SIGTERM, syscall.SIGINT)
				//收到信号
				s := <-signalCH
				logger.InfoF("进程 %d 收到 %s 信号，开始停止子进程\n", os.Getpid(), s.String())
				//结束子进程
				<-e.Kill()
				//结束父进程
				os.Exit(0)
			}()

			//初始化文件监视器
			mInfo := monitor.Info{}
			mInfo.Folder = make([]string, len(info.Folder))
			copy(mInfo.Folder, info.Folder)
			mInfo.Files = make([]string, len(info.Files))
			copy(mInfo.Files, info.Files)
			m, err := factory.New(info.Pattern, mInfo)
			if err != nil {
				logger.ErrorF("初始化监视器失败: %s\n", err)
				os.Exit(1)
			}
			if err := m.Init(); err != nil {
				logger.ErrorF("初始化监视器失败: %s\n", err)
				os.Exit(1)
			}

			//启动子进程
			e.Start()

			//启动监视器
			eventCH, errorCH, closedCH := m.Run()
			send := false
			for {
				select {
				case _ = <-eventCH:
					logger.InfoF("进程 %d 监视到文件变化\n", os.Getpid())
					if !send {
						send = true
					}
				case err := <-errorCH:
					logger.ErrorF("监视器异常: %s\n", err)
				case <-closedCH:
					logger.Info("监视器已经关闭")
					<-e.Kill()
					os.Exit(0)
				case <-time.After(time.Duration(info.Delay) * time.Second):
					if send {
						logger.InfoF("进程 %d 重启子进程\n", os.Getpid())
						e.Stop()
						e.Start()
						send = false
					}
				}
			}
		},
	}

	//绑定参数
	Run.Flags().StringVar(&info.Cmd, "cmd", info.Cmd, "启动命令")
	Run.Flags().StringSliceVar(&info.Args, "args", info.Args, "启动命令所需参数")
	Run.Flags().StringSliceVar(&info.Folder, "folder", info.Folder, "监视的文件夹")
	Run.Flags().StringSliceVar(&info.Files, "files", info.Files, "监视的文件")
	Run.Flags().UintVar(&info.Delay, "delay", info.Delay, "命令延迟执行秒数")
	Run.Flags().IntVar(&info.Signal, "signal", info.Signal, "子进程关闭信号")
	Run.Flags().IntVar(&info.Timeout, "timeout", info.Timeout, "等待子进程关闭超时秒数")
	Run.Flags().BoolVar(&info.AutoRestart, "autoRestart", info.AutoRestart, "是否自动重启子进程，子进程非守护类型不建议自动重启")
	Run.Flags().StringVar(&info.PreCmd, "preCmd", info.PreCmd, "预处理命令，启动命令执行前执行的命令")
	Run.Flags().IntVar(&info.PreCmdTimeout, "preCmdTimeout", info.PreCmdTimeout, "预处理命令执行超时秒数")
	Run.Flags().BoolVar(&info.PreCmdIgnoreError, "preCmdIgnoreError", info.PreCmdIgnoreError, "是否忽略预处理命令的错误")
	Run.Flags().StringVar(&info.Pattern, "pattern", info.Pattern, "监视文件变化的方式 poll 或 notify")
	Run.Flags()

	cmd.Root.AddCommand(Run)
}
