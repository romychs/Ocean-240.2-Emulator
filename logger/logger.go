package logger

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var logFile *os.File = nil
var logBuffer *bufio.Writer = nil

type LogFormatter struct {
}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	caller := "-"
	if entry.HasCaller() {
		caller = filepath.Base(entry.Caller.File) + ":" + strconv.Itoa(entry.Caller.Line)
	}
	lvl := strings.ToUpper(entry.Level.String())
	if len(lvl) > 0 {
		lvl = lvl[0:1]
	} else {
		lvl = "-"
	}
	logMessage := fmt.Sprintf("%s[%s][%s]: %s\n", entry.Time.Format("2006-01-02T15:04:05"), lvl, caller, entry.Message) //time.RFC3339

	return []byte(logMessage), nil
}

// InitLogging Initialise logging format and level
func InitLogging() {
	log.SetFormatter(new(LogFormatter))
	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)
	log.SetOutput(os.Stdout)
}

// FlushLogs Flush logs if logging to file
func FlushLogs() {
	if logBuffer != nil {
		err := logBuffer.Flush()
		if err != nil {
			return
		}
	}
}

// ReconfigureLogging Reconfigure logging by config values
func ReconfigureLogging(config *config.ServiceConfig) {
	// redirect logging if log file specified
	if len(config.LogFile) > 0 {
		var err error
		logFile, err = os.OpenFile(config.LogFile, os.O_WRONLY|syscall.O_CREAT|syscall.O_APPEND, 0664)
		if err != nil {
			log.Fatalf("Can not access to log file: %s\n%v", config.LogFile, err)
		}
		logBuffer = bufio.NewWriter(logFile)
		log.SetOutput(logBuffer)
	}
	if len(config.LogLevel) > 0 {
		level, err := log.ParseLevel(config.LogLevel)
		if err != nil {
			log.Fatalf("Can not parse log level: %s\n%v", config.LogLevel, err)
		}
		log.SetLevel(level)
	}
}

// CloseLogs If log output is file, flush and close it
func CloseLogs() {
	FlushLogs()
	if logFile != nil {
		err := logFile.Close()
		if err != nil {
			log.Warnf("Can not close log file: %s\n%v", logFile.Name(), err)
		}
	}
}
