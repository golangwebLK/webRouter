package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	//log.Logger写文件支持并发
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	logLevel    = 0 //默认的LogLevel为0，即所有级别的日志都打印

	logOut        *os.File
	day           int
	dayChangeLock sync.RWMutex
	logFile       string
)

const (
	DebugLevel = iota //iota=0
	InfoLevel         //InfoLevel=iota, iota=1
	WarnLevel         //WarnLevel=iota, iota=2
	ErrorLevel        //ErrorLevel=iota, iota=3
)

func SetLogLevel(level int) {
	logLevel = level
}

func SetLogFile(file string) {
	logFile = file
	now := time.Now()
	var err error
	if logOut, err = os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664); err != nil {
		panic(err)
	} else {
		debugLogger = log.New(logOut, "[DEBUG] ", log.LstdFlags)
		infoLogger = log.New(logOut, "[INFO] ", log.LstdFlags)
		warnLogger = log.New(logOut, "[WARN] ", log.LstdFlags)
		errorLogger = log.New(logOut, "[ERROR] ", log.LstdFlags)
		day = now.YearDay()
		dayChangeLock = sync.RWMutex{}
	}
}

// 检查是否需要切换日志文件，如果需要则切换
func checkAndChangeLogfile() {
	dayChangeLock.Lock()
	defer dayChangeLock.Unlock()
	now := time.Now()
	if now.YearDay() == day {
		return
	}

	logOut.Close()
	postFix := now.Add(-24 * time.Hour).Format("20060102") //昨天的日期
	if err := os.Rename(logFile, logFile+"."+postFix); err != nil {
		fmt.Printf("append date postfix %s to log file %s failed: %v\n", postFix, logFile, err)
		return
	}
	var err error
	if logOut, err = os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664); err != nil {
		fmt.Printf("create log file %s failed %v\n", logFile, err)
		return
	} else {
		debugLogger = log.New(logOut, "[DEBUG] ", log.LstdFlags)
		infoLogger = log.New(logOut, "[INFO] ", log.LstdFlags)
		warnLogger = log.New(logOut, "[WARN] ", log.LstdFlags)
		errorLogger = log.New(logOut, "[ERROR] ", log.LstdFlags)
		day = now.YearDay()
	}
}

func addPrefix() string {
	file, _, line := getLineNo()
	arr := strings.Split(file, "/")
	if len(arr) > 3 {
		arr = arr[len(arr)-3:]
	}
	return strings.Join(arr, "/") + ":" + strconv.Itoa(line)
}

func Debug(format string, v ...any) {
	if logLevel <= DebugLevel {
		checkAndChangeLogfile()
		debugLogger.Printf(addPrefix()+" "+format, v...)
	}
}

func Info(format string, v ...any) {
	if logLevel <= InfoLevel {
		checkAndChangeLogfile()
		infoLogger.Printf(addPrefix()+" "+format, v...) //format末尾如果没有换行符会自动加上
	}
}

func Warn(format string, v ...any) {
	if logLevel <= WarnLevel {
		checkAndChangeLogfile()
		warnLogger.Printf(addPrefix()+" "+format, v...)
	}
}

func Error(format string, v ...any) {
	if logLevel <= ErrorLevel {
		checkAndChangeLogfile()
		errorLogger.Printf(addPrefix()+" "+format, v...)
	}
}

func getLineNo() (string, string, int) {
	funcName, file, line, ok := runtime.Caller(3)
	if ok {
		return file, runtime.FuncForPC(funcName).Name(), line
	} else {
		return "", "", 0
	}
}
