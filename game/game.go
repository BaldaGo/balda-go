/**
 * @file game.go
 * @brief Game
 *
 * Contains Game type and methods to interact with it
 */
package game

import (
	// System
	// Third-party
	// Project
	"github.com/BaldaGo/balda-go/conf"
)

/**
 * @class Game
 * @brief Class, provide information about concrete game
 */
type Game struct {
	square Square ///< Gaming area
}

/**
 * @brief Create a new game
 * @return game Pointer to the created Game object
 */
func NewGame(cfg conf.GameConf) Game {
	return Game{
		square: NewSquare(cfg.AreaSize),
	}
}
