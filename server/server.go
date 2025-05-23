package server

import (
	"fmt"
	"net/http"
)

// Start starts the server.
func Start(addr string) error {
	http.HandleFunc("/echo", echo)
	return http.ListenAndServe(addr, nil)
}

func echo(w http.ResponseWriter, req *http.Request) {
	_, err := fmt.Fprintf(w, "Params: %#v\n", req.URL.Query())
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}
