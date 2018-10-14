package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nsf/termbox-go"
	"log"
	"net"
	"os"
	"time"
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
	defer func() {
		if r := recover(); r != nil {
			termbox.Close()
		}
	}()
	validateTerminalSize()

	mode, port, ip := readCommandLineFlags()

	game := NewGame()
	go handleTerminalEvents(game, game.finishGame)

	if *mode == Server {
		clientConn := waitForClient(port)
		go handleClientMessages(game, clientConn)
		launchGameServerLoop(game, clientConn)
	} else if *mode == Client {
		fmt.Println(ip)
		log.Println("Not implemented")
		panic(err)
	}
}

func handleClientMessages(game *Game, clientConn net.Conn) {
	for {
		clientMessage, err := bufio.NewReader(clientConn).ReadByte()
		if err != nil {
			log.Printf("Error during reading client clientMessage: %b", clientMessage)
			panic(err)
		}

		eventFromClient := GameEvent(clientMessage)
		if eventFromClient == RightBatUp || eventFromClient == RightBatDown {
			game.gameEvents <- eventFromClient
		} else {
			log.Printf("Right client send incorrect event: %b", eventFromClient)
			panic(err)
		}
	}
}

func waitForClient(port *int) net.Conn {
	fmt.Printf("Waiting for client on port %d\n", *port)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Printf("Error occured during creating server: %s", err)
		panic(err)
	}

	conn, err := ln.Accept()
	if err != nil {
		log.Printf("Error occured during accepting client connection: %s", err)
		panic(err)
	}

	return conn
}

func readCommandLineFlags() (mode *string, port *int, ip *string) {
	mode = flag.String("mode", Server, "working mode, either server (by default) or client")

	port = flag.Int("port", Port, "a port for server to listen or for client to connect")
	ip = flag.String("ip", "127.0.0.1", "ip address for client to connect")

	flag.Parse()

	return mode, port, ip
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

func launchGameServerLoop(game *Game, clientConn net.Conn) {
	ticker := time.NewTicker(time.Second / Fps)

mainLoop:
	for {
		select {
		case <-game.finishGame:
			break mainLoop
		case <-ticker.C:
			game.Tick()
			sendStateToClient(game, clientConn)
			visualize(game)
		}
	}
}

func sendStateToClient(game *Game, clientConn net.Conn) {
	state, err := json.Marshal(game)
	if err != nil {
		log.Printf("Error occured during creating server state: %s", err)
		panic(err)
	}

	_, err = clientConn.Write(state)
	if err != nil {
		log.Printf("Eror during writing state message: %s", err)
		panic(err)
	}

	_, err = clientConn.Write([]byte("\n"))
	if err != nil {
		log.Printf("Eror during writing line-ending for state message: %s", err)
		panic(err)
	}
}

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
