package log

import (
	"fmt"
	"log"
	"os"
)

type LEVEL int

const (
	LEVEL_DEBUG LEVEL = iota
	LEVEL_INFO
	LEVEL_WARN
	LEVEL_ERROR
	LEVEL_NONE
)

var (
	logger *log.Logger
	level  = LEVEL_DEBUG
)

func init() {
	logger = log.New(os.Stdout, "", 0)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func SetLogger(l *log.Logger) {
	logger = l
}

func SetLevel(l LEVEL) {
	level = l
}

func logPrint(prefix string, v ...any) {
	var arr []interface{}
	arr = append(arr, prefix)
	arr = append(arr, v...)
	_ = logger.Output(3, fmt.Sprintln(arr...))
}

func Debug(f string, v ...interface{}) {
	if level <= LEVEL_DEBUG {
		logPrint("[DEBUG]", fmt.Sprintf(f, v...))
	}
}

func Info(f string, v ...interface{}) {
	if level <= LEVEL_INFO {
		logPrint("[INFO]", fmt.Sprintf(f, v...))
	}
}

func Warn(f string, v ...interface{}) {
	if level <= LEVEL_WARN {
		logPrint("[WARN]", fmt.Sprintf(f, v...))
	}
}

func Error(f string, v ...interface{}) {
	if level <= LEVEL_ERROR {
		logPrint("[ERROR]", fmt.Sprintf(f, v...))
	}
}
