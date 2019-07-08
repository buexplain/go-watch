package executor

import (
	"github.com/buexplain/go-watch/logger"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

//命令执行器
type Executor struct {
	//信号队列
	signal chan string
	//信号插队队列
	jump chan string
	//启动所需信息
	Info Info
	//是否停止
	kill chan struct{}
	//当前正在执行的子进程
	cmd *exec.Cmd
}

func NewExecutor(info Info) *Executor {
	return &Executor{Info: info, signal: make(chan string, 100), kill: make(chan struct{})}
}

func (this *Executor) Start() {
	this.signal <- SIGNAL_START
}

func (this *Executor) Stop() {
	this.signal <- SIGNAL_STOP
}

func (this *Executor) Kill() <-chan struct{} {
	this.signal <- SIGNAL_KILL
	return this.kill
}

func (this *Executor) Init() *Executor {
	if this.jump != nil {
		return this
	}
	go func() {
		if this.jump == nil {
			this.jump = make(chan string, 100)
		}
		f := func(s string) bool {
			switch s {
			case SIGNAL_START:
				this.start()
				<-time.After(3 * time.Second)
				return false
			case SIGNAL_STOP:
				this.stop()
				return false
			case SIGNAL_KILL:
				this.Info.AutoRestart = false
				this.stop()
				close(this.kill)
				return true
			default:
				logger.ErrorF("执行器信号错误：%s\n", s)
				return false
			}
		}
		for {
			select {
			case s := <-this.jump:
				f(s)
			default:
				select {
				case s := <-this.signal:
					if s == SIGNAL_CANCEL_OBSTRUCT {
						continue
					}
					if f(s) {
						return
					}
				}
			}
		}
	}()
	return this
}

func (this *Executor) chunk(timeout int) []int {
	result := make([]int, 0)
	c := 2
	exist := 0
	current := 1
	total := 0
	for {
		current *= c
		total += current
		if total > timeout {
			t := timeout - exist
			if t > 0 {
				result = append(result, t)
			}
			break
		}
		exist += current
		result = append(result, current)
	}
	return result
}

func (this *Executor) isExited() bool {
	if this.cmd == nil {
		return true
	}
	if this.cmd.ProcessState != nil {
		if runtime.GOOS == "linux" {
			//发送0信号，判断子进程是否存在
			if err := this.cmd.Process.Signal(syscall.Signal(0x0)); err != nil {
				return true
			}
		} else if this.cmd.ProcessState.Exited() {
			return true
		}
	}
	return false
}

func (this *Executor) stop() {
	if this.cmd != nil && this.cmd.Process != nil {
		if this.isExited() {
			this.cmd = nil
			return
		}
		//子进程存在，发送信号
		if err := this.cmd.Process.Signal(syscall.Signal(this.Info.Signal)); err == nil {
			logger.InfoF("子进程 %d 信号: %s\n", this.cmd.Process.Pid, syscall.Signal(this.Info.Signal).String())
			stopTimeout := this.chunk(this.Info.Timeout)
			for _, v := range stopTimeout {
				if this.isExited() {
					this.cmd = nil
					break
				}
				//等待一段时间
				<-time.After(time.Duration(v) * time.Second)
			}
			//如果子进程依然存在，强杀
			if this.isExited() == false {
				logger.InfoF("子进程 %d 信号: %s\n", this.cmd.Process.Pid, syscall.SIGKILL.String())
				if err := this.cmd.Process.Signal(syscall.Signal(syscall.SIGKILL)); err == nil {
					<-time.After(3 * time.Second)
					if this.isExited() {
						this.cmd = nil
					} else {
						//要么它死，要么我死
						os.Exit(1)
					}
				} else {
					logger.ErrorF("子进程 %d 信号失败: %s -- > %s\n", this.cmd.Process.Pid, syscall.SIGKILL.String(), err)
					//要么它死，要么我死
					os.Exit(1)
				}
			}
		} else {
			logger.ErrorF("子进程 %d 信号失败: %s -- > %s\n", this.cmd.Process.Pid, syscall.Signal(this.Info.Signal).String(), err)
		}
	}
}

func (this *Executor) start() {
	if this.cmd != nil {
		return
	}
	go func() {
		//新建一条命令
		this.cmd = exec.Command(this.Info.Cmd, this.Info.Args...)

		//管道关联命令标准输出失败
		stdOut, err := this.cmd.StdoutPipe()
		if err != nil {
			logger.ErrorF("子进程管道关联命令标准输出失败: %s\n", err)
			this.cmd = nil
			return
		}

		//管道关联命令标准错误输出失败
		strErr, err := this.cmd.StderrPipe()
		if err != nil {
			logger.ErrorF("子进程管道关联命令标准错误输出失败: %s\n", err)
			this.cmd = nil
			return
		}

		//启动命令
		if err := this.cmd.Start(); err != nil {
			logger.ErrorF("子进程启动失败: %s\n", err)
			this.cmd = nil
			return
		}

		logger.InfoF("子进程 %d 启动: %s\n", this.cmd.Process.Pid, this.Info.Cmd+" "+strings.Join(this.Info.Args, " "))

		//标准输出与标准错误输出管道go程结束控制器
		var pipeWaitGroup *sync.WaitGroup = &sync.WaitGroup{}
		//标准输出与标准错误输出管道go程结束时发出的信号，用于判断是否正常退出
		var pipeQuitCH chan bool = make(chan bool, 2)
		//标准输出与标准错误输出管道都结束的信号
		var pipeQuitCHExit chan struct{} = make(chan struct{})

		//读取命令标准输出管道
		pipeWaitGroup.Add(1)
		go func() {
			isError := false
			defer func() {
				pipeQuitCH <- isError
				pipeWaitGroup.Done()
			}()
			if _, err := io.Copy(os.Stdout, stdOut); err != nil {
				if err != io.EOF {
					logger.ErrorF("读取子进程命令标准输出管道失败: %s\n", err)
					isError = true
				}
			}
		}()

		//读取命令标准错误输出管道
		pipeWaitGroup.Add(1)
		go func() {
			isError := false
			defer func() {
				pipeQuitCH <- isError
				pipeWaitGroup.Done()
			}()
			if _, err := io.Copy(os.Stderr, strErr); err != nil {
				if err != io.EOF {
					logger.ErrorF("读取子进程命令标准错误输出管道失败: %s\n", err)
					isError = true
				}
			}
		}()

		//等待标准输出与标准错误输出管道go程结束
		go func() {
			pipeWaitGroup.Wait()
			//发出等待标准输出与标准错误输出管道go程结束信号
			close(pipeQuitCHExit)
		}()

		//监听标准输出与标准错误输出管道go程结束时发出的信号
		go func() {
			defer func() {
				close(pipeQuitCH)
				logger.Info("子进程管道相关go程全部退出")
			}()
			//停止子进程信号发送锁
			isSendSignal := false
			for {
				select {
				case <-pipeQuitCHExit:
					//标准输出与标准错误输出管道go程都结束了，结束当前go程
					return
				case isError, _ := <-pipeQuitCH:
					//标准输出或标准错误输出管道go程有异常结束，发送停止子进程信号
					if isError && !isSendSignal {
						logger.Info("管道异常，发出停止子进程信号")
						isSendSignal = true
						//发送到插队的队列里面
						this.jump <- SIGNAL_STOP
						//发出取消阻塞信号
						this.signal <- SIGNAL_CANCEL_OBSTRUCT
					}
				}
			}
		}()

		//等待子进程结束
		if err := this.cmd.Wait(); err != nil {
			logger.ErrorF("子进程 %d 停止异常: %s\n", this.cmd.Process.Pid, err)
		} else {
			logger.InfoF("子进程 %d 停止正常\n", this.cmd.Process.Pid)
		}
		//判断是否自动重启
		if this.Info.AutoRestart {
			logger.Info("发出重启子进程信号")
			//发出重启信号
			this.jump <- SIGNAL_STOP
			this.jump <- SIGNAL_START
			//发出取消阻塞信号
			this.signal <- SIGNAL_CANCEL_OBSTRUCT
		}
	}()
}
