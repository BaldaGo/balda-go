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
	"os"
	// Third-party
	// Project
)

/**
 * @class Config
 * @brief Class, provides configuration of application from json file
 *
 * Config have struct same as json config struct
 */
type Config struct {
	//TODO: add all needed structs
	Host         string
	Port         uint
	LoggerFormat string
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
func NewConfig(file string, debug bool) *Config {
	if file == "" {
		file = "./conf/config.json"
	}
	f, err := os.Open(file)
	if err != nil {
		panic("Can't read config file: " + file)
	}

	decoder := json.NewDecoder(f)
	config := new(Config)

	if decoder.Decode(config) != nil {
		panic("Malformed config json file")
	}

	return config
}
