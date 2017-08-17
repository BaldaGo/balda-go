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

/*
 Constructor of Square object
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
func (a *Square) addUsedWord(word string) {
	a.usedWords = append(a.usedWords, word)
}

/*
 Verification that the candidate word wasn't used already in this game.
*/
func WordAlreadyInDict(dict []string, word []rune) bool {
	for i := range dict {
		if dict[i] == string(word) {
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
	areaCopy.deepCopy(area)
	areaCopy.matrix[x][y] = '!'

	if (findY == y) && (findX == x) {
		checker = true
	}

	if len(word) == 1 {
		areaCopy = nil
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

	areaCopy = nil //to kill him by garbage collector
	return result
}

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
	// TODO: 1) Add word checking in dicrionary (Алёна)
	tempArea := NewSquare(len(area.matrix[0]))
	tempArea.deepCopy(*area)

	tempArea.AddSymbol(x, y, symbol)

	for i := range area.matrix {
		for j := range area.matrix[i] {
			if tempArea.matrix[i][j] == rune(word[0]) &&
				(tempArea.findFull(x, y, i, j, word, false) != 0) {
				area.deepCopy(*tempArea)
				area.addUsedWord(string(word))
				tempArea = nil //to kill him by garbage collector
				return true
			}

		}
	}
	tempArea = nil //to kill him by garbage collector
	return false
}

func main() {

	area := NewSquare(5)

	//test game
	fmt.Println(area.checkWord(3, 0, 'в', []rune("ВвОроН")))
	fmt.Println(area.checkWord(1, 0, 'в', []rune("вворон")))
	fmt.Println(area.checkWord(3, 2, 'ы', []rune("воРы")))
	fmt.Println(area.checkWord(1, 3, 'г', []rune("рог")))
	fmt.Println(area.checkWord(1, 4, 'а', []rune("НОГА")))
	fmt.Println(area.checkWord(1, 2, 'ы', []rune("ВОРЫ")))
	area.PrintArea()

	area = nil

}
