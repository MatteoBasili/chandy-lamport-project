package utils

import (
	"github.com/DistributedClocks/GoVector/govec"
	"log"
	"os"
)

const OutputDir = "output/"

type Logger struct {
	// Logs
	Trace    *log.Logger
	Info     *log.Logger
	Warning  *log.Logger
	Error    *log.Logger
	GoVector *govec.GoLog
}

func InitLoggers(name string) *Logger {

	// Initialize log
	fLog, err := os.OpenFile(OutputDir+"Log_P"+name+".log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file:", err)
	}

	myLogger := Logger{}
	myLogger.Trace = log.New(fLog,
		"TRACE: \t\t[P"+name+"] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)

	myLogger.Info = log.New(fLog,
		"INFO: \t\t[P"+name+"] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)

	myLogger.Warning = log.New(fLog,
		"WARNING: \t[P"+name+"] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)

	myLogger.Error = log.New(fLog,
		"ERROR: \t\t[P"+name+"] ", log.Ltime|log.Lmicroseconds|log.Lshortfile)

	//Initialize GoVector logger
	config := govec.GetDefaultConfig()
	config.UseTimestamps = true
	myLogger.GoVector = govec.InitGoVector("P"+name, OutputDir+"GoVector/LogFileP"+name, config)

	return &myLogger
}
