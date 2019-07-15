# 监控你的程序文件变化并自动重启服务

## docker 编译
```bash
docker create -it -e GOPROXY=https://goproxy.io --name myGolang -v /opt:/root/src -w /root/src golang bash
docker start myGolang
docker exec -it myGolang bash
git clone https://github.com/buexplain/go-watch.git
cd go-watch
CGO_ENABLED=0 go build -o gowatch main.go
```

## 使用说明
```bash
chmod u+x gowatch
./gowatch run -h
```
```text
Usage:
   run [flags]

Flags:
      --args strings        启动命令所需参数
      --autoRestart         是否自动重启子进程，子进程非守护类型不建议自动重启
      --cmd string          启动命令
      --delay uint          命令延迟执行秒数 (default 1)
      --files strings       监视的文件
      --folder strings      监视的文件夹
  -h, --help                help for run
      --pattern string      监视文件变化的方式 poll 或 notify (default "poll")
      --preCmd string       预处理命令，启动命令执行前执行的命令
      --preCmdTimeout int   预处理命令执行超时时间 (default 10)
      --signal int          子进程关闭信号 (default 15)
      --timeout int         等待子进程关闭超时秒数 (default 5)
```

1. 参数输入格式，请参阅[cobra](https://github.com/spf13/cobra)库。
2. 如果操作系统不支持监视相关事件，或者是虚拟机与宿主机共享文件夹的情况下，那么只能使用 `--pattern=poll` 方式轮询监视文件。
3. 参数`--preCmd`的值只支持用`/bin/bash`执行。

## 使用示例
以PHP类项目[hyperf](https://github.com/hyperf-cloud/hyperf)为例子，使用方式如下：
```bash
./gowatch run --preCmd "php bin/hyperf.php di:init-proxy" --cmd "php" --args "/hyperf-skeleton/bin/hyperf.php, start" --files "/hyperf-skeleton/.env" --folder "/hyperf-skeleton/app/, /hyperf-skeleton/config/" --autoRestart=true
```

## License
[Apache-2.0](http://www.apache.org/licenses/LICENSE-2.0.html)