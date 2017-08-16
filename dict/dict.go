package dict

import utf8

var WordMap = map[string]bool{}
var WordsOfAS = []string{}

func Init(as int) {
	return nil
}

func CheckWord(word string) bool {
	_, ok := WordMap[word]
	return ok
}

func RandWordOfAS() string {
	return ""
}


