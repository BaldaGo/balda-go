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
	"github.com/BaldaGo/balda-go"
)

/// Logger object
var Log = logging.MustGetLogger("logger") // Ignore error because it is impossible that it happend

/**
 * @brief Initialize logger with given format string
 * @param[in] fmt Logger format string
 */
func InitLogger(config conf.Logger) error {
	backend := logging.NewLogBackend(os.Stderr, "> ", 0)

	format, err := logging.NewStringFormatter(config.LoggerFormat)
	if err != nil {
		return "Error occured while starting logger: " + err.Error()
	}

	log := logging.NewBackendFormatter(backend, format)

	logging.SetBackend(log)

	return nil
}
