package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"time"
)

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

func sendStateToClient(game *Game, clientConn *bufio.ReadWriter) {
	state, err := json.Marshal(game)
	if err != nil {
		panic(errors.Wrap(err, "Error occurred during creating server state"))
	}

	_, err = clientConn.Write(state)
	if err != nil {
		panic(errors.Wrap(err, "Error during writing state message"))
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

func handleClientMessages(game *Game, clientConn *bufio.ReadWriter) {
	defer handlePanic(game.finishGame)

	for {
		clientMessage, err := bufio.NewReader(clientConn).ReadByte()
		if err != nil {
			panic(errors.Wrapf(err, "Error during reading clientMessage %b", clientMessage))
		}

		eventFromClient := GameEvent(clientMessage)
		if eventFromClient == RightBatUp || eventFromClient == RightBatDown {
			game.gameEvents <- eventFromClient
		} else {
			panic(errors.Wrapf(err, "Error during checking eventFromClient %b", eventFromClient))
		}
	}
}
