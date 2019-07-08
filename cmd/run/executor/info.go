package executor

import (
	"runtime"
	"syscall"
)

type Info struct {
	Cmd         string
	Args        []string
	Signal      int
	Timeout     int
	AutoRestart bool
}

func NewInfo() Info {
	info := Info{}
	info.Args = make([]string, 0)
	info.Signal = int(syscall.SIGTERM)
	if runtime.GOOS == "windows" {
		info.Signal = int(syscall.SIGKILL)
	}
	info.Timeout = 5
	info.AutoRestart = false
	return info
}
