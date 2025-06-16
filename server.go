package main

import (
	"log"
	"net"
	"net/http"
)

var ServerPort = 0
var LastOpenedVideo string

// StartMediaServer starts an HTTP server to serve the last opened video.
func StartMediaServer() {
	// Listen on a random available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	ServerPort = listener.Addr().(*net.TCPAddr).Port
	http.HandleFunc("/lastVideo", func(w http.ResponseWriter, r *http.Request) {
		if LastOpenedVideo == "" {
			http.Error(w, "No video opened", http.StatusNotFound)
		} else {
			http.ServeFile(w, r, LastOpenedVideo)
		}
	})

	err = http.Serve(listener, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
