package server

/**
 * @file square.go
 * @brief Game 5x5 area
 *
 * Contans method for find and initialize Square object
 */

import (
	"fmt"
	"strings"
)

/*
 * @class Square
 *
 * Class, provides gaming 5*5 area
 */
type Square struct {
	matrix    [][]rune
	usedWords []string
}

/**
 * @brief Constructor of Square
 *
 * Create new Square and initialize them with random word
 */
func NewSquare(size int) *Square {
	area := new(Square)
	area.matrix = make([][]rune, size)
	for i := range area.matrix {
		area.matrix[i] = make([]rune, size)
		for j := range area.matrix[i] {
			area.matrix[i][j] = '-'
		}
	}
	for i := range area.matrix[2] {
		// TODO: Setting random five-letter word from dictionary. "ворон" for example.
		area.matrix[2][i] = []rune("ворон")[i]
	}
	return area
}

func (area *Square) destructor() {
	area = nil
}

/*
 A shallow copy is not suitable for our algorithm,
 because recursion should work with a copy of our matrix.
 This is why we use deep copy.
*/
func (b Square) deepCopy(a Square) {
	for i := range a.matrix {
		for j := range a.matrix[i] {
			b.matrix[i][j] = a.matrix[i][j]
		}
	}
}

/*
 To prevent the repetition of already founded words.
*/
func (area *Square) addUsedWord(word string) {
	area.usedWords = append(area.usedWords, word)
}

/*
 Verification that the candidate word wasn't used already in this game.
*/
func WordAlreadyInDict(listOfWords []string, word []rune) bool {
	for i := range listOfWords {
		if listOfWords[i] == string(word) {
			return true
		}
	}
	return false
}

/*
 Just pretty print of game area.
*/
func (area Square) PrintArea() {
	fmt.Print("  ")
	for j := range area.matrix {
		fmt.Printf("%d ", j)
	}
	for i := range area.matrix {
		fmt.Printf("\n%d", i)
		fmt.Printf("%c", area.matrix[i])
	}
	fmt.Print("\n")
}

/*
Adding symbol to game area cell.
*/
func (area Square) AddSymbol(x int, y int, symbol rune) {
	area.matrix[x][y] = symbol
}

/*
Recursively finding full word from first symbol for validating player's move.
*/
func (area Square) findFull(findX int, findY int, x int, y int, word []rune, checker bool) int {

	if x == len(area.matrix) || y == len(area.matrix) || (x|y) == -1 || area.matrix[x][y] != rune(word[0]) {
		return 0
	}

	areaCopy := NewSquare(len(area.matrix[0]))
	defer areaCopy.destructor() //to kill him by garbage collector
	areaCopy.deepCopy(area)
	areaCopy.matrix[x][y] = '!'

	if (findY == y) && (findX == x) {
		checker = true
	}

	if len(word) == 1 {
		if checker {
			return 1
		} else {
			return 0
		}
	}

	result := 0

	result += areaCopy.findFull(findX, findY, x, y-1, word[1:], checker) +
		areaCopy.findFull(findX, findY, x, y+1, word[1:], checker) +
		areaCopy.findFull(findX, findY, x-1, y, word[1:], checker) +
		areaCopy.findFull(findX, findY, x+1, y, word[1:], checker)

	return result
}

/**
 * @brief checkWord in Square.matrix on (x,y) position
 * @param[in] x Horisontal coordinate of required position into word
 * @param[in] y Vertical coordinate of required position into word
 * @param[in] rune Word to find
 * @param[in] symbol Letter added by player into word
 * @return error false if not found, otherwise true
 *
 */
/*
  Before calling findFull method, this method checks that:
  	1) Сandidate word wasn't userd already
 	2) The new symbol will not overlap with the existing one.
 	3) Candidate word is real word of the Russian Language (check in dictionary)
 	4) There is a letter on area with which candidate word begins.
*/
func (area *Square) checkWord(x int, y int, symbol rune, word []rune) bool {

	word = []rune(strings.ToLower(string(word)))

	if WordAlreadyInDict(area.usedWords, word) || area.matrix[x][y] != '-' {
		return false
	}
	// TODO: 1) Add word checking in dictionary (Алёна)
	tempArea := NewSquare(len(area.matrix[0]))
	defer tempArea.destructor() //to kill him by garbage collector
	tempArea.deepCopy(*area)

	tempArea.AddSymbol(x, y, symbol)

	for i := range area.matrix {
		for j := range area.matrix[i] {
			if tempArea.matrix[i][j] == rune(word[0]) &&
				(tempArea.findFull(x, y, i, j, word, false) != 0) {
				area.deepCopy(*tempArea)
				area.addUsedWord(string(word))
				return true
			}

		}
	}
	return false
}
