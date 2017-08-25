/**
 *
 * @file db.go
 * @brief Database
 *
 * Database API, consist of game process operations
 * and methods for get statistics.
 *
 * Important! Every table (class (struct)) of database
 * has a GORM (gorm.Model) build-in fields which contains:
 * 		1) Auto increment primary key ID
 * 		2) Field creating date.
 * 		3) Field updating date.
 * 		4) Field deleting date. (Read more about deleting in GORM
 * 								 http://jinzhu.me/gorm/crud.html#delete)
 */

package db

import (
	// System
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"strings"

	// Third-party
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"

	// Project
	"github.com/BaldaGo/balda-go/conf"
	"github.com/BaldaGo/balda-go/logger"
)

/**
 *
 * @class Database
 * @brief db object
 */

var db *gorm.DB // Database main variable
var dictSize uint

/**
 *
 * @class User
 * @brief Main table
 *
 * A table containing the basic information about each
 * player: login, password hash, ip address, number of games,
 * wins, points scored, lexicon size.
 */
type User struct {
	gorm.Model

	Name       string `gorm:"unique"`
	Wins       uint   `gorm:"default:0"`
	Password   uint32
	IpAddr     string
	Games      uint `gorm:"default:0"`
	Scores     uint `gorm:"default:0"`
	WordsCount uint `gorm:"default:0"`
}

/**
 *
 * @class UserConnection
 * @brief Table about from where (ip-address) came user.
 *
 */
type UserConnection struct {
	gorm.Model

	UserID uint
	IpAddr string
	User   User `gorm:"ForeignKey:UserID"`
}

/**
 *
 * @class RusWord
 * @brief This table contains a complete dictionary of the Russian language (120,000 words)
 *
 * A table loading takes 60 seconds when the server starts.
 */
type RusWord struct {
	gorm.Model

	Word       string
	Popularity uint `gorm:"default:0"`
}

/**
 *
 * @class GameSession
 * @brief The table contains information about past, current games and winners.
 *
 * Winner field is null if game is not end yet.
 */
type GameSession struct {
	gorm.Model

	WinnerID uint
	Winner   User `gorm:"ForeignKey:WinnerID"`
}

/**
 *
 * @class UsersLexicon
 * @brief The table contains the history of words that the player offered.
 * So you can make an estimate of his vocabulary.
 *
 * If a player uses the same word multiple times, the counter gives you information about it.
 */
type UsersLexicon struct {
	gorm.Model

	UserID    uint
	RusWordID uint
	Count     uint    `gorm:"default:1"`
	User      User    `gorm:"ForeignKey:UserID"`
	RusWord   RusWord `gorm:"ForeignKey:RusWordID"`
}

/**
 *
 * @class UserInGame
 * @brief The table contains the history of the games in which
 * each player was. Also, it has its final score.
 *
 */
type UserInGame struct {
	gorm.Model

	UserID      uint
	Score       uint
	GameID      uint
	User        User        `gorm:"ForeignKey:UserID"`
	GameSession GameSession `gorm:"ForeignKey:GameID"`
}

/**
 *
 * @brief Connecting to db.
 * Create or update tables from structs if they are not exist already.ew
 * @param[in] Database config
 * @return error
 *
 */
func Init(cfg conf.DatabaseConf) error {

	var err error
	var parsedOptions string
	options := []string{}

	for key, value := range cfg.Options {
		options = append(options, fmt.Sprintf("%s=%s", key, value))
	}

	parsedOptions = strings.Join(options, "&")

	if db, err = gorm.Open(cfg.Dialect,
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Name,
			parsedOptions)); err != nil {
		return err
	}

	if res := db.Set("gorm:insert_options", fmt.Sprintf("ENGINE=%s", cfg.Engine)).
		AutoMigrate(&User{},
			&RusWord{},
			&GameSession{},
			&UsersLexicon{},
			&UserInGame{},
			&UserConnection{}); res != nil {
		return res.Error
	}
	return nil
}

/**
 *
 * @brief Loading the dictionary of the Russian language.
 * @param[in] path to dictionary txt file.
 * @return error
 *
 * Uploading takes around of 1 minute
 */
