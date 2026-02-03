package main

import (
	"fmt"
	"log"
	"net/http"

	gameinit "server/game-init"
	state "server/state"
)

func main() {
	fmt.Println("Welcome to Sporcle!")

	globalState := state.NewGlobalState()
	mux := http.NewServeMux()
	gameinit.RegisterRoutes(mux, globalState)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
