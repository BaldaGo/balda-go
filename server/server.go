/**
 * @file server.go
 * @brief Server
 *
 * Core telnet game server
 */

package server

import (
	// System
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	// Third-party
	_ "github.com/go-sql-driver/mysql"

	// Project

	"github.com/BaldaGo/balda-go/conf"
	"github.com/BaldaGo/balda-go/dict"
	"github.com/BaldaGo/balda-go/game"
	"github.com/BaldaGo/balda-go/logger"
)

/*
 * @brief enum BroadcastFlags
 *
 * Specified, who will have broadcastins message
 */
const (
	BC_ALL   = 1 << iota ///< All users in this session accept message
	BC_SELF              ///< Only sender accept message
	BC_OTHER             ///< All users in this session without sender accept message
)

/**
 * @class Server
 * @brief Telnet game server core
 *
 * Main class, provides telnet-server,
 * which contributes sessions, users, games, scores and other
 */
type Server struct {
	host              string         ///< Host where server will run (default 127.0.0.1)
	port              int            ///< Port where server will run (default 8888)
	maxSessions       int            ///< Maximum number of running sessions at a time (default 1000)
	Timeout           time.Duration  ///< Timeout in seconds of waiting user play (default 30)
	TimeoutForLogin   time.Duration  ///< Timeout in seconds of waiting user login (default 300)
	Deadline          time.Duration  ///< Deadline for connection (in milliseconds) (default 1000)
	Pool              Pool           ///< Pool of goroutines
	MaxUsernameLength int            ///< Maximum length of user name
	Sessions          []Session      ///< Array of active sessions
	Users             map[string]int ///< Map of logins in each sessionID
	Signals           chan os.Signal ///< Channel of system signals like SIGINT and SIGKILL
	WaitTime          time.Duration
	SystemLogin       string
}

/**
 * @brief Constructor of class Server
 * @param[in] cfg Server configuration
 * @return server Pointer to new Server object
 *
 * Make light and eazy fast initialisation of server directly
 */
