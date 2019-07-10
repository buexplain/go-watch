package factory

import (
	"errors"
	"github.com/buexplain/go-watch/cmd/run/monitor"
	"github.com/buexplain/go-watch/cmd/run/monitor/notify"
	"github.com/buexplain/go-watch/cmd/run/monitor/poll"
	"strings"
)

type Monitor interface {
	Init() error
	Run() (<-chan monitor.Event, <-chan error, chan struct{})
}

func New(pattern string, info monitor.Info) (Monitor, error) {
	if strings.EqualFold(pattern, "poll") {
		return poll.New(info), nil
	} else if strings.EqualFold(pattern, "notify") {
		return notify.New(info), nil
	}
	return nil, errors.New("监视模式必须是 poll 或 notify")
}
