package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"time"
)

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

func connectToServer(finishGame chan bool, ip *string, port *int) *bufio.ReadWriter {
	defer handlePanic(finishGame)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *ip, *port))
	if err != nil {
		panic(errors.Wrap(err, "Error occurred during connecting to server"))
	}

	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
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
			panic(errors.Wrapf(err, "Error during reading server state message %b", serverStateMessage))
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
