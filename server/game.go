package server

/**
 * @class Game
 * @brief Class, provide information about concrete game
 */
type Game struct {
	square *Square ///< Gaming area
}

func NewGame() *Game {
	return &Game{square: NewSquare(5)}
}
