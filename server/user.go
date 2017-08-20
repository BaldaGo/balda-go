package server

import (
	// System
	"bufio"
	"errors"
	"net"
	"strconv"
	"strings"
	"unicode/utf8"
	// Third-party
	// Project
	"github.com/BaldaGo/balda-go/logger"
)

var (
	clear_home = []byte{27, 91, 72, 27, 91, 50, 74}
)

/**
 * @class User
 * @brief Class, provides information about User
 */
type User struct {
	conn      net.Conn ///< Connection with server
	login     string   ///< User's login
	sessionId uint     ///< Id of session
}

/**
 * @class Session
 * @brief Class, provides information about Session
 *
 * Session is a thing which aggregate users in one game
 */
type Session struct {
	id    uint   ///< Id of session
	users []User ///< Array of users in this session
	Game  *Game  ///< Game object
}

func (s *Server) login(c net.Conn) (*User, error) {
	c.Write([]byte("Welcome to balda game!\nPlease, enter your name to log in: "))

	io := bufio.NewReader(c)
	line, err := io.ReadString('\n')
	if err != nil {
		return nil, logger.Trace(err, "Communication error")
	}

	name := strings.Replace(strings.Replace(line, "\n", "", -1), "\r", "", -1)
	if name == "" {
		return nil, errors.New("Empty name")
	}

	if uint(utf8.RuneCountInString(name)) > s.MaxUsernameLength {
		return nil, errors.New("Too long name")
	}

	c.Write([]byte("Please, enter the number of game your want to assign: "))
	line, err = io.ReadString('\n')
	if err != nil {
		return nil, logger.Trace(err, "Communication error")
	}

	line = strings.Replace(strings.Replace(line, "\n", "", -1), "\r", "", -1)
	if line == "" {
		return nil, errors.New("Empty session id")
	}

	SessionID, err := strconv.ParseInt(line, 10, 32)
	if err != nil || uint(SessionID) > uint(len(s.Sessions)) {
		return nil, logger.Trace(err, "Session id must be a positive integer")
	}

	user := User{conn: c, login: name, sessionId: uint(SessionID)}

	s.Users[uint(SessionID)] = user
	s.Sessions[SessionID].id = user.sessionId
	s.Sessions[SessionID].users = append(s.Sessions[user.sessionId].users, user)

	return &user, nil
}

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
		}
	}

	return nil
}
