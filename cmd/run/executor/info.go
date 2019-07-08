package executor

import (
	"runtime"
	"syscall"
)

type Info struct {
	Cmd         string
	Signal      int
	Timeout     int
	AutoRestart bool
}

func NewInfo() Info {
	info := Info{}
	info.Signal = int(syscall.SIGTERM)
	if runtime.GOOS == "windows" {
		info.Signal = int(syscall.SIGKILL)
	}
	info.Timeout = 5
	info.AutoRestart = false
	return info
}
