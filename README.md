# 监控你的程序文件变化并自动重启服务

## docker 编译
```bash
docker create -it -e GOPROXY=https://goproxy.io --name myGolang -v /opt:/root/src -w /root/src golang bash
docker start myGolang
docker exec -it myGolang bash
git clone https://github.com/buexplain/go-watch.git
cd go-watch
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gowatch main.go
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
      --autoRestart      是否自动重启子进程，子进程非守护类型不建议自动重启
      --cmd string       启动命令
      --delay uint       命令延迟执行秒数 (default 3)
      --ext strings      监听的文件的扩展
      --folder strings   监听的文件夹
  -h, --help             help for run
      --signal int       子进程关闭信号 (default 9)
      --timeout int      等待子进程关闭超时秒数 (default 10)
```

## License
[Apache-2.0](http://www.apache.org/licenses/LICENSE-2.0.html)