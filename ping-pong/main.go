package main

import (
	"flag"
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
)

const (
	Fps = 25
)

//networking
const (
	Port   = 4242
	Client = "client"
	Server = "server"
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

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	validateTerminalSize()

	mode, port, ip := readCommandLineFlags()

	game := NewGame()
	defer handlePanic(game.finishGame)

	go handleTerminalEvents(game.gameEvents, game.finishGame)

	if *mode == Server {
		game.launchGameEventsHandler()

		clientConn := waitForClient(port)
		go handleClientMessages(game, clientConn)
		launchGameServerLoop(game, clientConn)
	} else if *mode == Client {
		serverConn := connectToServer(game.finishGame, ip, port)
		go handleServerMessages(game, serverConn)
		launchGameClientLoop(game, serverConn)
	}
}

func handlePanic(finishGameChan chan bool) {
	if r := recover(); r != nil {
		termbox.Close()

		fmt.Println(r)

		finishGameChan <- true
		os.Exit(0)
	}
}

func readCommandLineFlags() (mode *string, port *int, ip *string) {
	mode = flag.String("mode", Server, "working mode, either server (by default) or client")

	port = flag.Int("port", Port, "a port for server to listen or for client to connect")
	ip = flag.String("ip", "127.0.0.1", "ip address for client to connect")

	flag.Parse()

	return mode, port, ip
}

//TODO handle terminal events in more readable way
func handleTerminalEvents(gameEvents chan GameEvent, finishGame chan bool) {
terminalEventsLoop:
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlQ:
				break terminalEventsLoop
			case termbox.KeyArrowUp:
				gameEvents <- RightBatUp
			case termbox.KeyArrowDown:
				gameEvents <- RightBatDown
			default:
				switch ev.Ch {
				case 'w', 'W':
					gameEvents <- LeftBatUp
				case 's', 'S':
					gameEvents <- LeftBatDown
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}

	finishGame <- true
}
