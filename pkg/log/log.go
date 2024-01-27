package log

import (
	"fmt"
	"log"
	"os"
)

const (
	LEVEL_NONE = iota
	LEVEL_ERROR
	LEVEL_WARN
	LEVEL_INFO
	LEVEL_DEBUG
)

var (
	logger *log.Logger
	Level  = LEVEL_DEBUG
)

func init() {
	logger = log.New(os.Stdout, "", 0)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func SetLogger(l *log.Logger) {
	logger = l
}

func SetLevel(l int) {
	Level = l
}

func logPrint(prefix string, v ...any) {
	var arr []interface{}
	arr = append(arr, prefix)
	arr = append(arr, v...)
	_ = logger.Output(3, fmt.Sprintln(arr...))
}

func Debug(f string, v ...interface{}) {
	if Level >= LEVEL_DEBUG {
		logPrint("[DEBUG]", fmt.Sprintf(f, v...))
	}
}

func Info(f string, v ...interface{}) {
	if Level >= LEVEL_INFO {
		logPrint("[INFO]", fmt.Sprintf(f, v...))
	}
}

func Warn(f string, v ...interface{}) {
	if Level >= LEVEL_WARN {
		logPrint("[WARN]", fmt.Sprintf(f, v...))
	}
}

func Error(f string, v ...interface{}) {
	if Level >= LEVEL_ERROR {
		logPrint("[ERROR]", fmt.Sprintf(f, v...))
	}
}
