#!/usr/bin/env bash
env GOOS=linux GOARCH=amd64 go build -o ping_pong_linux main.go Game.go Server.go Client.go Ui.go