package main

import (
	"github.com/nsf/termbox-go"
	"time"
)

const (
	TickDelayMs   = 100 * time.Millisecond
	EmptySymbol   = ' '
	BallSymbol    = '*'
	BatBodySymbol = '|'
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

	clearTerminal(table)

	ball := table.ball
	termbox.SetCell(ball.x, ball.y, BallSymbol, Foreground, Background)

	termbox.Flush()
}

//TODO it's significantly cheaper to erase only previous states/cells instead of full screen
func clearTerminal(table *Table) {
	for x := 0; x <= table.width; x++ {
		for y := 0; y <= table.height; y++ {
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
