/**
 * @file dict.go
 * @brief Dictionary of allowed words
 *
 * Stores the dictionary of valid words in the game
 */

package dict

import (
	//System
	"bufio"
	"math/rand"
	"os"
	"time"
	"unicode/utf8"
	// Third-party
	// Project
)

var wordMap map[string]bool // Map of words from dictionary
var wordsOfAS []string      // Slice of words with AreaSize length

/**
 * @brief Initialization of wordMap, wordsOfAS and rand
 * @param[in] as Length side of the playing area
 * @param[in] path Relative path to the dictionary
 * @return err Error if it occured
 *
 * Reads words from the dictionary, fill wordsMap and wordsOfAS
 */
func Init(as int, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	wordMap = make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		_, exist := wordMap[str]
		if utf8.RuneCountInString(str) == as && !exist {
			wordsOfAS = append(wordsOfAS, str)
		}
		wordMap[str] = true
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

/**
 * @brief Predicate, check if word is in dictionary
 * @param[in] word Checking word
 * @return ok If ok is true, then word exists in dict
 */
func CheckWord(word string) bool {
	_, ok := wordMap[word]
	return ok
}

/**
 * @bried Return random word with AreaSize length
 * @return word Random word with AreaSize length
 */
func RandWordOfAS() string {
	rand.Seed(time.Now().UnixNano())
	return wordsOfAS[rand.Intn(len(wordsOfAS))]
}
