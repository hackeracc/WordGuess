package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var (
	dictionaryFile = flag.String("dictionary", "dictionary.txt",
		"Absolute path of the file which contains the dictionary of words")

	// Regex for a valid word (which contains only english alphabets).
	isLetter = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

	// Global map used to store all the dictionary words. The key is the length
	// of the word and value is the list of words matching that length.
	dictionaryMap map[int][]string
)

const (
	emptyChar = '_'

	// Enums for state of the game.
	Running GameState = iota
	// User lost while playing the game.
	Lost
	// User won while playing the game.
	Won

	// Error codes while creating a new game.
	NoError InputError = iota
	InvalidRetries
	InvalidLength
)

type GameState int
type InputError int

// Game struct, new instance is created for every new game to be played.
type Game struct {
	// Expected length of the chosen word.
	ExpectedLength int
	// List of current set of words chosen by the computer.
	CurrentSetOfWords []string
	// Total retries allowed.
	AllowedRetries int
	// Current retries left.
	CurrentRetries int
	// Used characters.
	UsedChars []rune
	// Current regex shown to the user.
	// Please note we use "_" to represent a character which is yet to be guessed.
	CurrentDisplayedWord []rune
	// Current state of the game.
	State GameState
}

// ******************* Methods to init the game ************************

// Method to init the game once. It loads all the dictionary words in memory.
// Custom dictionary words can also be passed. This is mainly used for testing.
// This method should be called only once and multiple instances of the game can
// be played.
func InitGame(customWordList []string) {
	var wordList []string
	if customWordList == nil || len(customWordList) == 0 {
		// Load all the words in memory.
		data, err := ioutil.ReadFile(*dictionaryFile)
		if err != nil {
			fmt.Println("Unable to read file ", *dictionaryFile, ",error ", err)
			os.Exit(1)
		}
		wordList = strings.Split(string(data), "\n")
	} else {
		wordList = customWordList
	}
	// Sanitize the strings in the dictionary and also do preprocessing to build
	// a map where key is the length of the word and value is the slice of all
	// words of that length.
	dictionaryMap = buildLenBasedDictionary(wordList)
}

// Method to initialize one instance of a new game.
// This method returns a new instance of the game if the input is valid.
// It returns the input error code in case there was an error in the input.
func NewGame(expectedLen, maxretries int) (*Game, InputError) {
	g := &Game{
		ExpectedLength: expectedLen,
		CurrentSetOfWords: dictionaryMap[expectedLen],
		AllowedRetries: maxretries,
		CurrentRetries: maxretries,
		CurrentDisplayedWord: make([]rune, expectedLen),
		State: Running,
	}
	// Validate the expected length and allowed retries values.
	if !validateLength(expectedLen) {
		return nil, InvalidLength
	}
	if !validateNumRetries(maxretries) {
		return nil, InvalidRetries
	}
	// Initialize the current display word as all empty characters.
	for i, _ := range g.CurrentDisplayedWord {
		g.CurrentDisplayedWord[i] = emptyChar
	}
	return g, NoError
}

// ******************* Methods to play the game ************************

// Method to play the game.
// This method expects the list of words as input. This list is the list of
// words which are of the length as the "expectedLen".
// This method then takes user input for each character. It then evaluates
// whether the user input should be accepted as a valid input.
// Following could be the possible scenarios:
// 1. User input is incorrect.
// 2. User input should be accepted at some particular location(s).
// This method tries to optimize the chances of the computer winning the game by
// checking the number of options it would have with various scenarios. It tries
// to select the scenario which has maximum number of options. This reduces
// the size of the input words in each iterations for the computer. User
// input is rejected or accepted based on when the computer would have maximum
// set of remaining words to choose from.
// Method returns true if player wins and returns false if player loses.
// Params:
// char: Input character from the user. This method expects that the input character
//  is a valid alphabet.
// Returns:
// Bool: true if its a correct guess.
// error: Returns an error with the user input. Error is returned if the input is
//   not a valid alphabet or the user input was already used.
func (g *Game) CheckUserInput(char rune) (bool, error) {
	// Check if game state is not running, return.
	if g.State != Running {
		err := errors.New("Unexpected scenario: input given for a game which is not running")
		return false, err
	}
	glog.Infof("Current word list %+v, input character %d", g.CurrentSetOfWords, char)
	if contains(g.UsedChars, char) {
		err := fmt.Errorf("Character %s has been used. " +
			"Please enter a new character.", string(char))
		return false, err
	}
	g.UsedChars = append(g.UsedChars, char)
	// Get the group with max possibilities.
	newSet, newRegex := getMaxSet(g.CurrentSetOfWords,
		g.CurrentDisplayedWord, char)
	g.CurrentSetOfWords = newSet
	glog.Infof("New word list after processing character %s: %v", string(char), g.CurrentSetOfWords)
	// Check if the new regex is same as the previous regex which means input was
	// not accepted.
	if newRegex == string(g.CurrentDisplayedWord) {
		// Reduce the retries only if its an incorrect guess.
		g.CurrentRetries --
		if g.CurrentRetries < 0 {
			g.State = Lost
		}
		return false, nil
	}
	g.CurrentDisplayedWord = []rune(newRegex)
	if !contains(g.CurrentDisplayedWord, emptyChar) {
		g.State = Won
		return true, nil
	}
	return true, nil
}

