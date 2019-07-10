package poll

import (
	"bytes"
	"errors"
	"github.com/buexplain/go-watch/cmd/run/monitor"
	"github.com/buexplain/go-watch/logger"
	"github.com/radovskyb/watcher"
	"time"
)

type Poll struct {
	Info    monitor.Info
	watcher *watcher.Watcher
}

func New(info monitor.Info) *Poll {
	return &Poll{Info: info}
}

func (this *Poll) Init() error {
	this.watcher = watcher.New()
	//错误
	errs := bytes.Buffer{}
	//递归监视文件夹
	for _, v := range this.Info.Folder {
		if e := this.watcher.AddRecursive(v); e != nil {
			errs.WriteString(e.Error())
			errs.WriteByte('\n')
		}
	}
	//监视文件
	for _, v := range this.Info.Files {
		if e := this.watcher.Add(v); e != nil {
			errs.WriteString(e.Error())
			errs.WriteByte('\n')
		}
	}
	//返回错误
	if errs.Len() > 0 {
		return errors.New(errs.String())
	}
	return nil
}

func (this *Poll) Run() (<-chan monitor.Event, <-chan error, chan struct{}) {
	eventCH := make(chan monitor.Event)
	errorCH := make(chan error)
	closed := make(chan struct{})

	go func() {
		if err := this.watcher.Start(time.Millisecond * 100); err != nil {
			logger.ErrorF("启动监视器失败: %s\n", err)
			this.watcher.Close()
			close(closed)
		}
	}()

	go func() {
		for {
			select {
			case e := <-this.watcher.Event:
				event := monitor.Event{}
				event.Name = e.Name()
				if e.Op == watcher.Create {
					event.Op = monitor.Create
				} else if e.Op == watcher.Remove {
					event.Op = monitor.Remove
				} else if e.Op == watcher.Write {
					event.Op = monitor.Write
				} else if e.Op == watcher.Rename {
					event.Op = monitor.Rename
				} else if e.Op == watcher.Chmod {
					event.Op = monitor.Chmod
				} else {
					break
				}
				eventCH <- event
			case err := <-this.watcher.Error:
				errorCH <- err
			case <-this.watcher.Closed:
				close(closed)
				return
			}
		}
	}()

	return eventCH, errorCH, closed
}
