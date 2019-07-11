package bash

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

//linux bash 命令对象
type Bash struct {
	cmd         string
	timeout     time.Duration
	terminateCh chan bool
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
	b.terminateCh = make(chan bool)
	b.command = exec.Command("/bin/bash", "-c", b.cmd)
	b.command.Stderr = &b.stderr
	b.command.Stdout = &b.stdout
	return b
}

func (this *Bash) Start() error {
	if err := this.command.Start(); err != nil {
		return err
	}

	errCH := make(chan error)
	go func() {
		defer close(errCH)
		errCH <- this.command.Wait()
	}()

	var err error

	select {
	case err = <-errCH:
	case <-time.After(this.timeout):
		err = this.terminate()
		if err == nil {
			err = errors.New(fmt.Sprintf("cmd run timeout time[%v]", this.timeout))
		}
	case <-this.terminateCh:
		err = this.terminate()
		if err == nil {
			err = errors.New(fmt.Sprintf("cmd is terminated"))
		}
	}

	return err
}

func (this *Bash) terminate() error {
	return this.command.Process.Signal(syscall.SIGKILL)
}

func (this *Bash) Stop() {
	this.terminateCh <- true
}

func (this *Bash) HasErr() bool {
	return this.stderr.Len() != 0
}

func (this *Bash) StdErr() string {
	return strings.TrimSpace(this.stderr.String())
}

func (this *Bash) StdOut() string {
	return strings.TrimSpace(this.stdout.String())
}