// Method to get the max set.
// Params:
// wordList: List of words from which the program can chose any word as the secret word.
// currWord: This is the string representation of the current word shown to the
//   user. Please note we use "_" to represent a character which is not yet guessed.
// char: Current input character from the user.
//
// Returns:
// 1. The new set of words which can be the potential candidates based on user input.
//    This will be based on evaluating all possibilities whether the user input
//    character is accepted or not.
// 2. The string representation of the word to be shown to the user after the
//    program has made a best decision whether the input character is to be accepted
//    or not.
func getMaxSet(wordList []string, currWord []rune, char rune) ([]string, string) {
	// Map to store all the possibilities. Possibilities can be:
	// 1. The input character is not accepted.
	// 2. The input character is accepted at a particular location.
	// This map stores the various possibilities as key (in form of a string) and
	// value is the list of words if that possibility is chosen.
	possiblitiesMap := make(map[string][]string)
	// Variable to store the length of the maximum set formed in the possibilitesMap.
	var maxSetLength int
	// Variable to store the possibility which has the maximum length as value
	// in the map possibilitiesMap.
	var maxSet string
	for _, word := range wordList {
		glog.Infof("Checking string %s, current input character %v", word, string(char))
		if !strings.ContainsRune(word, char) {
			glog.Infof("String does not contain rune")
			if _, ok := possiblitiesMap[string(currWord)]; ok{
				possiblitiesMap[string(currWord)] = append(possiblitiesMap[string(currWord)], word)
			} else {
				possiblitiesMap[string(currWord)] = []string{word}
			}
			if len(possiblitiesMap[string(currWord)]) > maxSetLength {
				maxSet = string(currWord)
				maxSetLength = len(possiblitiesMap[string(currWord)])
			}
		} else {
			// Character is present in the word.
			// Check the regex if the character is present.
			modifiedInput := make([]rune, len(currWord))
			copy(modifiedInput, currWord)
			for idx, wordChar := range word {
				if wordChar == char {
					modifiedInput[idx] = wordChar
				}
			}
			modifiedInputRegex := string(modifiedInput)
			if _, ok := possiblitiesMap[modifiedInputRegex]; ok {
				possiblitiesMap[modifiedInputRegex] = append(possiblitiesMap[modifiedInputRegex], word)
			} else {
				possiblitiesMap[modifiedInputRegex] = []string{word}
			}
			if len(possiblitiesMap[modifiedInputRegex]) > maxSetLength {
				maxSet = modifiedInputRegex
				maxSetLength = len(possiblitiesMap[modifiedInputRegex])
			}
		}
	}

	// Now that we have the max set, we can find if there is another set of the
	// same length which reveals less number of alphabets to the user.
	for possibility, possibilityWords := range possiblitiesMap {
		if len(possibilityWords) == maxSetLength {
			// Calculate number of hidden characters in both possibilities.
			n1 := strings.Count(possibility, string(emptyChar))
			n2 := strings.Count(maxSet, string(emptyChar))
			if (n1 > n2) {
				maxSet = possibility
			} else if n1 == n2 {
				// If both the possibilities reveal the same amount of characters,
				// we can pick the lexicographically smaller string. This is an
				// assumption that if user finds the first (or any of the first
				// few) character, it will be easier to guess the word.
				if possibility < maxSet {
					maxSet = possibility
				}
			}
		}
	}
	// The maxSet contains the regex for the largest length..
	glog.Infof("Possibilities map %+v", possiblitiesMap)
	glog.Infof("Max set %v", maxSet)
	return possiblitiesMap[maxSet], maxSet
}

// ********************  Preprocessing methods ************************

// Method to build a map where key is the length and value is the list of words
// for that length.
// This method also validates each word before adding it in memory.
// This method also converts all the words to lower case since our hangman is not
// case sensitive.
func buildLenBasedDictionary(wordList []string) map[int][]string {
	wordMap := make(map[int][]string)
	for _, word := range wordList {
		isValid := validateWord(word)
		if !isValid {
			glog.Errorf("Discarding word %s since it has some invalid characters", word)
		}
		wordMap[len(word)] = append(wordMap[len(word)], word)
	}
	return wordMap
}

// **************************  Validators *****************************

// Method to validate if there is any word in the wordList with the length
// "expectedLen".
func validateLength(expectedLen int) bool {
	if _, ok := dictionaryMap[expectedLen]; ok {
		return true
	}
	return false
}

// Validate the number of retries given as an input.
func validateNumRetries(retries int) bool {
	if retries < 0 || retries > *maxAllowedRetries {
		return false
	}
	return true
}

// Method to validate a word.
func validateWord(word string) bool {
	if isLetter(word) {
		return true
	}
	return false
}

// *************************  Helper methods ***************************

// Read a single character from stdin. This method also validates if its a valid
// character and does not return till a valid character is given as an input.
func readChar() rune {
	var char rune
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanned := scanner.Scan()
		for !scanned {
			scanned = scanner.Scan()
		}
		str := scanner.Text()
		if len(str) != 1 {
			fmt.Println("Invalid character, please input the character again")
			continue
		}
		char = rune(str[0])
		// Check if its a character.
		if !unicode.IsLetter(char) {
			fmt.Println("Invalid character, please input the character again")
			continue
		}
		break
	}
	return char
}

// Method to check if a slice of rune elements contains a particular character.
func contains(arr []rune, expectedChar rune) bool {
	for _, char := range arr {
		if expectedChar == char {
			return true
		}
	}
	return false
}
