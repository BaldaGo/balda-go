/**
 * @file user.go
 * @brief Users and Sessions
 *
 * Contains User and Session types,
 * Methods to login new user
 */
package server

import (
	// System
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"unicode/utf8"

	// Third-party

	// Project
	"github.com/BaldaGo/balda-go/db"
	"github.com/BaldaGo/balda-go/game"
	"github.com/BaldaGo/balda-go/logger"
)

var (
	clear_home = []byte{27, 91, 72, 27, 91, 50, 74} ///< Bytes sequence to clear user screen
)

/**
 * @class User
 * @brief Class, provides information about User
 */
type User struct {
	conn      net.Conn ///< Connection with server
	login     string   ///< User's login
	sessionId int      ///< Id of session
}

/**
 * @class Session
 * @brief Class, provides information about Session
 *
 * Session is a thing which aggregate users in one game
 */
type Session struct {
	Users []User    ///< Array of users in this session
	Game  game.Game ///< Game object
}

/**
 * @brief login user and associate them with session
 * @param[in] c Connection
 * @return user Pointer to new User object or error if it occured
 *
 * Prompt user, ask him to put his name and session id, login him and associate with session by id
 */
func login(ctx context.Context) error {
	var s *Server
	var c net.Conn
	var user chan User

	s = ctx.Value(ServerKey).(*Server)
	c = ctx.Value(ConnKey).(net.Conn)
	user = ctx.Value(ChanKey).(chan User)

	// Say welcome and ask username
	c.Write([]byte("Welcome to balda game!\nPlease, enter your name to log in: "))

	// Read and validate username
	io := bufio.NewReader(c)
	line, err := io.ReadString('\n')
	if err != nil {
		return logger.Trace(err, "Communication error")
	}

	name := strings.Replace(strings.Replace(line, "\n", "", -1), "\r", "", -1)
	if name == "" {
		return errors.New("Empty name")
	}

	if utf8.RuneCountInString(name) > s.MaxUsernameLength {
		return errors.New("Too long name")
	}

	logger.Log.Debugf("User from %s logined with login: %s", c.RemoteAddr(), name)

	// Read password
	pass, err := io.ReadString('\n')
	if err != nil {
		return logger.Trace(err, "Communication error")
	}

	pass = strings.Replace(strings.Replace(pass, "\n", "", -1), "\r", "", -1)
	if pass == "" {
		return errors.New("Empty password")
	}

	// Create user if not exists
	exists, err := db.CheckUser(name, pass)
	if err != nil {
		err = logger.Trace(err, "Database error")
		logger.Log.Critical(err.Error())
		return err
	}

	if !exists {
		u, err := db.AddUser(name, pass, c.RemoteAddr().String())
		if err != nil {
			err = logger.Trace(err, "Database error")
			logger.Log.Critical(err.Error())
			return err
		}
		logger.Log.Debug("New user created", *u)
	}

	// Read and validate session id
	c.Write([]byte("Please, enter the number of game your want to assign: "))
	line, err = io.ReadString('\n')
	if err != nil {
		return logger.Trace(err, "Communication error")
	}

	line = strings.Replace(strings.Replace(line, "\n", "", -1), "\r", "", -1)
	if line == "" {
		return errors.New("Empty session id")
	}

	SessionID, err := strconv.Atoi(line)
	if err != nil {
		return logger.Trace(err, "Session ID must be a positive integer")
	} else if int(SessionID) >= len(s.Sessions) {
		return errors.New(fmt.Sprintf("Session with ID=%d is not exists (Session ID is too big)", SessionID))
	} else if int(SessionID) < 0 {
		return errors.New("Session ID must be a positive integer")
	}

	if len(s.Sessions[SessionID].Users) >= s.Sessions[SessionID].Game.MaxUsersPerGame {
		return errors.New("Sorry, this game is already starts")
	}

	// Create a new user object
	newUser := User{conn: c, login: name, sessionId: SessionID}

	// Associate user with a session by session id
	err = s.Sessions[SessionID].Game.AddUser(name)
	if err != nil {
		return logger.Trace(err, "Can't accept the game")
	}

	s.Sessions[SessionID].Users = append(s.Sessions[newUser.sessionId].Users, newUser)
	s.Users[newUser.login] = SessionID

	if len(s.Sessions[SessionID].Users) == s.Sessions[SessionID].Game.MaxUsersPerGame {
		s.Sessions[SessionID].Game.StartGame()
		logger.Log.Info("Game started:", s.Sessions[SessionID].Game)
		errors := make(chan net.Conn)
		s.broadcast("Game started!", s.SystemLogin, BC_ALL, errors)
		for c := range errors {
			logger.Log.Warning("Error occured while sending 'Game started' to", c.RemoteAddr())
			c.Close()
		}
	}

	user <- newUser
	logger.Log.Debugf("User from %s associated with session: %d", c.RemoteAddr(), SessionID)

	return nil
}
