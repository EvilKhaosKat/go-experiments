package main

const (
	TableWidth  = 100
	TableHeight = 40
	BatLength   = 7
)

type Game struct {
	table                   Table
	leftPlayer, rightPlayer Player
}

//Table describes table state
type Table struct {
	width, height     int
	leftBat, rightBat Bat
	ball              Ball
}

type Bat struct {
	yCoor, length int
}

type Ball struct {
	x, y           int
	xSpeed, ySpeed int
}

type Player struct {
	name  string
	bat   Bat
	score int
}

func NewGame() Game {
	leftBat := newBat()
	rightBat := newBat()

	table := newTable(leftBat, rightBat)

	return Game{table,
		newPlayer("Left Player", leftBat),
		newPlayer("Right Player", rightBat),
	}
}

func newTable(leftBat, rightBat Bat) Table {
	return Table{TableWidth, TableHeight,
		leftBat,
		rightBat,
		newBall()}
}

func newBat() Bat {
	return Bat{TableHeight/2 - BatLength/2, BatLength}
}

func newPlayer(name string, bat Bat) Player {
	return Player{name, bat, 0}
}

func newBall() Ball {
	return Ball{TableWidth / 2, TableHeight / 2, 0, 0}
}
