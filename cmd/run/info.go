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
	Args        []string
	Folder      []string
	Ext         []string
	Delay       uint
	Signal      int
	Timeout     int
	AutoRestart bool
}

func NewInfo() *Info {
	info := new(Info)
	info.Args = make([]string, 0)
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
	f := "Cmd:         %s\n"
	f += "Args:        %+v\n"
	f += "Folder:      %+v\n"
	f += "Ext:         %+v\n"
	f += "Delay:       %d\n"
	f += "Signal:      %s（%d）\n"
	f += "Timeout:     %d\n"
	f += "AutoRestart: %v"

	return fmt.Sprintf(f,
		this.Cmd,
		this.Args,
		this.Folder,
		this.Ext,
		this.Delay,
		syscall.Signal(this.Signal).String(), this.Signal,
		this.Timeout,
		this.AutoRestart)
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
	tmpArgs := make([]string, 0, len(this.Args))
	for _, v := range this.Args {
		v = strings.Trim(v, " ")
		if len(v) > 0 {
			tmpArgs = append(tmpArgs, v)
		}
	}
	this.Args = tmpArgs

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
