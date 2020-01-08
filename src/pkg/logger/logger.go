package logger

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}

// Log messages to stdout to inform developer about action/messages inside the process
func Info(v ...interface{}) {
	log.Info(fmt.Sprintf("%s", v))
}

// Log messages to stdout to inform developer about warnings/errors inside the process
func Warning(v ...interface{}) {
	log.Warn(v)
}

// Log messages to stdout to inform developer about fatal errors inside the process
// stops the process with exit code 1
func Fatal(v ...interface{}) {
	log.Fatal(v)
}
