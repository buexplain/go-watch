package run

import (
	"fmt"
	"github.com/buexplain/go-watch/logger"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

type Info struct {
	Cmd         string
	Folder      []string
	Ext         []string
	Delay       uint
	Signal      int
	Timeout     int
	AutoRestart bool
}

func NewInfo() *Info {
	info := new(Info)
	info.Folder = make([]string, 0)
	info.Ext = make([]string, 0)
	info.Delay = 3
	info.Signal = int(syscall.SIGTERM)
	info.Timeout = 10

	if runtime.GOOS == "windows" {
		info.Signal = int(syscall.SIGKILL)
	}

	return info
}

func (this *Info) String() string {
	f := "Cmd:     %s\n"
	f += "Folder:  %+v\n"
	f += "Ext:     %+v\n"
	f += "Delay:   %d\n"
	f += "Signal:  %s（%d）\n"
	f += "Timeout: %d\n"

	return fmt.Sprintf(f,
		info.Cmd,
		info.Folder,
		info.Ext,
		info.Delay,
		syscall.Signal(info.Signal).String(), info.Signal,
		info.Timeout)
}

func (this *Info) IsTargetExt(s string) bool {
	if len(this.Ext) == 0 {
		return true
	}
	e := filepath.Ext(s)
	for _, v := range this.Ext {
		if strings.EqualFold(e, v) {
			return true
		}
	}
	return false
}

func (this *Info) Filter() bool {
	//过滤命令
	this.Cmd = strings.Trim(this.Cmd, " ")
	if this.Cmd == "" {
		logger.Error("参数缺失：--cmd")
		return false
	}

	//过滤扩展
	for k, v := range this.Ext {
		this.Ext[k] = "." + strings.TrimLeft(strings.Trim(v, ". "), ".")
	}

	//过滤文件夹
	tmpFolder := make([]string, 0, len(this.Folder))
	for _, v := range this.Folder {
		v = strings.Trim(strings.Replace(v, "\\", "/", -1), " ")
		if fi, err := os.Stat(v); err == nil && fi.IsDir() {
			tmpFolder = append(tmpFolder, v)
		}
	}
	this.Folder = tmpFolder
	if len(this.Folder) == 0 {
		logger.Error("参数缺失：--folder")
		return false
	}

	return true
}
