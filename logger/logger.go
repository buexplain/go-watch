package logger

import (
	"log"
	"os"
)

var stdOut = log.New(os.Stdout, "", log.LstdFlags)

func Error(v ...interface{}) {
	log.Println(v...)
}

func ErrorF(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func Info(v ...interface{}) {
	stdOut.Println(v...)
}

func InfoF(format string, v ...interface{}) {
	stdOut.Printf(format, v...)
}
