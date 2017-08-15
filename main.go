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
	"./conf"
	"./flags"
	"./logger"
	"./server"
)

func main() {
	flags := flags.NewFlags()
	config := conf.NewConfig(string(flags.ConfigFile), flags.Debug)

	logger.InitLogger(config.LoggerFormat)
	server := server.NewServer()

	logger.Log.Info("Server started listening on: %s:%d", config.Host, config.Port)
	server.PreRun()
	server.Run(config.Host, config.Port, flags.Debug)
	server.PostRun()
	logger.Log.Info("Server shutdowned")
}
