#!/usr/bin/env bash
env GOOS=darwin GOARCH=amd64 go build -o ping_pong main.go Game.go Server.go Client.go Ui.go