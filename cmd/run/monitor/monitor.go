package monitor

import (
	"bytes"
	"errors"
	"github.com/buexplain/go-watch/logger"
	"github.com/radovskyb/watcher"
	"os"
	"time"
)

type Monitor struct {
	Info    Info
	watcher *watcher.Watcher
}

func NewMonitor(info Info) *Monitor {
	return &Monitor{Info: info}
}

func (this *Monitor) Init() error {
	this.watcher = watcher.New()
	//错误
	err := bytes.Buffer{}
	//递归监视文件夹
	for _, v := range this.Info.Folder {
		if e := this.watcher.AddRecursive(v); e != nil {
			err.WriteString(e.Error())
			err.WriteByte('\n')
		}
	}
	//监视文件
	for _, v := range this.Info.Files {
		if e := this.watcher.Add(v); e != nil {
			err.WriteString(e.Error())
			err.WriteByte('\n')
		}
	}
	//返回错误
	if err.Len() > 0 {
		return errors.New(err.String())
	}
	return nil
}

func (this *Monitor) Run() (<-chan watcher.Event, <-chan error, <-chan struct{}) {
	go func() {
		if err := this.watcher.Start(time.Millisecond * 100); err != nil {
			logger.ErrorF("启动监视器失败: %s\n", err)
			os.Exit(1)
		}
	}()
	return this.watcher.Event, this.watcher.Error, this.watcher.Closed
}
