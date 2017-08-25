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
	Server ServerConf ///< Server configurations
	Logger LoggerConf ///< Logger configurations
	Database DatabaseConf ///< Database configurations
}

/**
 * @class Logger
 * @brief Class, provides configuration for Logger
 */
type LoggerConf struct {
	LoggerFormat string ///< Format string (hot to output ingformation)
	File         string ///< Log file (where to output information)
}

/**
 * @class Server
 * @brief Class, provides configuration for Server
 */
type ServerConf struct {
	Host              string        ///< Host where server will run (default 127.0.0.1)
	Port              int           ///< Port where server will run (default 8888)
	NumberOfGames     int           ///< Maximum number of running sessions at a time (default 1000)
	ReadingBufferSize int           ///< Size of reading buffer in bytes (default 1)
	WaitTime          time.Duration ///< Time in milliseconds that server wait if users connection was lost (default 100)
	Timeout           time.Duration ///< Timeout in milliseconds of long operatiobs (default 1000)
	Concurrency       int           ///< Number of workers in goroutines pool (default 4000)
	Game              GameConf      ///< Game configurations
	DictPath 		  string		///< Russian language Dictionary path
}

/**
 * @class DatabaseConf
 * @brief Class, provides configuration for db connection
 */
type DatabaseConf struct {
	Dialect 	string
	User 		string
	Password	string
	Host 		string
	Port 		int
	Name		string
	Options 	map[string]string
}

/**
 * @class Game
 * @brief Class, provides configuration for game process
 */
type GameConf struct {
	AreaSize           int ///< Length side of the playing area (default 5)
	NumberUsersPerGame int ///< Maximum number of gaming users at a time (default 4)
	MaxUsernameLength  int ///< Maximum username length (default 255)
}

/**
 * @brief Constructor of Config
 * @param[in] file Path to json config file
 * @return config Pointer to the filled Config object or error if it occured
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
