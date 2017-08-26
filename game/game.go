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

	// Third-party

	// Project
	"github.com/BaldaGo/balda-go/conf"
	"github.com/BaldaGo/balda-go/db"
	"fmt"
)

const databaseError string = "DATABASE_ERROR"

/**
 * @class Game
 * @brief Class, provide information about concrete game
 */
type Game struct {
	square          Square ///< Gaming area
	funcMap         map[string]interface{}
	users           []string
	score           map[string]int
	step            int
	onStart         bool
	skipped         int
	dbGameID        uint
	putting         Put
	onPut           bool
	AreaSize        int
	MaxUsersPerGame int
}

type Put struct {
	state   string
	x       int
	y       int
	sym     rune
	word    string
	funcMap map[string]interface{}
}

/**
 * @brief Create a new game
 * @return game Pointer to the created Game object
 */

func NewGame(cfg conf.GameConf) (*Game, error) {
	g := &Game{square: NewSquare(cfg.AreaSize)}
	g.AreaSize = cfg.AreaSize
	g.MaxUsersPerGame = cfg.NumberUsersPerGame
	g.funcMap = make(map[string]interface{})
	g.score = make(map[string]int)
	g.funcMap["area"] = g.PrintArea
	g.funcMap["words"] = g.printUsedWords
	g.funcMap["step"] = g.whoStep
	g.funcMap["score"] = g.showScore
	g.funcMap["skip"] = g.skip
	g.funcMap["put"] = g.put
	g.step = 0
	res, err := db.StartGame()
	if err != nil {
		return nil, err
	}
	g.dbGameID = res.ID
	g.onStart = false
	g.onPut = false
	g.putting.funcMap = make(map[string]interface{})
	g.putting.funcMap["coordX"] = g.coordX
	g.putting.funcMap["coordY"] = g.coordY
	g.putting.funcMap["letter"] = g.letter
	g.putting.funcMap["word"] = g.word

	return g, nil
}

func (game *Game) Continue(str string, user string) (bool, string) {
	if !game.onStart {
		return true, "Game didn't start"
	}
	if str == "area" || str == "words" || str == "step" || str == "score" {
		return true, game.funcMap[str].(func() string)()
	}
	if game.step >= len(game.users) {
		game.step = game.step % len(game.users)
	}
	if user != game.users[game.step] {
		return true, "Not your step is now or not correct command."
	}
	if str == "skip" {
		return game.funcMap[str].(func() (bool, string))()
	}
	if str == "put" {
		return true, game.funcMap[str].(func() string)()
	}
	if game.onPut {
		return game.putting.funcMap[game.putting.state].(func(string) (bool, string))(str)
	}
	return true, "Don't understand you."
}

func (game *Game) AddUser(login string) error {
	if game.onStart || len(game.users) >= game.MaxUsersPerGame {
		return errors.New("Can't add user")
	}
	game.users = append(game.users, login)
	game.score[login] = 0

	if _, err := db.NewUserInSession(login, game.dbGameID); err != nil{
		return err
	}

	return nil
}

func (game *Game) StartGame() error {
	game.onStart = true


	return nil
}

func (game *Game) FinishGame(winner string) (string, error) {
	game.onStart = false
	err := db.GameOver(game.score, game.dbGameID, winner)
	if err != nil{
		return databaseError, err
	}

	return winner, nil
}

func (game *Game) PrintArea() string {
	return game.square.StrPrintArea()
}

func (game *Game) printUsedWords() string {
	return game.square.StrPrintUsedWords()
}

func (game *Game) whoStep() string {
	if game.step < len(game.users) {
		return game.users[game.step]
	}
	return ""
}

func (game *Game) showScore() string {
	str := ""
	//fmt.Println("IN shoeScore")
	for us, sc := range game.score {
		//fmt.Println("iterating")
		str = strings.Join([]string{str, us, " : ", strconv.Itoa(sc), "\n"}, "")
	}
	str = strings.TrimSuffix(str, "\n")
	return str
}

func (game *Game) skip() (bool, string) {
	game.skipped++
	if game.skipped == len(game.users) {
		game.FinishGame("")
		return false, "Game over. No winner. All users skipped."
	}
	game.step++
	if game.step == len(game.users) {
		game.step = 0
	}
	return true, "You skipped"
}

func (game *Game) put() string {
	game.onPut = true
	game.putting.state = "coordX"
	return "Please, enter X coordinate"
}

