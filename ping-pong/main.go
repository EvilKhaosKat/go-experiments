package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/pkg/errors"
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

func launchGameClientLoop(game *Game, serverConn *bufio.ReadWriter) {
	ticker := time.NewTicker(time.Second / Fps)

mainLoop:
	for {
		select {
		case <-game.finishGame:
			break mainLoop
		case <-ticker.C:
			go sendStateToServer(game, serverConn)
			visualize(game)
		}
	}
}

func sendStateToServer(game *Game, serverConn *bufio.ReadWriter) {
	defer handlePanic(game.finishGame)

	for gameEvent := range game.gameEvents {
		gameEventData := []byte{
			byte(gameEvent),
			'\n',
		}

		if _, err := serverConn.Write(gameEventData); err != nil {
			panic(errors.Wrapf(err, "Error during sending client event %s to server", gameEventData))
		}

		err := serverConn.Flush()
		if err != nil {
			panic(err)
		}
	}
}

func handleServerMessages(game *Game, serverConn *bufio.ReadWriter) {
	defer handlePanic(game.finishGame)

	for {
		serverStateMessage, _, err := serverConn.ReadLine()
		if err != nil {
			panic(errors.Wrapf(err, "Error during reading server state message: %b", serverStateMessage))
		}

		var serverGameState Game

		if err := json.Unmarshal([]byte(serverStateMessage), &serverGameState); err != nil {
			panic(errors.Wrapf(err,
				"Error during parsing server state %s", serverStateMessage))

		}

		game.Table = serverGameState.Table
		game.LeftPlayer = serverGameState.LeftPlayer
		game.RightPlayer = serverGameState.RightPlayer
	}
}

func connectToServer(finishGame chan bool, ip *string, port *int) *bufio.ReadWriter {
	defer handlePanic(finishGame)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *ip, *port))
	if err != nil {
		panic(errors.Wrap(err, "Error occurred during connecting to server"))
	}

	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
}

func handleClientMessages(game *Game, clientConn *bufio.ReadWriter) {
	defer handlePanic(game.finishGame)

	for {
		clientMessage, err := bufio.NewReader(clientConn).ReadByte()
		if err != nil {
			panic(errors.Wrapf(err, "Error during reading client clientMessage: %b", clientMessage))
		}

		eventFromClient := GameEvent(clientMessage)
		if eventFromClient == RightBatUp || eventFromClient == RightBatDown {
			game.gameEvents <- eventFromClient
		} else {
			panic(errors.Wrapf(err, "Error during reading client clientMessage: %b", eventFromClient))
		}
	}
}

func waitForClient(port *int) *bufio.ReadWriter {
	fmt.Printf("Waiting for client on port %d\n", *port)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(errors.Wrap(err, "Error occurred during creating server"))
	}

	conn, err := ln.Accept()
	if err != nil {
		panic(errors.Wrapf(err, "Error occurred during accepting client connection"))
	}

	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
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

func launchGameServerLoop(game *Game, clientConn *bufio.ReadWriter) {
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

func sendStateToClient(game *Game, clientConn *bufio.ReadWriter) {
	state, err := json.Marshal(game)
	if err != nil {
		panic(errors.Wrap(err, "Error occured during creating server state"))
	}

	_, err = clientConn.Write(state)
	if err != nil {
		panic(errors.Wrap(err, "Eror during writing state message"))
	}

	_, err = clientConn.Write([]byte{'\n'})
	if err != nil {
		panic(errors.Wrap(err, "Error during writing line-ending for state message"))
	}

	err = clientConn.Flush()
	if err != nil {
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
