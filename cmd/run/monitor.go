package run

import (
	"github.com/buexplain/go-watch/cmd/run/executor"
	"github.com/buexplain/go-watch/logger"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"time"
)

type Monitor struct {
	Info     *Info
	Delay    chan struct{}
	Executor *executor.Executor
}

func NewMonitor(executor *executor.Executor) *Monitor {
	return &Monitor{Info: info, Executor: executor, Delay: make(chan struct{}, 200)}
}

func (this *Monitor) delay() {
	isSend := false
	for {
		select {
		case <-this.Delay:
			isSend = true
			break
		case <-time.After(time.Duration(this.Info.Delay) * time.Second):
			if isSend {
				isSend = false
				logger.Info("监听到文件变化，重启子进程")
				this.Executor.Stop()
				this.Executor.Start()
			}
			break
		}
	}
}

func (this *Monitor) Run() bool {
	go this.delay()
Loop:
	if !info.Filter() {
		return false
	}

	watch, err := fsnotify.NewWatcher()
	if err != nil {
		logger.ErrorF("初始化监听器失败: %s\n", err)
		return false
	}

	defer func() {
		if watch != nil {
			_ = watch.Close()
		}
	}()

	for _, v := range this.Info.Folder {
		if err := filepath.Walk(v, func(path string, info os.FileInfo, err error) error {
			return watch.Add(path)
		}); err != nil {
			logger.ErrorF("监听器添加文件或者文件夹失败: %s\n", err)
			return false
		}
	}

	for {
		select {
		case event, ok := <-watch.Events:
			if !ok {
				logger.Info("监听器已经关闭")
				return true
			}
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					this.Delay <- struct{}{}
					_ = watch.Close()
					watch = nil
					goto Loop
				}
			}
			if info.IsTargetExt(event.Name) {
				this.Delay <- struct{}{}
			}
			break
		case err := <-watch.Errors:
			logger.ErrorF("监听器异常: %s\n", err)
			return false
		}
	}
}
