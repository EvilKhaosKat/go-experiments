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
	initialY := game.table.leftBat.y

	//when
	game.gameEvents <- LeftBatDown //increase y
	game.gameEvents <- RightBatUp  //decrease y

	game.Tick()

	//events handling is async, has buffer of 1 event, so to be sure moving events were handled - send 2 more events.
	//looks dirty though.
	game.gameEvents <- RightPlayerScores
	game.gameEvents <- RightPlayerScores

	//then
	newLeftBatY := game.table.leftBat.y
	if initialY-newLeftBatY != -BatSpeed {
		t.Errorf("Left bat must be higher, but has coor %d, initial y coor %d", newLeftBatY, initialY)
	}

	newRightBatY := game.table.rightBat.y
	if initialY-newRightBatY != BatSpeed {
		t.Errorf("Right bat must be lower, but has coor %d, initial y coor %d", newLeftBatY, initialY)
	}
}

func TestBallMoving(t *testing.T) {
	//given
	game := NewGame()
	ball := game.table.ball

	xInitial := ball.x
	yInitial := ball.y
	xSpeed := ball.xSpeed
	ySpeed := ball.ySpeed

	//when
	game.Tick()

	//then
	xNew := ball.x
	if xNew != xInitial+xSpeed {
		t.Errorf("X coordinate of ball is wrong, supposed to be %d, but it is %d", xInitial+xSpeed, xNew)
	}

	yNew := ball.y
	if yNew != yInitial+ySpeed {
		t.Errorf("Y coordinate of ball is wrong, supposed to be %d, but it is %d", yInitial+ySpeed, yNew)
	}
}
