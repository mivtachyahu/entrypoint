package logger

import (
	"io"
	"log"
)

var (
	// Trace - Trace Logging
	Trace *log.Logger
	// Info - Info Logging
	Info *log.Logger
	// Warning - Warning Logging
	Warning *log.Logger
	// Error - Error Logging
	Error *log.Logger
)

// LogInit - Initialise the logging styles
func LogInit(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {
	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}