func LoadDictionary(path string) error {

	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)

	logger.Log.Info("Loading dictionary. Please wait... (approximately 60 seconds)")

	for scanner.Scan() {
		dictSize++
		var word = RusWord{Word: scanner.Text()}
		if res := db.Create(&word); res.Error != nil {
			return res.Error
		}
	}

	logger.Log.Info(fmt.Sprintf("Done. %d words uploaded. Database is ready for game", dictSize))

	return nil
}

/**
 *
 * @brief Get hash of line.
 * @param[in] s line to hashing.
 * @return 32 bit integer value.
 *
 * Used to encrypt a user's password.
 */
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()

}

/**
 *
 * @brief Add new user to db.
 * @param[in] username
 * @param[in] password
 * @param[in] ip-addr
 * @return the record just created for the new user.
 * @return error
 *
 */
func AddUser(username string, password string, ip string) (*User, error) {

	newUser := User{Name: username, Password: hash(password), IpAddr: ip}
	if res := db.Create(&newUser); res.Error != nil {
		return nil, res.Error
	}
	return &newUser, nil
}

/**
 *
 * @brief Valid username - password login.
 * @param[in] username of user
 * @param[in] password of user
 * @param[in] ip address of user
 * @return true if valid, false if invalid login
 * @return error
 *
 */
func CheckUser(username string, password string) (bool, error) {

	user := &User{}
	if res := db.
		Where("name = ? and password = ?", username, hash(password)).
		Find(&user); res.Error != nil {
		return false, res.Error
	}
	return true, nil
}

/**
 *
 * @brief Adds the user's id and the IP address with which it came.
 * @param[in] username of user
 * @param[in] ip of user
 * @return the record just created for the new user connection.
 * @return error
 *
 */
func AddUserConnection(username string, ip string) (*UserConnection, error) {

	user := User{}
	if res := db.Where("name = ?", username).First(&user); res.Error != nil {
		return nil, res.Error
	}

	newUserConnection := UserConnection{UserID: user.ID, IpAddr: ip}
	if res := db.Create(&newUserConnection); res.Error != nil {
		return nil, res.Error
	}

	return &newUserConnection, nil
}

/**
 *
 * @brief Create new game with empty winner.
 * @return the record just created for the new game.
 * @return error
 *
 */
func StartGame() (*GameSession, error) {

	gameSession := GameSession{}
	if res := db.Create(&gameSession); res.Error != nil {
		return nil, res.Error
	}
	return &gameSession, nil
}

/**
 *
 * @brief Adds a user to the game to which he connected
 * @param[in] username of user
 * @param[in] sessionID of game that was returned to game logic from StartGame
 * @return the record just created for the new [user in game] connection.
 * @return error
 *
 */
func NewUserInSession(username string, sessionID uint) (*UserInGame, error) {

	user := User{}
	if res := db.Where("name = ?", username).First(&user); res.Error != nil {
		return nil, res.Error
	}

	newUserInGame := UserInGame{UserID: user.ID, GameID: sessionID}
	if res := db.Create(&newUserInGame); res.Error != nil {
		return nil, res.Error
	}

	return &newUserInGame, nil
}

/**
 *
 * @brief Drops a user from the game session in which he is.
 * @param[in] username of user
 * @return the record just created for the new [user in game] connection.
 * @return error
 *
 * Finds the last user session (single) and delete it.
 */
func RemoveUserFromSession(username string) (*UserInGame, error) {

	user := User{}
	if res := db.Where("name = ?", username).First(&user); res.Error != nil {
		return nil, res.Error
	}

	// TODO: try to change line 312 to "...user.ID).First(&session).Order("ID DESC").Delete... "
	session := UserInGame{}
	if res := db.
		Where("user_id = ?", user.ID).
		Find(&session).
		Order("ID").
		Delete(&session); res.Error != nil {
		return nil, res.Error
	}

	return &session, nil
}

/**
 *
 * @brief Adds word to user's personal lexicon vocabulary.
 * @param[in] username of user
 * @param[in] new word
 * @return the record just created for the new user's word.
 * @return error
 *
 * Increments userslexicon value if word was used already
 */
