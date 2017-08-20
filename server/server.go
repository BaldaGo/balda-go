/**
 * @file server.go
 * @brief Server
 *
 * Core telnet game server
 */

package server

import (
	// System
	"net"
	"strconv"
	"time"

	// Third-party

	// Project
	"github.com/BaldaGo/balda-go/conf"
	"github.com/BaldaGo/balda-go/dict"
	"github.com/BaldaGo/balda-go/logger"
)

/**
 * @class Server
 * @brief Telnet game server core
 *
 * Main class, provides telnet-server,
 * which contributes sessions, users, games, scores and other
 */
type Server struct {
	host              string
	port              uint
	maxSessions       uint
	readingBufferSize uint
	WaitTime          time.Duration
	Timeout           time.Duration
	Pool              *Pool
	MaxUsernameLength uint
	Sessions          []Session
	Users             map[uint]User
}

/**
 * @brief Constructor of class Server
 * @return server Pointer to new Server object
 *
 * Make light and eazy fast initialisation of server directly
 */
func New(cfg conf.ServerConf) *Server {
	s := new(Server)
	s.host = cfg.Host
	s.port = cfg.Port
	s.maxSessions = cfg.MaxSessions
	s.readingBufferSize = cfg.ReadingBufferSize
	s.WaitTime = cfg.WaitTime
	s.Timeout = cfg.Timeout

	s.MaxUsernameLength = cfg.Game.MaxUsernameLength

	logger.Log.Debug("New server created")
	return s
}

/**
 * @brief Initialisation of Game server
 *
 * Create area and dict, fill other heavy game fields of server
 */
func (s *Server) PreRun(cfg conf.ServerConf) {
	dict.Init(AreaSize, "dict/dictionary.txt")

	s.Pool = NewPool(cfg.Concurrency)
	s.Users = make(map[uint]User)
	s.Sessions = make([]Session, cfg.NumberOfGames)

	for i := 0; i < len(s.Sessions); i++ {
		s.Sessions[i].Game = NewGame()
	}

	s.Pool.Run()
	logger.Log.Debugf("Server is configurated with next options: %+v\n", cfg)
}

/**
 * @brief Start Server on given host and port
 */
func (s *Server) Run() error {
	l, err := net.Listen("tcp", net.JoinHostPort(s.host, strconv.Itoa(int(s.port))))
	if err != nil {
		err = logger.Trace(err, "Can't establish tcp connection")
		logger.Log.Critical(err.Error())
		return err
	}
	defer l.Close()

	logger.Log.Infof("Server started listening on: %s:%d", s.host, s.port)

	for {
		conn, err := l.Accept()
		if err != nil {
			err = logger.Trace(err, "Failed to accept request. Retrying...")
			logger.Log.Critical(err.Error())
			continue
		}

		logger.Log.Infof("New user connected from %s", conn.RemoteAddr())

		if _, err := s.Pool.AddWithTimeout([]interface{}{s, conn}, s.Timeout*time.Millisecond); err != nil {
			logger.Log.Critical("Can't accept new connection")
		}
	}

	return nil
}

/**
 * @brief Shutdown the Server
 *
 * Unlock, free all allocated memory and handlers, save data
 */
func (s *Server) PostRun() {
	s = nil
	logger.Log.Debug("Server destroyed")
}

/**
 * @brief Goroutine, which called when new telnet connection established
 * @param[in] c Connection with new listening client
 * @param[in] pull Sessions pull (channel)
 *
 * Listening connection with new user
 */
func (t *Task) work() interface{} {
	var s *Server
	var c net.Conn

	s = t.args[0].(*Server)
	c = t.args[1].(net.Conn)

	user, err := s.login(c)
	if err != nil {
		logger.Tracef(err, "User from %s can't log in", c.RemoteAddr())
		logger.Log.Warning(err.Error())
		c.Close()
		time.Sleep(s.WaitTime * time.Millisecond)
		return err
	}

	c.Write([]byte("Hello," + user.login + "!\n"))

	// Seems like the length of the buffer needs to be small, otherwise will have to wait for buffer to fill up
	buffer := make([]byte, s.readingBufferSize)

	for {
		n, err := c.Read(buffer)

		if err != nil {
			logger.Log.Warningf("Communication error")
			break
		}

		if n > 0 {
			logger.Log.Debugf("Readed %d bytes from client", n)
			//TODO: Validate and parse request, build and broad cast response
		}
	}

	return nil
}
