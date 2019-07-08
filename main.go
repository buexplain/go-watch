package main

import (
	"github.com/buexplain/go-watch/cmd"
	_ "github.com/buexplain/go-watch/cmd/run"
	_ "github.com/buexplain/go-watch/cmd/version"
	"github.com/buexplain/go-watch/logger"
	"os"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
