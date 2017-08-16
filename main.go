/**
 * @file main.go
 * @brief Balda the game
 *
 * @authors Nikita-Boyarskikh gabolaev AlenaFedotova
 * @version 1.0.0
 *
 * Simple casual logic game with telnet interface
 */

package main

import (
	// System

	// Third-party

	// Project
	"github.com/BaldaGo/game/conf"
	"github.com/BaldaGo/game/flags"
	"github.com/BaldaGo/game/logger"
	"github.com/BaldaGo/game/server"
)

func main() {
	flags := flags.NewFlags()
	config, err := conf.NewConfig(string(flags.ConfigFile), flags.Debug)
	if err != nil {
		panic(err)
	}

	logger.InitLogger(config.LoggerFormat)
	server := server.NewServer()

	logger.Log.Info("Server started listening on: %s:%d", config.Host, config.Port)

	server.PreRun()
	err = server.Run(config.Host, config.Port, flags.Debug)
	if err != nil {
		panic(err)
	}
	server.PostRun()

	logger.Log.Info("Server shutdowned")
}
