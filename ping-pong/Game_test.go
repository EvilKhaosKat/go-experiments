package main

import (
	"testing"
)

func TestPlayersScores(t *testing.T) {
	//given
	game := NewGame()
	game.launchGameEventsHandler()

	//when
	game.gameEvents <- LeftPlayerScores
	game.gameEvents <- RightPlayerScores
	game.gameEvents <- RightPlayerScores

	//events handling is async, has buffer of 1 event, so to be sure Score events were handled - send 2 more events.
	//looks dirty though.
	game.gameEvents <- LeftBatDown
	game.gameEvents <- RightBatDown

	//then
	leftPlayerScore := game.LeftPlayer.Score
	if leftPlayerScore != 1 {
		t.Errorf("Left player must have Score 1, but has %d", leftPlayerScore)
	}

	rightPlayerScore := game.RightPlayer.Score
	if rightPlayerScore != 2 {
		t.Errorf("Right player must have Score 2, but has %d", rightPlayerScore)
	}
}

func TestBatsMoving(t *testing.T) {
	//given
	game := NewGame()
	game.launchGameEventsHandler()
	initialY := game.Table.LeftBat.Y

	//when
	game.gameEvents <- LeftBatDown //increase Y
	game.gameEvents <- RightBatUp  //decrease Y

	game.Tick()

	//events handling is async, has buffer of 1 event, so to be sure moving events were handled - send 2 more events.
	//looks dirty though.
	game.gameEvents <- RightPlayerScores
	game.gameEvents <- RightPlayerScores

	//then
	newLeftBatY := game.Table.LeftBat.Y
	if initialY-newLeftBatY != -BatSpeed {
		t.Errorf("Left Bat must be higher, but has coor %d, initial Y coor %d", newLeftBatY, initialY)
	}

	newRightBatY := game.Table.RightBat.Y
	if initialY-newRightBatY != BatSpeed {
		t.Errorf("Right Bat must be lower, but has coor %d, initial Y coor %d", newLeftBatY, initialY)
	}
}

func TestBallMoving(t *testing.T) {
	//given
	game := NewGame()
	game.launchGameEventsHandler()
	ball := game.Table.Ball

	xInitial := ball.X
	yInitial := ball.Y
	xSpeed := ball.XSpeed
	ySpeed := ball.YSpeed

	//when
	game.Tick()

	//then
	xNew := ball.X
	if xNew != xInitial+xSpeed {
		t.Errorf("X coordinate of Ball is wrong, supposed to be %d, but it is %d", xInitial+xSpeed, xNew)
	}

	yNew := ball.Y
	if yNew != yInitial+ySpeed {
		t.Errorf("Y coordinate of Ball is wrong, supposed to be %d, but it is %d", yInitial+ySpeed, yNew)
	}
}
