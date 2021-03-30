package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context.
type HangmanTestSuite struct {
	suite.Suite
}

func (s *HangmanTestSuite) SetupSuite() {
	InitGame([]string{"last", "fast", "bets", "code"})
}

func (s *HangmanTestSuite) TestConflictingOptions() {
	game, errCode := NewGame(4, 5)
	assert.Equal(s.T(), NoError, errCode)
	isValid, err := game.CheckUserInput('a')
	// With the given dictionary, if 'a' is accepted, the groups will be of same
	// size. Based on our logic, 'a' should not be accepted.
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 4, game.CurrentRetries)
	assert.Equal(s.T(), []string{"bets", "code"}, game.CurrentSetOfWords)
}


func (s *HangmanTestSuite) TestWinningScenario() {
	game, errCode := NewGame(4, 2)
	assert.Equal(s.T(), NoError, errCode)
	isValid, err := game.CheckUserInput('a')
	// User input not accepted.
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 1, game.CurrentRetries)

	isValid, err = game.CheckUserInput('b')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 0, game.CurrentRetries)

	isValid, err = game.CheckUserInput('c')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), true, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 0, game.CurrentRetries)

	isValid, err = game.CheckUserInput('o')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), true, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 0, game.CurrentRetries)

	isValid, err = game.CheckUserInput('d')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), true, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 0, game.CurrentRetries)

	isValid, err = game.CheckUserInput('e')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), true, isValid)
	assert.Equal(s.T(), Won, game.State)
	assert.Equal(s.T(), 0, game.CurrentRetries)
}

func (s *HangmanTestSuite) TestDuplicateInputs() {
	game, errCode := NewGame(4, 8)
	assert.Equal(s.T(), NoError, errCode)
	isValid, err := game.CheckUserInput('i')
	// User input not accepted.
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 7, game.CurrentRetries)

	// Duplicate input key is rejected.
	isValid, err = game.CheckUserInput('i')
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 7, game.CurrentRetries)
}

func (s *HangmanTestSuite) TestLosingScenario() {
	game, errCode := NewGame(4, 3)
	assert.Equal(s.T(), NoError, errCode)
	isValid, err := game.CheckUserInput('i')
	// User input not accepted.
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 2, game.CurrentRetries)

	isValid, err = game.CheckUserInput('a')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 1, game.CurrentRetries)

	isValid, err = game.CheckUserInput('e')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), true, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 1, game.CurrentRetries)
	// Program could pick up any word out of "bets" and "code". But based on our
	// logic we expect to pick up lexicographically smaller string.
	assert.Equal(s.T(), []string{"code"}, game.CurrentSetOfWords)

	isValid, err = game.CheckUserInput('u')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, isValid)
	assert.Equal(s.T(), Running, game.State)
	assert.Equal(s.T(), 0, game.CurrentRetries)

	isValid, err = game.CheckUserInput('t')
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), false, isValid)
	assert.Equal(s.T(), Lost, game.State)
	assert.Equal(s.T(), -1, game.CurrentRetries)
}

func (s *HangmanTestSuite) TestInvalidInputs() {
	// Test case 1: Invalid length.
	game, errCode := NewGame(5, 3)
	assert.Nil(s.T(), game)
	assert.Equal(s.T(), InvalidLength, errCode)

	// Test case 2: Very large number of retries.
	game, errCode = NewGame(4, 15)
	assert.Nil(s.T(), game)
	assert.Equal(s.T(), InvalidRetries, errCode)

	// Test case 3: Negative number of retries.
	game, errCode = NewGame(4, -1)
	assert.Nil(s.T(), game)
	assert.Equal(s.T(), InvalidRetries, errCode)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestHangmanTestSuite(t *testing.T) {
	suite.Run(t, new(HangmanTestSuite))
}
