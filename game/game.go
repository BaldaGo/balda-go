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
	"github.com/BaldaGo/balda-go/db"
	"github.com/fatih/structs"
)

const databaseError string = "DATABASE_ERROR"

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
	dbGameID        uint
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

	stat_topusers func(string, int, int) (bool, string, error) `description:"Shows top of users. Parameters: mode(score, games, wins), limit"`
	stat_topwords func(int, int) (bool, string, error) `description:"Shows top of words. Parameters: limit"`
	stat_wordtopusers func(string, int, int) (bool, string, error) `description:"Shows top of users used this word. Parameters: word, limit"`
	stat_user func(string, int, int) (bool, string, error) `description:"Shows top of users. Parameters: username, limit"`
}

/**
 * @brief Create a new game
 * @return game Pointer to the created Game object
 */

func NewGame(cfg conf.GameConf) (*Game, error) {
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
	g.meth.stat_topusers = g.GetTopUsersByMode
	g.meth.stat_topwords = g.GetTopWords
	g.meth.stat_wordtopusers = g.GetWordTopUsers
	g.meth.stat_user = g.GetUserAllGamesStat
	g.stepUser = 0
	g.onStart = false
	g.onPut = false
	g.putting.funcMap = make(map[string]interface{})
	g.putting.funcMap["coordX"] = g.coordX
	g.putting.funcMap["coordY"] = g.coordY
	g.putting.funcMap["letter"] = g.letter
	g.putting.funcMap["word"] = g.word
  res, err := db.StartGame()
	if err != nil {
		return nil, err
	}
	g.dbGameID = res.ID

	return g, nil
}

func (game *Game) Continue(str string, user string) (bool, string, error) {
	arr := strings.Split(str, " ")
	if arr[0] == "stat_topusers" {
		n, err := strconv.Atoi(arr[2])
		if err != nil {
			return true, "Not correct command, not integer in limit", err
		}
		if arr[1] != "score" && arr[1] != "games" && arr[1] != "wins" {
			return true, "Not correct command, bad mode. You must use one of: score, games, wins.", err
		}
		return game.meth.stat_topusers(arr[1], n, 0)
	}
	if arr[0] == "stat_topwords" {
		n, err := strconv.Atoi(arr[1])
		if err != nil {
			return true, "Not correct command, not integer in limit", err
		}
		return game.meth.stat_topwords(n, 0)
	}
	if arr[0] == "stat_wordtopusers" {
		n, err := strconv.Atoi(arr[2])
		if err != nil {
			return true, "Not correct command, not integer in limit", err
		}
		return game.meth.stat_wordtopusers(arr[1], n, 0)
	}
	if arr[0] == "stat_user" {
		n, err := strconv.Atoi(arr[2])
		if err != nil {
			return true, "Not correct command, not integer in limit", err
		}
		return game.meth.stat_user(arr[1], n, 0)
	}

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

	if _, err := db.NewUserInSession(login, game.dbGameID); err != nil{
		return err
	}
	game.scoreMap[login] = 0
	return nil
}

func (game *Game) StartGame() error {
	game.onStart = true


	return nil
}

func (game *Game) FinishGame(winner string) (string, error) {
	game.onStart = false
	err := db.GameOver(game.scoreMap, game.dbGameID, winner)
	if err != nil{
		return databaseError, err
	}

	return winner, nil
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
		sc := utf8.RuneCountInString(game.putting.word)
		nowPlayer := game.users[game.stepUser]
		game.scoreMap[nowPlayer] += sc

		if _, err := db.AddWord(nowPlayer, str); err != nil{
			return false, databaseError, err
		}
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
		game.stepUser++
		if game.stepUser == len(game.users) {
			game.stepUser = 0
		}
		return true, strings.Join([]string{"Success", game.area()}, "\n\r"), nil
	}
	return true, "You can't add this word. Try again.", nil
}

func (game *Game) GetTopUsersByMode(mode string, limit int, offset int) (bool, string, error){

	res, err := db.GetTop(mode, uint(limit), uint(offset))
	if err != nil {
		return true, databaseError, err
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

	return true, pretty, nil
}

func (game *Game) GetTopWords(limit int, offset int) (bool, string, error){

	res, err := db.TopWords(uint(limit), uint(offset))
	if err != nil {
		return true, databaseError, err
	}
	var prepare []string
	for i := range res{
		prepare = append(prepare,
			fmt.Sprintf("Word: %s, Usage count: %d",
				res[i].Word,
				res[i].Popularity))
	}

	pretty := strings.Join(prepare, "\n\r")
	return true, pretty, nil
}

func (game *Game) GetWordTopUsers(word string, limit int, offset int) (bool, string, error){

	res, err := db.WordTopUsers(word, uint(limit), uint(offset))
	if err != nil {
		return true, databaseError, err
	}

	var prepare []string
	for key, value := range res{
		prepare = append(prepare,
			fmt.Sprintf("User: %s, Usage count: %d",
				key,
				value))
	}

	pretty := strings.Join(prepare, "\n\r")
	return true, pretty, nil
}

func (game *Game) GetUserAllGamesStat(username string, limit int, offset int) (bool, string, error){

	res, err := db.UserAllGamesStat(username, uint(limit), uint(offset))
	if err != nil {
		return true, databaseError, err
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
	return true, pretty, nil
}
