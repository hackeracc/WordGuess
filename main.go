package main

import (
	"flag"
	"fmt"
	"math/rand"
	"unicode"
)

var (
	maxAllowedRetries = flag.Int("max_allowed_retries", 10,
		"Max number of allowed retries.")
)

// Driver method to start the hangman game.
func StartHangman() {
	// Initialize the game.
	InitGame(nil)
	for {
		fmt.Println("Do you want to play a new game? (Y/N): ")
		inputChar := readChar()
		if unicode.ToLower(inputChar) == 'n' {
			break
		}
		if unicode.ToLower(inputChar) != 'y' {
			fmt.Println("Invalid input character, please enter a valid input (y/n)")
			continue
		}
		fmt.Println("Enter the expected length of the word: ")
		var expectedLen int
		_, err := fmt.Scan(&expectedLen)
		if err != nil {
			fmt.Println("Invalid input given, error: ", err)
			continue
		}
		// Get number of retries.
		fmt.Println("Enter the expected number of retries(max allowed " +
			"retries: ", *maxAllowedRetries, "):")
		var expectedRetries int
		_, err = fmt.Scan(&expectedRetries)
		if err != nil {
			fmt.Println("Invalid input given for number of retries, error ", err)
			continue
		}
		game, errCode := NewGame(expectedLen, expectedRetries)
		if errCode != NoError {
			if errCode == InvalidLength {
				fmt.Println("Sorry we do not have any words of length ",
					expectedLen, " in the dictionary. Please try again!")
			} else if errCode == InvalidRetries {
				fmt.Println("Invalid value of expected retries, please try again")
			} else {
				// Adding a generic case. This if else should be extended with
				// more error codes in future if needed.
				fmt.Println("Oops, input validation failed! Please try again.")
			}
			continue
		}
		// Start checking the user input character.
		for {
			fmt.Println(string(game.CurrentDisplayedWord))
			fmt.Println("Enter a character (previous characters: ",
				string(game.UsedChars), ", remaining tries", game.CurrentRetries, "): ")
			char := readChar()
			acceptedChar, err := game.CheckUserInput(char)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if acceptedChar {
				if game.State == Running {
					fmt.Println("You guessed a right character!!")
				} else if game.State == Won {
					fmt.Println("You won! Congratulations!!!")
					break
				} else {
					// Pick any random word and show it to the user.
					randomIndex := rand.Intn(len(game.CurrentSetOfWords))
					fmt.Println("All retries finished, you lose!! Chosen word was: ",
						game.CurrentSetOfWords[randomIndex])
					break
				}
			} else {
				if game.State == Running {
					fmt.Println("Sorry its a wrong input. Remaining tries: ", game.CurrentRetries)
				} else if game.State == Lost {
					// Pick any random word and show it to the user.
					randomIndex := rand.Intn(len(game.CurrentSetOfWords))
					fmt.Println("All retries finished, you lose!! Chosen word was: ",
						game.CurrentSetOfWords[randomIndex])
					break
				}
			}
		}
	}
}

func main() {
	flag.Parse()
	StartHangman()
}