func New(cfg conf.ServerConf) Server {
	var s Server
	s.host = cfg.Host
	s.port = cfg.Port
	s.Deadline = cfg.Deadline * time.Millisecond
	s.TimeoutForLogin = cfg.TimeoutForLogin * time.Second
	s.SystemLogin = cfg.SystemLogin
	s.WaitTime = cfg.WaitTime * time.Millisecond

	s.Timeout = cfg.Game.Timeout * time.Second
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
	s.Sessions = make([]Session, cfg.NumberOfGames)
	s.Users = make(map[string]int)
	s.Signals = make(chan os.Signal, 1)
	signal.Notify(s.Signals, os.Interrupt)

	for i := 0; i < len(s.Sessions); i++ {
		s.Sessions[i].Game = game.NewGame(cfg.Game)
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
	fullNetPath := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	addr, err := net.ResolveTCPAddr("tcp", fullNetPath)
	if err != nil {
		err = logger.Tracef(err, "Can't resolve '%s'", fullNetPath)
		logger.Log.Critical(err.Error())
		return err
	}

	l, err := net.ListenTCP(addr.Network(), addr)
	if err != nil {
		err = logger.Trace(err, "Can't establish tcp connection")
		logger.Log.Critical(err.Error())
		return err
	}

	defer l.Close()

	l.SetDeadline(time.Now().Add(s.Deadline))
	logger.Log.Infof("Server started listening on: %s:%d", s.host, s.port)

	terminated := false
	for {
		select {
		case sig := <-s.Signals:
			if sig == os.Interrupt {
				s.Signals <- sig
				logger.Log.Debug("Terminated")
				terminated = true
			}
		default:
			conn, err := l.Accept()

			if err != nil {
				l.SetDeadline(time.Now().Add(s.Deadline))
				break
			}

			logger.Log.Infof("New user connected from %s", conn.RemoteAddr())
			ctx, cancel := context.WithCancel(context.WithValue(context.WithValue(context.Background(), ConnKey, conn), ServerKey, s))
			s.Pool.cancels = append(s.Pool.cancels, cancel)
			s.Pool.Add(work, ctx)
		}

		if terminated {
			for _, i := range s.Pool.cancels {
				i()
			}
			break
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
	s.Pool.Stop()
	s = nil
	logger.Log.Debug("Server destroyed")
}

/**
 * @brief Goroutine, which called when new telnet connection established
 *
 * Listening connection with new user, login him and start his game
 */
func work(ctx context.Context) error {
	var s *Server
	var c net.Conn

	s = ctx.Value(ServerKey).(*Server)
	c = ctx.Value(ConnKey).(net.Conn)

	users := make(chan User, 1)
	buffer := make(chan []byte)
	errors := make(chan net.Conn)

	context := context.WithValue(context.WithValue(context.WithValue(context.Background(), ConnKey, c), ChanKey, users), ServerKey, s)
	if err := SyncFuncWithTimeout(login, context, s.TimeoutForLogin); err != nil {
		s.broadcast("You're too slow! Sorry... Bye", s.SystemLogin, BC_SELF, errors)
		err = logger.Tracef(err, "User from %s can't log in", c.RemoteAddr())
		logger.Log.Warning(err.Error())
		time.Sleep(s.WaitTime)
		c.Close()
		return err
	}

	user := <-users
	logger.Log.Infof("User from %s logined as %s and associated with session %d", c.RemoteAddr(), user.login, user.sessionId)

	s.broadcast(fmt.Sprintf("Welcome %s!\n\rPlease, wait other players...", user.login), user.login, BC_ALL, errors)
	go asyncReadBytes(c, buffer, errors)

	terminated := false
	for {
		select {
		case result := <-buffer:
			logger.Log.Debugf("Readed '%s' from client", result)

			// Game interactive
			play, response := s.Sessions[user.sessionId].Game.Continue(string(result), user.login)
			logger.Log.Debugf("Generic answers '%s'. Continue: %b", response, play)
			if !play {
				s.broadcast(response, user.login, BC_ALL, errors)
				logger.Log.Infof("Game over! %s", response)
				for _, u := range s.Sessions[user.sessionId].Users {
					u.conn.Close()
				}
				terminated = true
			} else {
				s.broadcast(response, user.login, BC_ALL, errors)
			}

			go asyncReadBytes(c, buffer, errors)

		case <-time.After(s.Timeout):
			logger.Log.Warning("Timeout while reading...")
			s.broadcast(fmt.Sprintf("%s doesn't catch his move", user.login), user.login, BC_OTHER, errors)
			s.broadcast("You're too slow!", s.SystemLogin, BC_SELF, errors)

			if terminated {
				c.Close()
				break
			}

		case c := <-errors:
			logger.Log.Warningf("User from %s failed", c.RemoteAddr())
			return c.Close()

		case <-ctx.Done():
			logger.Log.Debug("Terminated")
			terminated = true
		}

		if terminated {
			logger.Log.Debug("Work finished")
			return nil
		}
	}
}

/**
 * @brief Write bytes to user
 * @param[in] c Connection
 * @param[in] bytes Array of bytes to write
 * @param[in] errors Channel with failed connections
 */
func asyncWriteBytes(c net.Conn, bytes []byte, errors chan<- net.Conn) {
	n, err := c.Write(bytes)
	if err != nil || n != len(bytes) {
		if c == nil {
			logger.Log.Warningf("Error while writing to connection (%s)", c.RemoteAddr(), err.Error())
		} else {
			logger.Log.Warningf("Error while writing to connection to %s (%s)", c.RemoteAddr(), err.Error())
			errors <- c
		}
	}
}

/**
 * @brief Read bytes from user and push it into channel
 * @param[in] c Connection
 * @param[in] buffer Output channel with results
 * @param[in] errors Channel with failed connections
 */
func asyncReadBytes(c net.Conn, buffer chan<- []byte, errors chan<- net.Conn) {
	io := bufio.NewReader(c)
	buf, err := io.ReadString('\n')
	line := strings.Replace(strings.Replace(buf, "\n", "", -1), "\r", "", -1)
	if err != nil || line == "" {
		if c == nil {
			logger.Log.Warningf("Error while reading from connection (%s)", c.RemoteAddr(), err.Error())
		} else {
			logger.Log.Warningf("Error while reading from connection from %s (%s)", c.RemoteAddr(), err.Error())
			errors <- c
		}
	} else {
		buffer <- []byte(line)
	}
}

func (s *Server) broadcast(raw string, login string, flags int, errors chan<- net.Conn) {
	msg := fmt.Sprintf("%s> %s\n\r", login, raw)
	sessionID := s.Users[login]
	msg = fmt.Sprintf("%s\n", msg)
	for _, i := range s.Sessions[sessionID].Users {
		switch flags {
		case BC_ALL:
			go asyncWriteBytes(i.conn, []byte(msg), errors)
		case BC_SELF:
			if i.login == login {
				go asyncWriteBytes(i.conn, []byte(msg), errors)
			}
		case BC_OTHER:
			if i.login != login {
				go asyncWriteBytes(i.conn, []byte(msg), errors)
			}
		default: // BC_ALL
			go asyncWriteBytes(i.conn, []byte(msg), errors)
		}
	}
}
