package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	// LogFile is the file path of generated logs
	LogFile            string = "runtime.log"
	VerboseDescription string = `
	// Verbose control level of log verbosity
	// 0) Only Info, Error
	// 1) Add Debug, Warning
	// 2) Add Trace
`
	Verbose int = 0
)

type Logger struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
}

func NewLogger(prefix string) (logger *Logger) {
	logger = new(Logger)
	stdoutWriters := []io.Writer{os.Stdout}
	stderrWriters := []io.Writer{os.Stderr}
	if LogFile != "" {
		file, err := os.OpenFile(LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open error log file:", err)
		}
		stdoutWriters = append(stdoutWriters, file)
		stderrWriters = append(stderrWriters, file)
	}

	logger.Info = log.New(io.MultiWriter(stdoutWriters...),
		fmt.Sprintf("[%s INFO]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Error = log.New(io.MultiWriter(stderrWriters...),
		fmt.Sprintf("[%s ERROR]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Trace = log.New(io.MultiWriter(stdoutWriters...),
		fmt.Sprintf("[%s TRACE]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Warning = log.New(io.MultiWriter(stdoutWriters...),
		fmt.Sprintf("[%s WARNING]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Debug = log.New(os.Stdout,
		fmt.Sprintf("[%s DEBUG]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	if Verbose < 1 {
		logger.Warning.SetOutput(ioutil.Discard)
		logger.Debug.SetOutput(ioutil.Discard)
	}
	if Verbose < 2 {
		logger.Trace.SetOutput(ioutil.Discard)
	}
	return
}

func (logger *Logger) DiscardAll() {
	logger.Info.SetOutput(ioutil.Discard)
	logger.Error.SetOutput(ioutil.Discard)
	logger.Warning.SetOutput(ioutil.Discard)
	logger.Trace.SetOutput(ioutil.Discard)
	logger.Debug.SetOutput(ioutil.Discard)
}
