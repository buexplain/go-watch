package run

import (
	"fmt"
	"github.com/buexplain/go-watch/logger"
	"os"
	"runtime"
	"strings"
	"syscall"
)

type Info struct {
	Cmd         string
	Args        []string
	Folder      []string
	Files       []string
	Delay       uint
	Signal      int
	Timeout     int
	AutoRestart bool
	Pattern     string
}

func NewInfo() *Info {
	info := new(Info)
	info.Args = make([]string, 0)
	info.Folder = make([]string, 0)
	info.Files = make([]string, 0)
	info.Delay = 1
	info.Signal = int(syscall.SIGTERM)
	info.Timeout = 5
	info.Pattern = "poll"

	if runtime.GOOS == "windows" {
		info.Signal = int(syscall.SIGKILL)
	}

	return info
}

func (this *Info) String() string {
	f := "Cmd:         %s\n"
	f += "Args:        %+v\n"
	f += "Folder:      %+v\n"
	f += "Files:       %+v\n"
	f += "Delay:       %d\n"
	f += "Signal:      %s（%d）\n"
	f += "Timeout:     %d\n"
	f += "AutoRestart: %v\n"
	f += "Pattern:     %s"

	return fmt.Sprintf(f,
		this.Cmd,
		this.Args,
		this.Folder,
		this.Files,
		this.Delay,
		syscall.Signal(this.Signal).String(), this.Signal,
		this.Timeout,
		this.AutoRestart,
		this.Pattern)
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

	//过滤文件夹
	tmpFolder := make([]string, 0, len(this.Folder))
	for _, v := range this.Folder {
		v = strings.Trim(strings.Replace(v, "\\", "/", -1), " ")
		if fi, err := os.Stat(v); err == nil && fi.IsDir() {
			tmpFolder = append(tmpFolder, v)
		}
	}
	this.Folder = tmpFolder

	//过滤文件
	tmpFiles := make([]string, 0, len(this.Folder))
	for _, v := range this.Files {
		v = strings.Trim(strings.Replace(v, "\\", "/", -1), " ")
		if fi, err := os.Stat(v); err == nil && !fi.IsDir() {
			tmpFiles = append(tmpFiles, v)
		}
	}
	this.Files = tmpFiles

	if len(this.Folder) == 0 && len(this.Files) == 0 {
		logger.Error("参数缺失：--folder or --files")
		return false
	}

	if !strings.EqualFold(this.Pattern, "poll") && !strings.EqualFold(this.Pattern, "notify") {
		logger.Error("监视模式必须是 poll 或 notify")
		return false
	}

	return true
}
