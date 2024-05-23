package utils

import (
	"github.com/DistributedClocks/GoVector/govec"
	"log"
	"os"
	"path/filepath"
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

	// Check if output directory exists: if not, create it
	if _, err := os.Stat(OutputDir); os.IsNotExist(err) {
		err := os.MkdirAll(OutputDir, os.ModePerm)
		if err != nil {
			log.Fatalln("Failed to create output directory:", err)
		}
	}

	// Remove all content from the output directory (if it is not empty)
	/*err := removeContents(OutputDir)
	if err != nil {
		log.Fatalln("Error while removing files from output directory:", err)
	}
	*/
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

func removeContents(dirPath string) error {
	// Read all files and directories in the specified directory
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Iterates over all files and directories found
	for _, file := range files {
		// Create the full path to the file or directory
		filePath := filepath.Join(dirPath, file.Name())

		// Check if it is a directory
		if file.IsDir() {
			// Recursively remove the contents of the directory
			err = os.RemoveAll(filePath)
			if err != nil {
				return err
			}
		} else {
			// Remove file
			err = os.Remove(filePath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
