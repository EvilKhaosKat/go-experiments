package main

import (
	"testing"
)

func TestPlayersScores(t *testing.T) {
	//given
	game := NewGame()

	//when
	game.gameEvents <- LeftPlayerScores
	game.gameEvents <- RightPlayerScores
	game.gameEvents <- RightPlayerScores

	//events handling is async, has buffer of 1 event, so to be sure score events were handled - send 2 more events.
	//looks dirty though.
	game.gameEvents <- LeftBatDown
	game.gameEvents <- RightBatDown

	//then
	leftPlayerScore := game.leftPlayer.score
	if leftPlayerScore != 1 {
		t.Errorf("Left player must have score 1, but has %d", leftPlayerScore)
	}

	rightPlayerScore := game.rightPlayer.score
	if rightPlayerScore != 2 {
		t.Errorf("Right player must have score 2, but has %d", rightPlayerScore)
	}
}

func TestBatsMoving(t *testing.T) {
	//given
	game := NewGame()
	initialY := game.table.leftBat.yCoor

	//when
	game.gameEvents <- LeftBatDown //increase y
	game.gameEvents <- RightBatUp  //decrease y

	game.Tick()

	//events handling is async, has buffer of 1 event, so to be sure moving events were handled - send 2 more events.
	//looks dirty though.
	game.gameEvents <- RightPlayerScores
	game.gameEvents <- RightPlayerScores

	//then
	newLeftBatY := game.table.leftBat.yCoor
	if initialY-newLeftBatY != -BatMovingSpeed {
		t.Errorf("Left bat must be higher, but has coor %d, initial y coor %d", newLeftBatY, initialY)
	}

	newRightBatY := game.table.rightBat.yCoor
	if initialY-newRightBatY != BatMovingSpeed {
		t.Errorf("Right bat must be lower, but has coor %d, initial y coor %d", newLeftBatY, initialY)
	}
}
