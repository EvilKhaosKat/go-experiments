package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
)

//terminal representation
const (
	EmptySymbol   = ' '
	BallSymbol    = '*'
	BatBodySymbol = '#'
	BorderSymbol  = '.'
	Foreground    = termbox.ColorWhite
	Background    = termbox.ColorBlack
)

//visualize is a main method for terminal ui, using current game state prints out all the information for player.
func visualize(game *Game) {
	table := game.Table

	clearTerminal(table.Width, table.Height)
	drawBorders(table.Width, table.Height)

	visualizeBall(table.Ball)
	visualizeBat(table.LeftBat)
	visualizeBat(table.RightBat)
	visualizeScore(game.LeftPlayer, game.RightPlayer)

	termbox.Flush()
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

func visualizeScore(leftPlayer *Player, rightPlayer *Player) {
	printLeftPlayerScore(leftPlayer.Score)
	printRightPlayerScore(rightPlayer.Score)
}

func drawBorders(width int, height int) {
	for x := 0; x <= width; x++ {
		termbox.SetCell(x, height, BorderSymbol, Foreground, Background)
	}

	for y := 0; y <= height; y++ {
		termbox.SetCell(width+1, y, BorderSymbol, Foreground, Background)
	}
}

func visualizeBall(ball *Ball) {
	termbox.SetCell(ball.X, ball.Y, BallSymbol, Foreground, Background)
}

func visualizeBat(bat *Bat) {
	batHeadCoor := bat.Y
	for y := bat.Y; y < batHeadCoor+bat.Length; y++ {
		termbox.SetCell(bat.X, y, BatBodySymbol, Foreground, Background)
	}
}

func printLeftPlayerScore(score int) {
	printPlayerScore(0, score)
}

func printRightPlayerScore(score int) {
	printPlayerScore(TableWidth, score)
}

func printPlayerScore(xCoor, score int) {
	termbox.SetCell(xCoor, TableHeight+1, scoreToRune(score), Foreground, Background)
}

func scoreToRune(score int) rune {
	return rune(score + '0')
}

//TODO it's significantly cheaper to erase only previous states/cells instead of full screen
func clearTerminal(width, height int) {
	for x := 0; x <= width+BallMaxSpeed; x++ {
		for y := 0; y <= height+2; y++ {
			termbox.SetCell(x, y, EmptySymbol, Foreground, Background)
		}
	}
}
