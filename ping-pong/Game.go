package main

const (
	TableWidth  = 100
	TableHeight = 40
	BatLength   = 7
)

type Game struct {
	table                   *Table
	leftPlayer, rightPlayer *Player
}

//Table describes table state
type Table struct {
	width, height     int
	leftBat, rightBat *Bat
	ball              *Ball
}

type Bat struct {
	xCoor, yCoor, length, ySpeed int
}

type Ball struct {
	x, y           int
	xSpeed, ySpeed int
}

type Player struct {
	name  string
	bat   *Bat
	score int
}

func NewGame() Game {
	leftBat := newBat(0)
	rightBat := newBat(TableWidth)

	table := newTable(leftBat, rightBat)

	return Game{table,
		newPlayer("Left Player", leftBat),
		newPlayer("Right Player", rightBat),
	}
}

func (game *Game) Tick() {
	game.updateBallCoor()

	table := game.table
	game.updateBatCoor(table.leftBat)
	game.updateBatCoor(table.rightBat)
}

func (game *Game) updateBallCoor() {
	table := game.table

	leftBat := table.leftBat
	rightBat := table.rightBat

	ball := table.ball

	height := table.height
	width := table.width

	updateBallX(ball, leftBat, rightBat, width)
	updateBallY(ball, height)
}

func (game *Game) updateBatCoor(bat *Bat) {
	bat.yCoor = bat.yCoor + bat.ySpeed

	height := game.table.height
	if bat.yCoor+bat.length > height {
		bat.yCoor = height - bat.length
		bat.ySpeed = 0
	}

	if bat.yCoor < 0 {
		bat.yCoor = 0
		bat.ySpeed = 0
	}
}

func updateBallX(ball *Ball, leftBat, rightBat *Bat, width int) {
	ball.x = ball.x + ball.xSpeed

	if ball.x < 0 {
		impactY := ball.y + ball.ySpeed/2
		if isBallTouchesBat(leftBat, impactY) {
			ball.x = -ball.x
			ball.xSpeed = -ball.xSpeed
		} else {
			//TODO add
		}
	}

	if ball.x > width {
		impactY := ball.y + ball.ySpeed/2
		if isBallTouchesBat(rightBat, impactY) {
			ball.x = width - (ball.x - width)
			ball.xSpeed = -ball.xSpeed
		} else {
			//TODO add
		}
	}
}

func isBallTouchesBat(bat *Bat, impactY int) bool {
	return bat.yCoor <= impactY && (bat.yCoor+bat.length) >= impactY
}

func updateBallY(ball *Ball, height int) {
	ball.y = ball.y + ball.ySpeed
	if ball.y > height {
		ball.y = height - (ball.y - height)
		ball.ySpeed = -ball.ySpeed
	}
	if ball.y < 0 {
		ball.y = -ball.y
		ball.ySpeed = -ball.ySpeed
	}
}

func newTable(leftBat, rightBat *Bat) *Table {
	return &Table{TableWidth, TableHeight,
		leftBat,
		rightBat,
		newBall()}
}

func newBat(xCoor int) *Bat {
	return &Bat{xCoor, TableHeight/2 - BatLength/2, BatLength, 0}
}

func newPlayer(name string, bat *Bat) *Player {
	return &Player{name, bat, 0}
}

func newBall() *Ball {
	return &Ball{TableWidth / 2, TableHeight / 2, 1, 1}
}
