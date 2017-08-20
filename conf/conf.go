/**
 * @file conf.go
 * @brief Configuration
 *
 * Provides function NewConfig which get path to json config file
 * and returns pointer to filled Config object
 */

package conf

import (
	// System
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
	// Third-party
	// Project
)

/**
 * @class Config
 * @brief Class, provides configuration for application from json file
 *
 * Config have struct same as json config struct
 * It has a basic modules structure
 */
type Config struct {
	Server ServerConf
	Logger LoggerConf
}

/**
 * @class Logger
 * @brief Class, provides configuration for Logger
 */
type LoggerConf struct {
	LoggerFormat string
	File         string
}

/**
 * @class Server
 * @brief Class, provides configuration for Server
 */
type ServerConf struct {
	Host              string
	Port              uint
	MaxSessions       uint
	ReadingBufferSize uint
	WaitTime          time.Duration
	Timeout           time.Duration
	Concurrency       int
	NumberOfGames     uint
	Game              GameConf
}

/**
 * @class Game
 * @brief Class, provides configuration for game process
 */
type GameConf struct {
	AreaSize           uint
	NumberUsersPerGame uint
	MaxUsernameLength  uint
}

/**
 * @brief Constructor of Config
 * @param[in] file Path to json config file
 * @return config Pointer to the filled Config object
 *
 * Read the file by the given filename @param file,
 * open it and parse from json to internal Go struct and return it
 **/
func New(file string) (*Config, error) {
	if file == "" {
		file = "./conf/config.json"
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't read config file: %s (%s)", file, err.Error()))
	}

	decoder := json.NewDecoder(f)
	config := new(Config)

	if decoder.Decode(config) != nil {
		return nil, errors.New("Malformed config json file")
	}

	return config, nil
}
