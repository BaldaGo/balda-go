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
	_ "github.com/go-sql-driver/mysql"

	// Project

	"github.com/BaldaGo/balda-go/conf"
	"github.com/BaldaGo/balda-go/dict"
	"github.com/BaldaGo/balda-go/logger"
	"github.com/BaldaGo/balda-go/game"
)

/// Reading channel
type ReadingChan struct {
	err     error  ///< Error
	content []byte ///< Message
}

/**
 * @class Server
 * @brief Telnet game server core
 *
 * Main class, provides telnet-server,
 * which contributes sessions, users, games, scores and other
 */
type Server struct {
	host              string        ///< Host where server will run (default 127.0.0.1)
	port              int           ///< Port where server will run (default 8888)
	maxSessions       int           ///< Maximum number of running sessions at a time (default 1000)
	readingBufferSize int           ///< Size of reading buffer in bytes (default 1)
	WaitTime          time.Duration ///< Time in milliseconds that server wait if users connection was lost (default 100)
	Timeout           time.Duration ///< Timeout in milliseconds of long operatiobs (default 1000)
	Pool              *Pool         ///< Pool of goroutines
	MaxUsernameLength int           ///< Maximum length of user name
	Sessions          []Session     ///< Array of active sessions
	Users             map[int]User  ///< Map of SessionID => User
}

/**
 * @brief Constructor of class Server
 * @param[in] cfg Server configuration
 * @return server Pointer to new Server object
 *
 * Make light and eazy fast initialisation of server directly
 */
func New(cfg conf.ServerConf) *Server {
	s := new(Server)
	s.host = cfg.Host
	s.port = cfg.Port
	s.readingBufferSize = cfg.ReadingBufferSize
	s.WaitTime = cfg.WaitTime
	s.Timeout = cfg.Timeout

	s.MaxUsernameLength = cfg.Game.MaxUsernameLength

	logger.Log.Debug("New server created")
	return s
}

/**
 * @brief Initialisation of Game server
 * @param[in] cfg Server configuration
 *
 * Create area and dict, fill other heavy game fields of server
 */

func (s *Server) PreRun(cfg conf.ServerConf) error {
	dict.Init(cfg.Game.AreaSize, cfg.DictPath)

	s.Pool = NewPool(cfg.Concurrency)
	s.Users = make(map[int]User)
	s.Sessions = make([]Session, cfg.NumberOfGames)

	var err error
	for i := 0; i < len(s.Sessions); i++ {
		s.Sessions[i].Game, err = game.NewGame(cfg.Game)
		if err != nil{
			return err
		}
	}

	s.Pool.Run()
	logger.Log.Debugf("Server is configurated with next options: %+v\n", cfg)
	return nil
}

/**
 * @brief Start Server with given parameters
 * @return err Error if critical error occured
 */
func (s *Server) Run() error {
	l, err := net.Listen("tcp", net.JoinHostPort(s.host, strconv.Itoa(s.port)))
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

		s.Pool.Add([]interface{}{s, conn})
	}

	return nil
}

/**
 * @brief Shutdown the Server
 *
 * Unlock, free all allocated memory and handlers, save data
 */
func (s *Server) PostRun() {
	errors := s.Pool.Stop()
	for _, e := range errors {
		e = logger.Trace(e, "While stopping server occured an error in goroutine")
		logger.Log.Critical(e.Error())
	}
	s = nil
	logger.Log.Debug("Server destroyed")
}

/**
 * @brief Goroutine, which called when new telnet connection established
 * @return err Error if it occured
 *
 * Listening connection with new user, login him and start his game
 */
func (t Task) work() error {
	var s *Server
	var c net.Conn

	s = t.args[0].(*Server)
	c = t.args[1].(net.Conn)

	user, err := s.login(c)
	if err != nil {
		logger.Tracef(err, "User from %s can't log in", c.RemoteAddr())
		logger.Log.Warning(err.Error())
		c.Close()
		return err
	}

	c.Write([]byte("Hello," + user.login + "!\n"))

	// Seems like the length of the buffer needs to be small, otherwise will have to wait for buffer to fill up
	buffer := make(chan ReadingChan)

	go asyncReadBytes(c, s.readingBufferSize, buffer)
	for {
		select {
		case result := <-buffer:
			if result.err != nil {
				result.err = logger.Trace(result.err, "Error in thread")
				logger.Log.Warning(result.err.Error())
				c.Close()
				return result.err
			} else {
				logger.Log.Debugf("Readed '%s' from client", result.content)
				//TODO: Validate and parse request, build and broad cast response

				go asyncReadBytes(c, s.readingBufferSize, buffer)
			}
		case <-time.After(s.Timeout * time.Millisecond):
			logger.Log.Warning("Timeout while reading...")
		}
	}

	return nil
}

/**
 * @brief Read bytes from user and push it into channel
 * @param[in] c Connection
 * @param[in] readingBufferSize Size of reading buffer
 * @param[in] buffer Channel
 */
func asyncReadBytes(c net.Conn, readingBufferSize int, buffer chan<- ReadingChan) {
	buf := make([]byte, readingBufferSize)
	n, err := c.Read(buf)
	if err != nil {
		err = logger.Trace(err, "Communication error")
		logger.Log.Warningf(err.Error())
		buffer <- ReadingChan{err: err}
		return
	}

	if n > 0 {
		buffer <- ReadingChan{content: buf}
	}
}
