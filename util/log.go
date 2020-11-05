package util

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
)

type Logger struct {
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Debug   *log.Logger
	Default *log.Logger
}

func (logger *Logger) Init(prefix string) {
	logPath := path.Join("logs", "runtime.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	logger.Trace = log.New(io.MultiWriter(file, os.Stdout),
		fmt.Sprintf("[%s TRACE]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Info = log.New(io.MultiWriter(file, os.Stdout),
		fmt.Sprintf("[%s INFO]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Warning = log.New(io.MultiWriter(file, os.Stdout),
		fmt.Sprintf("[%s WARNING]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Error = log.New(io.MultiWriter(file, os.Stderr),
		fmt.Sprintf("[%s ERROR]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Debug = log.New(os.Stdout,
		fmt.Sprintf("[%s DEBUG]: ", prefix),
		log.Ldate|log.Ltime|log.Lshortfile)

	logger.Default = log.New(io.MultiWriter(file, os.Stdout), "", 0)
}
