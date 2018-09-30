package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	TableWidth       = 115
	TableHeight      = 40
	BatLength        = 10
	ScoreToWon       = 10
	BallInitialSpeed = 1
	BallMaxSpeed     = 7
	BatSpeed         = 1
)

//Game is a main ping-pong struct, will all the information about state, and handful methods like 'Tick'.
//Player wins when gets 10 scores.
type Game struct {
	table                   *Table
	leftPlayer, rightPlayer *Player
	gameEvents              chan GameEvent
}

func (game *Game) String() string {
	return fmt.Sprintf("Game{leftPlayer: %s, rightPlayer: %s}", game.leftPlayer, game.rightPlayer)
}

//GameEvent describes events can occure in games, such as reaction on player command to move bat,
// or if player scores.
type GameEvent int

const (
	LeftPlayerScores = GameEvent(iota)
	LeftPlayerWon
	LeftBatUp
	LeftBatDown
	RightPlayerScores
	RightPlayerWon
	RightBatUp
	RightBatDown
	BallStrickesBat
)

//Table describes table state.
//Table has 2 dimensions, with the top in left upper corner.
type Table struct {
	width, height     int
	leftBat, rightBat *Bat
	ball              *Ball
}

type Bat struct {
	x, y, length, ySpeed int
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

func (player *Player) String() string {
	return fmt.Sprintf("Player{name: %s, score: %d}", player.name, player.score)
}

func NewGame() *Game {
	rand.Seed(time.Now().UTC().UnixNano())

	leftBat := newBat(0)
	rightBat := newBat(TableWidth)

	table := newTable(leftBat, rightBat)

	gameEvents := make(chan GameEvent, 1)

	game := &Game{table,
		newPlayer("Left Player", leftBat),
		newPlayer("Right Player", rightBat),
		gameEvents,
	}

	go handleGameEvents(game)

	return game
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
	return &Ball{TableWidth / 2, TableHeight / 2, BallInitialSpeed, BallInitialSpeed}
}

func handleGameEvents(game *Game) {
	leftBat := game.table.leftBat
	rightBat := game.table.rightBat
	ball := game.table.ball

	for event := range game.gameEvents {
		switch event {
		case LeftPlayerScores:
			newScore := game.leftPlayer.score + 1
			game.leftPlayer.score = newScore
			game.resetBallPosition()

			checkGameFinishes(game, newScore, LeftPlayerWon)
		case RightPlayerScores:
			newScore := game.rightPlayer.score + 1
			game.rightPlayer.score = newScore
			game.resetBallPosition()

			checkGameFinishes(game, newScore, RightPlayerWon)

		case LeftBatUp:
			leftBat.ySpeed = -BatSpeed
		case LeftBatDown:
			leftBat.ySpeed = BatSpeed
		case RightBatUp:
			rightBat.ySpeed = -BatSpeed
		case RightBatDown:
			rightBat.ySpeed = BatSpeed

		case BallStrickesBat:
			if randomBool() && randomBool() {
				ball.xSpeed = increaseUpToMax(ball.xSpeed, BallMaxSpeed)
				break
			}

			if randomBool() && randomBool() {
				ball.ySpeed = increaseUpToMax(ball.ySpeed, BallMaxSpeed)
				break
			}
		}
	}
}

func checkGameFinishes(game *Game, newScore int, event GameEvent) {
	if newScore >= ScoreToWon {
		game.leftPlayer.score = 0
		game.rightPlayer.score = 0
		game.gameEvents <- event
	}
}

func (game *Game) resetBallPosition() {
	ball := game.table.ball

	ball.x = TableWidth / 2
	ball.y = TableHeight / 2

	ball.xSpeed = BallInitialSpeed
	ball.ySpeed = BallInitialSpeed

	ball.xSpeed = -ball.xSpeed

	if randomBool() {
		ball.ySpeed = -ball.ySpeed
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

	ball := table.ball

	height := table.height
	width := table.width

	game.updateBallX(ball, width)
	game.updateBallY(ball, height)
}

func (game *Game) updateBatCoor(bat *Bat) {
	bat.y = bat.y + bat.ySpeed
	bat.ySpeed = 0

	height := game.table.height
	if bat.y+bat.length > height {
		bat.y = height - bat.length
	}

	if bat.y < 0 {
		bat.y = 0
	}
}

//TODO rewrite collision logic
//TODO add collision tests for x and y
func (game *Game) updateBallX(ball *Ball, width int) {
	leftBat := game.table.leftBat
	rightBat := game.table.rightBat

	ball.x = ball.x + ball.xSpeed

	if ball.x < 0 {
		impactY := ball.y + ball.ySpeed/2
		if isBallTouchesBat(leftBat, impactY) {
			ball.x = -ball.x
			ball.xSpeed = -ball.xSpeed

			game.gameEvents <- BallStrickesBat
		} else {
			game.gameEvents <- RightPlayerScores
		}
	}

	if ball.x > width {
		impactY := ball.y + ball.ySpeed/2
		if isBallTouchesBat(rightBat, impactY) {
			ball.x = width - (ball.x - width)
			ball.xSpeed = -ball.xSpeed

			game.gameEvents <- BallStrickesBat
		} else {
			game.gameEvents <- LeftPlayerScores
		}
	}
}

func isBallTouchesBat(bat *Bat, impactY int) bool {
	return bat.y <= impactY && (bat.y+bat.length) >= impactY
}

func (game *Game) updateBallY(ball *Ball, height int) {
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

func randomBool() bool {
	return rand.Intn(2) == 0
}

func increaseUpToMax(num, max int) int {
	resultNum := abs(num)

	if resultNum < max {
		resultNum += 1
	}

	if num < 0 {
		return -resultNum
	} else {
		return resultNum
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
