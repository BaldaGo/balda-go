/**
 * @file logger
 * @brief Logging
 *
 * Provides function InitLogger, which initialize a logger, and logger object
 */

package logger

import (
	// System
	"errors"
	"fmt"
	"os"

	// Third-party
	"github.com/op/go-logging"
	// Project
	"github.com/BaldaGo/balda-go/conf"
)

/// Logger object
var Log = logging.MustGetLogger("logger") // Ignore error because it is impossible that it happend

/**
 * @brief Initialize logger with given format string
 * @param[in] fmt Logger format string
 */
func Init(config conf.LoggerConf) error {
	file := os.Stderr
	if config.File != "" {
		var err error
		file, err = os.Open(config.File)
		if err != nil {
			return Trace(err, "Error occured while opening log file")
		}
	}

	backend := logging.NewLogBackend(file, "> ", 0)

	format, err := logging.NewStringFormatter(config.LoggerFormat)
	if err != nil {
		return Trace(err, "Error occured while starting logger")
	}

	log := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(log)

	return nil
}

func Trace(err error, msgs ...interface{}) error {
	return errors.New(fmt.Sprintf("%s (%s)", msgs, err.Error()))
}

func Tracef(err error, format string, msgs ...interface{}) error {
	return Trace(err, fmt.Sprintf(format, msgs))
}
