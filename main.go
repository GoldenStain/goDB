package main

import (
	"github.com/GoldenStain/goDB/server"
)

func main() {
	db := server.ConnectDB()
	server.StartServer(db)
}
