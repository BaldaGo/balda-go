package dict

import(
	"os"
	"bufio"
	"math/rand"
	"time"
	"unicode/utf8"
)

var wordMap map[string]bool  // Map of words from dictionary
var wordsOfAS []string       // Slice of words with AreaSize length

// as = AreaSize
// path = "dict/dictionary.txt"
// Initialization of map, slice and rand
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
		if (utf8.RuneCountInString(str) == as && !exist) {
			wordsOfAS = append(wordsOfAS, str)
		}
		wordMap[str] = true
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	return nil
}

// If word is in dictionary
func CheckWord(word string) bool {
	_, ok := wordMap[word]
	return ok
}

// Random word with AreaSize length
func RandWordOfAS() string {
	rand.Seed(time.Now().UnixNano())
	return wordsOfAS[rand.Intn(len(wordsOfAS))]
}