func AddWord(username string, word string) (*UsersLexicon, error) {

	user := User{}
	if res := db.Where("name = ?", username).First(&user); res.Error != nil {
		return nil, res.Error
	}
	rusWord := RusWord{}
	if res := db.Where("word = ?", word).First(&rusWord); res.Error != nil {
		return nil, res.Error
	}

	rusWord.Popularity++

	userLexicon := UsersLexicon{}
	if res := db.
		Where("user_id = ? and rus_word_id = ?", user.ID, rusWord.ID).
		First(&userLexicon); res.Error != nil {

		userLexicon = UsersLexicon{UserID: user.ID, RusWordID: rusWord.ID}
		if res := db.Create(&userLexicon); res.Error != nil {
			return nil, res.Error
		}

		user.WordsCount++
		if res := db.Save(&user); res.Error != nil {
			return nil, res.Error
		}

	} else {
		userLexicon.Count++
		if res := db.Save(&userLexicon); res.Error != nil {
			return nil, res.Error
		}
	}
	if res := db.Save(&rusWord); res.Error != nil {
		return nil, res.Error
	}
	return &userLexicon, nil
}

/**
 *
 * @brief Ends the game and earns points.
 * @param[in] game final statistics which contains players scores and info about winner.
 * @param[in] game session id returned from start game method
 * @return error
 *
 * for all players scores += this game scores
 * for all players games ++
 * for winner wins ++
 */
func GameOver(gameStatistics map[string][2]uint, gameID uint) error {

	for key, value := range gameStatistics {

		user := User{}
		if res := db.Where("name = ?", key).First(&user); res.Error != nil {
			return res.Error
		}
		user.Games++
		user.Scores += value[0]

		if value[1] == 1 {
			user.Wins++
			gameSession := GameSession{}
			if res := db.Where("id = ?", gameID).First(&gameSession); res.Error != nil {
				return res.Error
			}
			gameSession.WinnerID = user.ID
			if res := db.Save(&gameSession); res.Error != nil {
				return res.Error
			}
		}

		userInGame := UserInGame{}
		if res := db.
			Where("user_id = ? and game_id = ?", user.ID, gameID).
			First(&userInGame); res.Error != nil {
			return res.Error
		}
		userInGame.Score = value[0]
		if res := db.Save(&userInGame); res.Error != nil {
			return res.Error
		}
		if res := db.Save(&user); res.Error != nil {
			return res.Error
		}

	}
	return nil
}

/**
 *
 * @brief Normalizing of limit and offset if any of these out of range.
 * @param[in] range size
 * @param[in] limit by link (or pointer i don't understand what the fuck is going on with memory in golang)
 * @param[in] offset by link (i really don't give a shit what is it. Tha main thing is -  it works.)
 * @return error
 *
 */
func normalizeLimitAndOrder(lenOfTable uint, limit *uint, offset *uint) {

	if *offset >= lenOfTable {
		*offset = 0
	}
	if *limit > lenOfTable-*offset {
		*limit = (lenOfTable - *offset) % *limit
	}
}

/**
 *
 * @brief Get top of users by one of their fields.
 * @param[in] mode of sorting (scores, games, wins)
 * @param[in] limit
 * @param[in] offset
 * @return slice of users objects
 * @return error
 *
 */
func GetTop(mode string, limit uint, offset uint) ([]User, error) {

	top := []User{}

	if res := db.Order(fmt.Sprintf("%s desc", mode)).Find(&top); res.Error != nil {
		return nil, res.Error
	}
	normalizeLimitAndOrder(uint(len(top)), &limit, &offset)

	if res := db.
		Order(fmt.Sprintf("%s desc", mode)).
		Limit(limit).
		Offset(offset).
		Find(&top); res.Error != nil {
		return nil, res.Error
	}
	return top, nil
}

/**
 *
 * @brief Get top of the most popular words.
 * @param[in] limit
 * @param[in] offset
 * @return slice of word objects
 * @return error
 *
 */
