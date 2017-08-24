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
)

/**
 * @brief Enter point of the application
 *
 * Read command line flags, environment variables, configurations, dictionary
 * Starts server
 */
func main() {
	flags := flags.New()

	config, err := conf.New(string(flags.ConfigFile))
	if err != nil {
		panic(err)
	}

	logger.Init(config.Logger)

	server := server.New(config.Server)
	if res := server.PreRun(config.Server); res.Error != nil {
		panic(res.Error())
	}
	defer server.PostRun()

	err = server.Run()
	if err != nil {
		panic(err)
	}

	logger.Log.Info("Server shutdowned")
}
