package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	TableWidth       = 115
	TableHeight      = 36
	BatLength        = 10
	ScoreToWon       = 10
	BallInitialSpeed = 1
	BatSpeed         = 2
	BallMaxSpeed     = 7
	GameEventsBuffer = 5
)

//Game is a main ping-pong struct, will all the information about state, and handful methods like 'Tick'.
//Player wins when gets 10 scores.
type Game struct {
	Table                   *Table
	LeftPlayer, RightPlayer *Player
	gameEvents              chan GameEvent
	finishGame              chan bool
}

func (game *Game) String() string {
	return fmt.Sprintf("Game{LeftPlayer: %s, RightPlayer: %s}", game.LeftPlayer, game.RightPlayer)
}

//GameEvent describes events can occur in games, such as reaction on player command to move Bat,
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

//Table describes Table state.
//Table has 2 dimensions, with the top in left upper corner.
type Table struct {
	Width, Height     int
	LeftBat, RightBat *Bat
	Ball              *Ball
}

type Bat struct {
	X, Y, Length, YSpeed int
}

type Ball struct {
	X, Y           int
	XSpeed, YSpeed int
}

type Player struct {
	Name  string
	Bat   *Bat
	Score int
}

func (player *Player) String() string {
	return fmt.Sprintf("Player{Name: %s, Score: %d}", player.Name, player.Score)
}

//NewGame constructs Game object and inner objects, filling them with suitable defaults.
func NewGame() *Game {
	rand.Seed(time.Now().UTC().UnixNano())

	leftBat := newBat(0)
	rightBat := newBat(TableWidth)

	table := newTable(leftBat, rightBat)

	gameEvents := make(chan GameEvent, GameEventsBuffer)
	finishGame := make(chan bool, 1)

	game := &Game{table,
		newPlayer("Left Player", leftBat),
		newPlayer("Right Player", rightBat),
		gameEvents,
		finishGame,
	}

	return game
}

//launchGameEventsHandler launches async handling of game events
func (game *Game) launchGameEventsHandler() {
	go handleGameEvents(game)
}

//handleGameEvents processes all the variety of game events, such as scoring, moving bat, etc.
func handleGameEvents(game *Game) {
	leftBat := game.Table.LeftBat
	rightBat := game.Table.RightBat
	ball := game.Table.Ball

	for event := range game.gameEvents {
		switch event {
		case LeftPlayerScores:
			newScore := game.LeftPlayer.Score + 1
			game.LeftPlayer.Score = newScore
			game.resetBallPosition()

			checkGameFinishes(game, newScore, LeftPlayerWon)
		case RightPlayerScores:
			newScore := game.RightPlayer.Score + 1
			game.RightPlayer.Score = newScore
			game.resetBallPosition()

			checkGameFinishes(game, newScore, RightPlayerWon)

		case LeftBatUp:
			leftBat.YSpeed = -BatSpeed
		case LeftBatDown:
			leftBat.YSpeed = BatSpeed
		case RightBatUp:
			rightBat.YSpeed = -BatSpeed
		case RightBatDown:
			rightBat.YSpeed = BatSpeed

		case BallStrickesBat:
			if randomBool() && randomBool() {
				ball.XSpeed = increaseUpToMax(ball.XSpeed, BallMaxSpeed)
				break
			}

			if randomBool() && randomBool() && randomBool() {
				ball.YSpeed = increaseUpToMax(ball.YSpeed, BallMaxSpeed)
				break
			}
		}
	}
}

func checkGameFinishes(game *Game, newScore int, event GameEvent) {
	if newScore >= ScoreToWon {
		game.LeftPlayer.Score = 0
		game.RightPlayer.Score = 0
		game.gameEvents <- event
	}
}

func (game *Game) resetBallPosition() {
	ball := game.Table.Ball

	ball.X = TableWidth / 2
	ball.Y = TableHeight / 2

	ball.XSpeed = BallInitialSpeed
	ball.YSpeed = BallInitialSpeed

	ball.XSpeed = -ball.XSpeed

	if randomBool() {
		ball.YSpeed = -ball.YSpeed
	}
}

//Tick is a core method for Game, it updates worlds physics.
func (game *Game) Tick() {
	game.updateBallCoor()

	table := game.Table
	game.updateBatCoor(table.LeftBat)
	game.updateBatCoor(table.RightBat)
}

func (game *Game) updateBallCoor() {
	table := game.Table

	ball := table.Ball

	height := table.Height
	width := table.Width

	game.updateBallX(ball, width)
	game.updateBallY(ball, height)
}

func (game *Game) updateBatCoor(bat *Bat) {
	bat.Y = bat.Y + bat.YSpeed
	bat.YSpeed = 0

	height := game.Table.Height
	if bat.Y+bat.Length >= height {
		bat.Y = height - bat.Length
	}

	if bat.Y <= 0 {
		bat.Y = 0
	}
}

//TODO rewrite collision logic
//TODO add collision tests for X and Y
func (game *Game) updateBallX(ball *Ball, width int) {
	leftBat := game.Table.LeftBat
	rightBat := game.Table.RightBat

	ball.X = ball.X + ball.XSpeed

	if ball.X < 0 {
		impactY := ball.Y + ball.YSpeed/2
		if isBallTouchesBat(leftBat, impactY) {
			ball.X = -ball.X
			ball.XSpeed = -ball.XSpeed

			game.gameEvents <- BallStrickesBat
		} else {
			game.gameEvents <- RightPlayerScores
		}
	}

	if ball.X > width {
		impactY := ball.Y + ball.YSpeed/2
		if isBallTouchesBat(rightBat, impactY) {
			ball.X = width - (ball.X - width)
			ball.XSpeed = -ball.XSpeed

			game.gameEvents <- BallStrickesBat
		} else {
			game.gameEvents <- LeftPlayerScores
		}
	}
}

func isBallTouchesBat(bat *Bat, impactY int) bool {
	return bat.Y <= impactY && (bat.Y+bat.Length) >= impactY
}

func (game *Game) updateBallY(ball *Ball, height int) {
	ball.Y = ball.Y + ball.YSpeed
	if ball.Y > height {
		ball.Y = height - (ball.Y - height)
		ball.YSpeed = -ball.YSpeed
	}
	if ball.Y < 0 {
		ball.Y = -ball.Y
		ball.YSpeed = -ball.YSpeed
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
	return &Ball{TableWidth / 2, TableHeight / 2, BallInitialSpeed, BallInitialSpeed}
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
