package logger

import (
	"log"
	"os"
	"strings"
)

var Debug bool = false

var l *log.Logger

func init() {
	l = log.New(os.Stderr, "[DEBUG] ", log.LstdFlags)
}

func Printf(fmt string, args ...interface{}) {
	if Debug {
		if !strings.HasSuffix(fmt, "\n") {
			fmt += "\n"
		}
		log.Printf(fmt, args...)
	}
}
