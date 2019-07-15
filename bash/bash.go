package bash

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var ErrCmdRunTimeout = errors.New("cmd run timeout")
var ErrCmdIsTerminated = errors.New("cmd is terminated")

//linux bash 命令对象
type Bash struct {
	cmd         string
	timeout     time.Duration
	terminateCH chan bool
	closed      chan struct{}
	command     *exec.Cmd
	stdout      bytes.Buffer
	stderr      bytes.Buffer
}

func NewBash(cmd string, timeout time.Duration) *Bash {
	b := new(Bash)
	b.cmd = cmd
	b.timeout = timeout
	if b.timeout <= 0*time.Second {
		b.timeout = 3600 * time.Second
	}
	b.terminateCH = make(chan bool)
	b.closed = make(chan struct{})
	b.command = exec.Command("/bin/bash", "-c", b.cmd)
	b.command.Stderr = &b.stderr
	b.command.Stdout = &b.stdout
	return b
}

//阻塞执行命令
func (this *Bash) Start() error {
	if err := this.command.Start(); err != nil {
		return err
	}

	errCH := make(chan error)

	go func() {
		defer func() {
			close(errCH)
		}()
		err := this.command.Wait()
		select {
		case <-this.closed:
			return
		default:
			errCH <- err
		}
	}()

	var err error

	select {
	case err = <-errCH:
	case <-time.After(this.timeout):
		err = this.terminate()
		if err == nil {
			err = ErrCmdRunTimeout
		}
	case <-this.terminateCH:
		err = this.terminate()
		if err == nil {
			err = ErrCmdIsTerminated
		}
	}

	close(this.closed)

	return err
}

func (this *Bash) terminate() error {
	return this.command.Process.Signal(syscall.SIGKILL)
}

//停止命令的执行
func (this *Bash) Stop() {
	select {
	case <-this.closed:
		return
	default:
		this.terminateCH <- true
	}
}

//检查程序是否有错误
//必须在 Start 方法结束后调用
func (this *Bash) HasErr() bool {
	//根据程序状态检查程序是否正常退出
	if this.command.ProcessState != nil {
		if this.command.ProcessState.Success() {
			return false
		}
		return true
	}
	//根据标准错误输出检查程序是否正常退出
	return this.stderr.Len() != 0
}

//返回标准错误输出信息
//必须在 Start 方法结束后调用
func (this *Bash) StdErr() string {
	return strings.TrimSpace(this.stderr.String())
}

//返回标准输出信息
//必须在 Start 方法结束后调用
func (this *Bash) StdOut() string {
	return strings.TrimSpace(this.stdout.String())
}
