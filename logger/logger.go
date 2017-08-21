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
 * @return err Error if it occured
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

/**
 * @brief Add additional information into error message to trace it
 * @param[in] err Error
 * @param[in] msgs Additional information
 * @return err Patched error
 */
func Trace(err error, msgs ...interface{}) error {
	return errors.New(fmt.Sprintf("%s (%s)", msgs, err.Error()))
}

/**
 * @brief Add format additional information into error message to trace it
 * @param[in] err Error
 * @param[in] format Format string
 * @param[in] msgs Additional information
 * @return err Patched error
 *
 * Same as Trace, but get args with format string
 */
func Tracef(err error, format string, msgs ...interface{}) error {
	return Trace(err, fmt.Sprintf(format, msgs))
}