func (game *Game) coordX(str string) (bool, string) {
	i, err := strconv.Atoi(str)
	if err != nil || i < 0 || i >= game.AreaSize {
		return true, "Invalid. Try again."
	}
	game.putting.x = i
	game.putting.state = "coordY"
	return true, "Please, enter Y coordinate"
}

func (game *Game) coordY(str string) (bool, string) {
	i, err := strconv.Atoi(str)
	if err != nil || i < 0 || i >= game.AreaSize {
		return true, "Invalid. Try again."
	}
	game.putting.y = i
	game.putting.state = "letter"
	return true, "Please, enter rune"
}

func (game *Game) letter(str string) (bool, string) {
	if utf8.RuneCountInString(str) != 1 {
		return true, "Invalid. Try again."
	}
	game.putting.sym = []rune(str)[0]
	game.putting.state = "word"
	return true, "Please, enter word"
}

func (game *Game) word(str string) (bool, string, error) {

	game.putting.word = str
	game.onPut = false
	ok := game.square.CheckWord(game.putting.y, game.putting.x, game.putting.sym, []rune(game.putting.word))
	if ok {
		game.step++
		if game.step == len(game.users) {
			game.step = 0
		}
		sc := utf8.RuneCountInString(game.putting.word)
		nowPlayer := game.users[game.step]
		game.score[nowPlayer] += sc

		if _, err := db.AddWord(nowPlayer, str); err != nil{
			return false, databaseError, err
		}

		if game.square.IsFull() {
			winner := ""
			hs := 0
			for us, sc := range game.score {
				if sc > hs {
					hs = sc
					winner = us
				}
				if sc == hs {
					winner = ""
				}
			}
			game.FinishGame(winner)
			return false, strings.Join([]string{"Game over.", game.showScore(), "Our winner:", winner}, "\n"), nil
		}
		return true, strings.Join([]string{"Success", game.PrintArea()}, "\n"), nil
	}
	return true, "You can't add this word. Try again.", nil
}

func (game *Game) GetTopUsersByMode(mode string, limit int, offset int) (string, error){

	res, err := db.GetTop(mode, uint(limit), uint(offset))
	if err != nil {
		return databaseError, err
	}
	var prepare []string
	for i := range res{
		prepare = append(prepare,
			fmt.Sprintf("Login: %s, Scores: %d, Games: %d, Wins: %d",
				res[i].Name,
				res[i].Scores,
				res[i].Games,
				res[i].Wins))
	}

	pretty := strings.Join(prepare, "\n\r")

	return pretty, nil
}

func (game *Game) GetTopWords(limit int, offset int) (string, error){

	res, err := db.TopWords(uint(limit), uint(offset))
	if err != nil {
		return databaseError, err
	}
	var prepare []string
	for i := range res{
		prepare = append(prepare,
			fmt.Sprintf("Word: %s, Usage count: %d",
				res[i].Word,
				res[i].Popularity))
	}

	pretty := strings.Join(prepare, "\n\r")
	return pretty, nil
}

func (game *Game) GetWordTopUsers(word string, limit int, offset int) (string, error){

	res, err := db.WordTopUsers(word, uint(limit), uint(offset))
	if err != nil {
		return databaseError, err
	}

	var prepare []string
	for key, value := range res{
		prepare = append(prepare,
			fmt.Sprintf("User: %s, Usage count: %d",
				key,
				value))
	}

	pretty := strings.Join(prepare, "\n\r")
	return pretty, nil
}

func (game *Game) GetUserAllGamesStat(username string, limit int, offset int) (string, error){

	res, err := db.UserAllGamesStat(username, uint(limit), uint(offset))
	if err != nil {
		return databaseError, err
	}

	var prepare []string
	for key, value := range res{
		var usersLocalList []string
		for j := range value.Users{
			usersLocalList = append(usersLocalList,
				fmt.Sprintf("\t\tUser: %s, Scores: %d",
					value.Users[j].Name,
					value.Users[j].Scores))
		}
		anotherUsers := strings.Join(usersLocalList, "\n\r")
		prepare = append(prepare,
			fmt.Sprintf("GameID: %d, Winner: %s \n\rAnother players: \n\r%s",
				key,
				value.Winner,
				anotherUsers))
	}

	pretty := strings.Join(prepare, "\n\r\n\r")
	return pretty, nil
}

