# Word guess with a twist

Instructions to build the code:
1. Install latest version of golang.
2. Install logging library using the following command 
go get "github.com/golang/glog"

Setup GOPATH etc appropriately.
Build the code using "go build"

Instructions to run the code:
1. You can download the executable named "hangman"
2. A default dictionary of words is included in the repo. But if a different dictionary is needed, please specify the path of the dictionary using the gflag "--dictionary=<>"
3. Max retries allowed are 10 retries by default. If you want to allow more retries for your hangman, you can change it by using the gflag "--max_allowed_retries=<>"
Command becomes: "./hangman --dictionary=<> --max_allowed_retries=<>"

Instructions to play the game:
1. Start a new game.
2. Chose the expected length of the word. The program returns an error if no word of that length exists in the dictionary.
3. Input the expected number of retries. Program allows a max retry of 10 by default.
4. Start giving a single character whenever prompted.

Assumptions:
1. Number of retries given is the number of incorrect guesses allowed.
2. The game is not case sensitive.
3. Dictionary words with special characters are discarded.

Cheating algorithm:
1. The program does not select a single word but keeps a list of words which can be the "secret word" that user is trying to guess.
2. With every user input, program tries to optimize how it can have the maximum group of words which can be potential candidates for the "secret word".
3. This is done by calculating the size of list of words if the user input was accepted (at various positions) or it was rejected.
4. With every iteration, the list of potential words keep reducing.
5. In case the program has multiple options with the same user input (the potential list of words is same when accepting or rejecting the user input), the program tries to select the option where min characters in the word are revealed.

Testing:
Ran the Unit test added in the repo.
Also did some manual testing with various scenarios.
