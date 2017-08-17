/**
 * @file server.go
 * @brief Server
 *
 * Core telnet game server
 */

package server

import (
	// System
	"fmt"
	"strconv"

	// Third-party

	// Project
	"github.com/BaldaGo/game/logger"
	"github.com/reiver/go-oi"
	"github.com/reiver/go-telnet"
)

/**
 * @class Server
 * @brief Telnet game server core
 *
 * Main class, provides telnet-server,
 * which contributes sessions, users, games, scores and other
 */
type Server struct {
	square  Square
	handler telnet.Handler
}

/// @class handler - handle telnet connection
type handler struct{}

/**
 * @brief Constructor of class Server
 * @return server Pointer to new Server object
 */
func NewServer() *Server {
	return new(Server)
}

/**
 * @brief Initialisation of Server
 *
 * Create handle and fill other fields of server
 */
func (s *Server) PreRun() {
	s.handler = handler{}

	//TODO
}

/**
 * @brief Start Server on given host and port
 * @param[in] host Hostname where server will listen connections
 * @param[in] port Port where server will listen connections
 * @param[in] debug Flag, write debug info if given
 */
func (s *Server) Run(host string, port uint, debug bool) error {
	if err := telnet.ListenAndServe(host+":"+strconv.Itoa(int(port)), s.handler); err != nil {
		logger.Log.Critical("Server failed to startup on: %s:%d (%s)", host, port, err.Error())
		return err
	}

	//TODO
	return nil
}

/**
 * @brief Shutdown the Server
 *
 * Unlock, free all allocated memory and handlers, save data
 */
func (s *Server) PostRun() {
	//TODO
}

/**
 * @brief Goroutine, which called when new telnet connection established
 * @param[in] ctx Context of connection
 * @param[in] w io.Writer object
 * @param[in] r io.Reader object
 *
 * Echo server
 */
func (h handler) ServeTELNET(ctx telnet.Context, w telnet.Writer, r telnet.Reader) {
	var buffer [1]byte // Seems like the length of the buffer needs to be small, otherwise will have to wait for buffer to fill up.
	p := buffer[:]

	for {
		n, err := r.Read(p)

		if n > 0 {
			logger.Log.Info(fmt.Sprintf("Readed %d bytes from client", n))
			//TODO: Validate and parse request, build and broad cast response
			oi.LongWrite(w, p[:n]) // Echo
		}

		if err != nil {
			logger.Log.Warning("An error occured: %s", err.Error())
			break
		}
	}
}
