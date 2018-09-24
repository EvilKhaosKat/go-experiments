package main

import (
	"github.com/nsf/termbox-go"
	"time"
)

const (
	TickDelayMs   = 50 * time.Millisecond
	EmptySymbol   = ' '
	BallSymbol    = '*'
	BatBodySymbol = '|'
	BorderSymbol  = '█'
	Foreground    = termbox.ColorWhite
	Background    = termbox.ColorBlack
)

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	finishGame := make(chan bool)
	//TODO check whether terminal size is big enough
	//termbox.Size()

	go handleTerminalEvents(finishGame)
	launchGame(finishGame)
}

func launchGame(finishGame chan bool) {
	game := NewGame()
	visualize(game)

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

func visualize(game Game) {
	table := game.table

	clearTerminal(table.width, table.height)
	//drawBorders(table.width, table.height)

	visualizeBall(table.ball)
	visualizeBat(table.leftBat)
	visualizeBat(table.rightBat)

	termbox.Flush()
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

//TODO it's significantly cheaper to erase only previous states/cells instead of full screen
func clearTerminal(width, height int) {
	for x := 0; x <= width+1; x++ {
		for y := 0; y <= height+1; y++ {
			termbox.SetCell(x, y, EmptySymbol, Foreground, Background)
		}
	}
}

func handleTerminalEvents(finishGame chan bool) {
	//wait for esc or ctrl+q pressed, and then exit
terminalEventsLoop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc ||
				ev.Key == termbox.KeyCtrlQ {
				break terminalEventsLoop
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}

	finishGame <- true
}
