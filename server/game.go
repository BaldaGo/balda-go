/**
 * @file game.go
 * @brief Game
 *
 * Contains Game type and methods to interact with it
 */
package server

/**
 * @class Game
 * @brief Class, provide information about concrete game
 */
type Game struct {
	square *Square ///< Gaming area
}

/**
 * @brief Create a new game
 * @return game Pointer to the created Game object
 */
func NewGame() *Game {
	return &Game{square: NewSquare(5)}
}
