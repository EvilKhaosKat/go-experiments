package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
	"time"
)

const (
	TickDelayMs   = 50 * time.Millisecond
	EmptySymbol   = ' '
	BallSymbol    = '*'
	BatBodySymbol = '#'
	BorderSymbol  = '.'
	Foreground    = termbox.ColorWhite
	Background    = termbox.ColorBlack
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	validateTerminalSize()

	game := NewGame()

	finishGame := make(chan bool)
	go handleTerminalEvents(game, finishGame)
	launchGameLoop(game, finishGame)
}

func validateTerminalSize() {
	width, height := termbox.Size()
	reqWidth, reqHeight := getRequiredScreenSize()
	if width < reqWidth || height < reqHeight {
		termbox.Close()

		fmt.Printf("Screen size is not sufficient. %dx%d minimum is required, %dx%d actually.\n",
			reqWidth, reqHeight, width, height)

		os.Exit(1)
	}
}

func getRequiredScreenSize() (width, height int) {
	return TableWidth + 3, TableHeight + 3
}

func launchGameLoop(game *Game, finishGame chan bool) {
	ticker := time.NewTicker(TickDelayMs)

mainLoop:
	for {
		select {
		case <-finishGame:
			break mainLoop
		case <-ticker.C:
			game.Tick()
			visualize(game)
		}
	}
}

func visualize(game *Game) {
	table := game.table

	clearTerminal(table.width, table.height)
	drawBorders(table.width, table.height)

	visualizeBall(table.ball)
	visualizeBat(table.leftBat)
	visualizeBat(table.rightBat)
	visualizeScore(game.leftPlayer, game.rightPlayer)

	termbox.Flush()
}

func visualizeScore(leftPlayer *Player, rightPlayer *Player) {
	printLeftPlayerScore(leftPlayer.score)
	printRightPlayerScore(rightPlayer.score)
}

func drawBorders(width int, height int) {
	for x := 0; x <= width; x++ {
		termbox.SetCell(x, height+1, BorderSymbol, Foreground, Background)
	}

	for y := 0; y <= height; y++ {
		termbox.SetCell(width+1, y, BorderSymbol, Foreground, Background)
	}
}

func visualizeBall(ball *Ball) {
	termbox.SetCell(ball.x, ball.y, BallSymbol, Foreground, Background)
}

func visualizeBat(bat *Bat) {
	batHeadCoor := bat.yCoor
	for y := bat.yCoor; y < batHeadCoor+bat.length; y++ {
		termbox.SetCell(bat.xCoor, y, BatBodySymbol, Foreground, Background)
	}
}

func printLeftPlayerScore(score int) {
	printPlayerScore(0, score)
}

func printRightPlayerScore(score int) {
	printPlayerScore(TableWidth, score)
}

func printPlayerScore(xCoor, score int) {
	termbox.SetCell(xCoor, TableHeight+2, scoreToRune(score), Foreground, Background)
}

func scoreToRune(score int) rune {
	return rune(score + '0')
}

//TODO it's significantly cheaper to erase only previous states/cells instead of full screen
func clearTerminal(width, height int) {
	for x := 0; x <= width+1; x++ {
		for y := 0; y <= height+1; y++ {
			termbox.SetCell(x, y, EmptySymbol, Foreground, Background)
		}
	}
}

//TODO handle terminal events in more readable way
func handleTerminalEvents(game *Game, finishGame chan bool) {
terminalEventsLoop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlQ:
				break terminalEventsLoop
			case termbox.KeyArrowUp:
				game.gameEvents <- RightBatUp
			case termbox.KeyArrowDown:
				game.gameEvents <- RightBatDown
			default:
				switch ev.Ch {
				case 'w', 'W':
					game.gameEvents <- LeftBatUp
				case 's', 'S':
					game.gameEvents <- LeftBatDown
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}

	finishGame <- true
}
