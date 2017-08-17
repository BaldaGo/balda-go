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
	"os"
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
	ServerConf Server
	LoggerConf Logger
}

/**
 * @class Logger
 * @brief Class, provides configuration for Logger
 */
type Logger struct {
	LoggerFormat string
}

/**
 * @class Server
 * @brief Class, provides configuration for Server
 */
type Server struct {
	Host string
	Port uint
}

/**
 * @brief Constructor of Config
 * @param[in] file Path to json config file
 * @param[in] debug Flag, config with debug if given
 * @return config Pointer to the filled Config object
 *
 * Read the file by the given filename @param file,
 * open it and parse from json to internal Go struct and return it
 **/
func NewConfig(file string) (*Config, error) {
	if file == "" {
		file = "./conf/config.json"
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, errors.New("Can't read config file: " + file)
	}

	decoder := json.NewDecoder(f)
	config := new(Config)

	if decoder.Decode(config) != nil {
		return nil, errors.New("Malformed config json file")
	}

	return config, nil
}
