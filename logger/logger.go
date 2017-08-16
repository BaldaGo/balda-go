/**
 * @file logger
 * @brief Logging
 *
 * Provides function InitLogger, which initialize a logger, and logger object
 */

package logger

import (
	// System
	"os"

	// Third-party
	"github.com/op/go-logging"
	// Project
)

/// Logger object
var Log = logging.MustGetLogger("logger")

/**
 * @brief Initialize logger with given format string
 * @param[in] fmt Logger format string
 */
func InitLogger(fmt string) error {
	backend := logging.NewLogBackend(os.Stderr, "> ", 0)

	format, err := logging.NewStringFormatter(fmt)
	if err != nil {
		return err
	}

	log := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(log)

	return nil
}
