/**
 * @file game.go
 * @brief Game
 *
 * Contains Game type and methods to interact with it
 */
package game

import (
	// System
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
	"fmt"

	// Third-party

	// Project
	"github.com/BaldaGo/balda-go/conf"
	"github.com/fatih/structs"
	//"github.com/BaldaGo/balda-go/db"
)

/**
 * @class Game
 * @brief Class, provide information about concrete game
 */
type Game struct {
	square          Square ///< Gaming area
	users           []string
	scoreMap        map[string]int
	stepUser        int
	onStart         bool
	skipped         int
	putting         Put
	onPut           bool
	AreaSize        int
	MaxUsersPerGame int
	meth            methods
}

type Put struct {
	state   string
	x       int
	y       int
	sym     rune
	word    string
	funcMap map[string]interface{}
}

type methods struct {
	area   func() string `description:"Shows game area"`
	words  func() string `description:"Shows used words"`
	step   func() string `description:"Shows name of user who's step is now"`	
	score  func() string `description:"Shows score of every user in game"`	
	help   func() string `description:"Help for you"`	
	skip   func() (bool, string, error) `description:"Command to skip (if your step is now)"`
	put    func() string `description:"Command to put letter and tell word (if your step is now)"`
}

/**
 * @brief Create a new game
 * @return game Pointer to the created Game object
 */
func NewGame(cfg conf.GameConf) *Game {
	g := &Game{square: NewSquare(cfg.AreaSize)}
	g.AreaSize = cfg.AreaSize
	g.MaxUsersPerGame = cfg.NumberUsersPerGame
	g.scoreMap = make(map[string]int)
	g.meth.area = g.area
	g.meth.words = g.words
	g.meth.step = g.step
	g.meth.score = g.score
	g.meth.help = g.help
	g.meth.skip = g.skip
	g.meth.put = g.put
	g.stepUser = 0
	g.onStart = false
	g.onPut = false

	g.putting.funcMap = make(map[string]interface{})
	g.putting.funcMap["coordX"] = g.coordX
	g.putting.funcMap["coordY"] = g.coordY
	g.putting.funcMap["letter"] = g.letter
	g.putting.funcMap["word"] = g.word

	return g
}

func (game *Game) Continue(str string, user string) (bool, string, error) {
	if !game.onStart {
		return true, "Game didn't start", nil
	}
	if str == "area" {
		return true, game.area(), nil
	}
	if str == "words" {
		return true, game.words(), nil
	}
	if str == "step" {
		return true, game.step(), nil
	}
	if str == "score" {
		return true, game.score(), nil
	}
	if str == "help" {
		return true, game.help(), nil
	}

	if game.stepUser >= len(game.users) {
		game.stepUser = game.stepUser % len(game.users)
	}
	if user != game.users[game.stepUser] {
		return true, "Not your step is now or not correct command.", nil
	}
	if str == "skip" {
		return game.skip()
	}
	if str == "put" {
		return true, game.put(), nil
	}
	if game.onPut {
		return game.putting.funcMap[game.putting.state].(func(string) (bool, string, error))(str)
	}
	return true, "Don't understand you.", nil
}

func (game *Game) AddUser(login string) error {
	if game.onStart || len(game.users) >= game.MaxUsersPerGame {
		return errors.New("Can't add user to game")
	}
	game.users = append(game.users, login)
	game.scoreMap[login] = 0
	return nil
}

func (game *Game) StartGame() {
	game.onStart = true
}

func (game *Game) FinishGame(winner string) string {
	game.onStart = false
	return winner
}

func (game *Game) area() string {
	return game.square.StrPrintArea()
}

func (game *Game) words() string {
	return game.square.StrPrintUsedWords()
}

func (game *Game) step() string {
	if game.stepUser < len(game.users) {
		return game.users[game.stepUser]
	}
	return ""
}

func (game *Game) score() string {
	str := ""
	for us, sc := range game.scoreMap {
		str = strings.Join([]string{str, us, " : ", strconv.Itoa(sc), "\n\r"}, "")
	}
	str = strings.TrimSuffix(str, "\n\r")
	return str
}

func (game *Game) help() string {
	help_messeges := []string{"Game balda"}
	m := structs.New(&methods{})
	for _, f := range structs.Names(&game.meth) {
		help_messeges = append(help_messeges, fmt.Sprintf("%s\t%s", f, m.Field(f).Tag("description")))
	}
	return strings.Join(help_messeges, "\n\r")
}

func (game *Game) skip() (bool, string, error) {
	game.skipped++
	if game.skipped == len(game.users) {
		game.FinishGame("")
		return false, "Game over. No winner. All users skipped.", nil
	}
	game.stepUser++
	if game.stepUser == len(game.users) {
		game.stepUser = 0
	}
	return true, "You skipped", nil
}

func (game *Game) put() string {
	game.onPut = true
	game.putting.state = "coordX"
	return "Entering X coordinate"
}

func (game *Game) coordX(str string) (bool, string, error) {
	i, err := strconv.Atoi(str)
	if err != nil || i < 0 || i >= game.AreaSize {
		return true, "Invalid. Try again.", nil
	}
	game.putting.x = i
	game.putting.state = "coordY"
	return true, "Entering Y coordinate", nil
}

func (game *Game) coordY(str string) (bool, string, error) {
	i, err := strconv.Atoi(str)
	if err != nil || i < 0 || i >= game.AreaSize {
		return true, "Invalid. Try again.", nil
	}
	game.putting.y = i
	game.putting.state = "letter"
	return true, "Entering rune", nil
}

func (game *Game) letter(str string) (bool, string, error) {
	if utf8.RuneCountInString(str) != 1 {
		return true, "Invalid. Try again.", nil
	}
	game.putting.sym = []rune(str)[0]
	game.putting.state = "word"
	return true, "Entering word", nil
}

func (game *Game) word(str string) (bool, string, error) {
	game.putting.word = str
	game.onPut = false
	ok := game.square.CheckWord(game.putting.y, game.putting.x, game.putting.sym, []rune(game.putting.word))
	if ok {
		game.stepUser++
		if game.stepUser == len(game.users) {
			game.stepUser = 0
		}
		sc := utf8.RuneCountInString(game.putting.word)
		game.scoreMap[game.users[game.stepUser]] += sc
		if game.square.IsFull() {
			winner := ""
			hs := 0
			for us, sc := range game.scoreMap {
				if sc > hs {
					hs = sc
					winner = us
				}
				if sc == hs {
					winner = ""
				}
			}
			game.FinishGame(winner)
			return false, strings.Join([]string{"Game over.", game.score(), "Our winner:", winner}, "\n\r"), nil
		}
		return true, strings.Join([]string{"Success", game.area()}, "\n\r"), nil
	}
	return true, "You can't add this word. Try again.", nil
}
