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
	"github.com/BaldaGo/balda-go/conf"
	"github.com/BaldaGo/balda-go/flags"
	"github.com/BaldaGo/balda-go/logger"
	"github.com/BaldaGo/balda-go/server"
	"github.com/BaldaGo/balda-go/dict"
)

func main() {
	AreaSize := 5
	dict.Init(AreaSize, "dict/dictionary.txt")

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