func TopWords(limit uint, offset uint) ([]RusWord, error) {

	top := []RusWord{}

	normalizeLimitAndOrder(dictSize, &limit, &offset)
	if res := db.
		Order("popularity DESC").
		Limit(limit).
		Offset(offset).
		Find(&top); res.Error != nil {
		return nil, res.Error
	}

	return top, nil
}

/**
 *
 * @brief Return the top players who most often use this word.
 * @param[in] word
 * @param[in] limit
 * @param[in] offset
 * @return map[username of player]number of word uses
 * @return error
 *
 */
func WordTopUsers(word string, limit uint, offset uint) (map[string]uint, error) {

	topLexicons := []UsersLexicon{}

	rusWordField := RusWord{}
	if res := db.Where("word = ?", word).Find(&rusWordField); res.Error != nil {
		return nil, res.Error
	}

	if res := db.Where("rus_word_id = ? ", rusWordField.ID).Find(&topLexicons); res.Error != nil {
		return nil, res.Error
	}
	normalizeLimitAndOrder(uint(len(topLexicons)), &limit, &offset)

	if res := db.
		Where("rus_word_id = ? ", rusWordField.ID).
		Order("count DESC").
		Limit(limit).
		Offset(offset).
		Find(&topLexicons); res.Error != nil {
		return nil, res.Error
	}
	topUsers := make(map[string]uint)

	for i := range topLexicons {
		if res := db.Model(&(topLexicons[i])).Related(&(topLexicons[i]).User); res.Error != nil {
			return nil, res.Error
		}
		topUsers[topLexicons[i].User.Name] = topLexicons[i].Count
	}

	return topUsers, nil
}

/**
 *
 * @brief Returns the list of players in the current game
 * @param[in] username of player
 * @return slice of user objects
 * @return error
 *
 */
func GetCurrentGameUsersList(username string) ([]User, error) {

	lastGame := UserInGame{}

	user := User{}
	if res := db.Where("name = ?", username).First(&user); res.Error != nil {
		return nil, res.Error
	}

	if res := db.Where("user_id = ?", user.ID).Find(&lastGame); res.Error != nil {
		return nil, res.Error
	}

	allCurrentGamePlayersSessions := []UserInGame{}
	if res := db.
		Where("game_id = ?", lastGame.GameID).
		Preload("User").
		Find(&allCurrentGamePlayersSessions); res.Error != nil {
		return nil, res.Error
	}
	resultUsers := []User{}
	for i := range allCurrentGamePlayersSessions {
		resultUsers = append(resultUsers, allCurrentGamePlayersSessions[i].User)
	}

	return resultUsers, nil
}

type gameFullStat struct {
	winner string
	Users  []User
}

/**
 *
 * @brief Returns the full statistics about all user's games
 * @param[in] username of player
 * @return map of gameFullStat structs for every game
 * @return error
 *
 */
func GetUserAllGamesStat(username string, limit uint, offset uint) (map[uint]gameFullStat, error) {

	result := make(map[uint]gameFullStat)

	user := User{}
	if res := db.Where("name = ?", username).First(&user); res.Error != nil {
		return nil, res.Error
	}

	normalizeLimitAndOrder(user.Games, &limit, &offset)

	userGamesList := []UserInGame{}
	if res := db.
		Where("user_id = ?", user.ID).
		Preload("GameSession.Winner").
		Limit(limit).
		Offset(offset).
		Find(&userGamesList); res.Error != nil {
		return nil, res.Error
	}

	for i := range userGamesList {

		anotherUsersInThisGame := []UserInGame{}

		if res := db.
			Where("game_id = ?", userGamesList[i].GameID).
			Preload("User").
			Find(&anotherUsersInThisGame); res.Error != nil {
			return nil, res.Error
		}

		usersList := []User{}
		for j := range anotherUsersInThisGame {
			usersList = append(usersList, anotherUsersInThisGame[j].User)
		}
		result[userGamesList[i].GameID] = gameFullStat{winner: userGamesList[i].GameSession.Winner.Name, Users: usersList}

	}
	return result, nil
}
