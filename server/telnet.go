package server

import (
	// System
	"net"
	// Third-party
	// Project
	"github.com/BaldaGo/balda-go/logger"
)

/**
 * @brief Handshake telnet and disable buffering
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
