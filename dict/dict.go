package dict

import(
	"os"
	"log"
	"bufio"
	"math/rand"
	"time"
	"unicode/utf8"
)

var WordMap map[string]bool  // Map of words from dictionary
var WordsOfAS []string       // Slice of words with AreaSize length
var source rand.Source
var randPtr *rand.Rand

// as = AreaSize
// path = "dict/dictionary.txt"
// Initialization of map, slice and rand
func Init(as int, path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	WordMap = make(map[string]bool)
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		str := scanner.Text()
		_, ok := WordMap[str]
		if (utf8.RuneCountInString(str) == as && !ok) {
			WordsOfAS = append(WordsOfAS, str)
		}
		WordMap[str] = true
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	
	source = rand.NewSource(time.Now().UnixNano())
	randPtr = rand.New(source)
}

// If word is in dictionary
func CheckWord(word string) bool {
	_, ok := WordMap[word]
	return ok
}

// Random word with AreaSize length
func RandWordOfAS() string {
	return WordsOfAS[randPtr.Intn(len(WordsOfAS))]
}


