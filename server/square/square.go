/**
 * @file square.go
 * @brief Game 5x5 area
 *
 * Contans method for find and initialize Square object
 */

package square

/**
 * @class Square
 *
 * Class, provides gaming 5*5 area
 */
type Square struct {
	matrix [][]rune
}

/**
 * @brief Constructor of Square
 *
 * Create new Square and initialize them with random word
 */
func NewSquare() *Square {
	//TODO: initialize
	return new(Square)
}

/**
 * @brief Find word in Square.matrix on (x,y) position
 * @param[in] x Horisontal coordinate of required position into word
 * @param[in] y Vertical coordinate of required position into word
 * @param[in] word Word to find
 * @return error Error if not found, otherwise nil
 *
 * @todo: Implement alghoritm
 */
func (s Square) find(x int, y int, word string) error {
	//TODO: implement this method
	return nil
}
