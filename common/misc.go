package common

import (
	"fmt"
	"log"
	"os"
)

const (
	cPidPattern = "%d"
	cPidFile    = "/var/run/trial.pid"
	cLogFile    = "/var/log/trial.log"
)

// SetupLog sets logging to cLogFile.
func SetupLog() {
	f, err := os.OpenFile(cLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		os.Exit(1)
	}
	log.SetOutput(f)
}

// WritePid writes our pid at start up (used for monitoring the service).
func WritePid() {
	pid := os.Getpid()
	f, err := os.Create(cPidFile)
	if err != nil {
		// Write to log.
		log.Printf("Failed to write PID file; %v", err)
	} else {
		f.WriteString(fmt.Sprintf(cPidPattern, pid))
	}
	f.Close()
}
