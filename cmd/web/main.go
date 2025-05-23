package main

import (
	"fmt"
	"os"

	"github.com/lafeingcrokodil/flashcards/server"
)

func main() {
	err := server.Start(":8080")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
