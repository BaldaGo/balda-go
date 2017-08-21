/**
 * @file telnet.go
 * @brief Telnet protocol
 *
 * Contains specific for this protocol methods
 */
package server

import (
	// System
	"errors"
	"net"
	// Third-party
	// Project
	"github.com/BaldaGo/balda-go/logger"
)

/**
 * @brief Handshake telnet and disable buffering
 * @param[in] conn Connection
 * @return err Error if it occured
 *
 * Send to remote client initial telnet hanbshake settings message
 */
func initTelnet(conn net.Conn) error {
	// https://tools.ietf.org/html/rfc854
	telnetOptions := []byte{
		255, 253, 34, // IAC DO LINEMODE
		255, 250, 34, 1, 0, 255, 240, // IAC SB LINEMODE MODE 0 IAC SE
		255, 251, 1, // IAC WILL ECHO
	}
	_, err := conn.Write(telnetOptions)
	if err != nil {
		logger.Log.Warning("Error occured while initializing telnet connection")
		return logger.Trace(err, "Error occured while initializing telnet connection")
	}

	logger.Log.Debug("Established telnet connection")
	return nil
}

/**
 * @brief Read handshake telnet message from client
 * @param[in] conn Connection
 * @return err Error if it occured
 *
 * Read bytes from client and check if it is a short message
 */
func readTelnet(conn net.Conn) error {
	// https://tools.ietf.org/html/rfc854
	reply := make([]byte, 1)
	bytesRead := 0
	shortCommand := false

	for {
		_, err := conn.Read(reply)
		if err != nil {
			return err
		}
		bytesRead++

		if reply[0] != 250 && bytesRead == 1 {
			shortCommand = true
		}

		if shortCommand && bytesRead == 2 {
			return nil
		} else if reply[0] == 240 {
			return nil
		}
	}
}

/**
 * @brief
 * @param[in] c Connection
 * @return err Error if it exists
 *
 * Find ESC [ sequens in given bytes stream and read control symols
 */
func readControlButtons(c net.Conn) error {
	if err := initTelnet(c); err != nil {
		err = logger.Trace(err, "Error occured while initialization telnet")
		logger.Log.Warning(err.Error())
		c.Close()
		return err
	}

	button := make([]byte, 1)

	// Read all possible bytes and try to find a sequence of:
	// ESC [ button
	escpos := 0
	for {
		_, err := c.Read(button)
		if err != nil {
			err = logger.Trace(err, "Communication error")
			logger.Log.Warning(err.Error())
			c.Close()
			return err
		}

		// Check if telnet want to negotiate something
		if escpos == 0 && button[0] == 255 {
			readTelnet(c)
		} else if escpos == 0 && button[0] == 3 {
			// Ctrl+C
			return nil
		} else if escpos == 0 && button[0] == 32 {
			// Space
			return nil
		} else if escpos == 0 && button[0] == 27 {
			escpos = 1
		} else if escpos == 1 && button[0] == 91 {
			escpos = 2
		} else if escpos == 2 {
			break
		} else {
			logger.Log.Debug("User push wrong button")
			return errors.New("User push wrong button")
		}
	}

	return nil
}
