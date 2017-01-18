package main

import (
	"net/http"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

func failTheWebsocketConnection(msg string, w http.ResponseWriter, r *http.Request) {
	// websocket rfc 7.1.7
	fmt.Printf("xxx error %s\n", msg)
}

func main() {
	address := ":8080"
	path := "/socket.io/"

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			failTheWebsocketConnection(fmt.Sprintf("method is not get(%s)\n", r.Method), w, r)
			return
		}
		connection, upgrade := r.Header.Get("Connection"), r.Header.Get("Upgrade")
		websocketVersion := r.Header.Get("Sec-Websocket-Version")
		if connection != "Upgrade" {
			failTheWebsocketConnection(fmt.Sprintf("connection is not Upgrade(%s)\n", connection), w, r)
			return
		}
		if upgrade != "websocket" {
			failTheWebsocketConnection(fmt.Sprintf("upgrade is not websocket(%s)\n", upgrade), w, r)
			return
		}
		if websocketVersion != "13" {
			failTheWebsocketConnection(fmt.Sprintf("websocket-version is not 13(%s)\n", websocketVersion), w, r)
			return
		}
		websocketKey := r.Header.Get("Sec-Websocket-Key")
		//origin := r.Header.Get("Origin")
		hash := sha1.New()
		websocketAccept :=  websocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
		websocketAccept = base64.StdEncoding.EncodeToString(hash.Sum([]byte(websocketAccept)))
		w.WriteHeader(101)
		w.Header().Add("Upgrade", "websocket")
		w.Header().Add("Connection", "Upgrade")
		w.Header().Add("Sec-Websocket-Accept", websocketAccept)

	})

	http.Handle("/", http.FileServer(http.Dir("./test-client")))

	fmt.Printf("Listen at address %s\n", address)
	http.ListenAndServe(address, nil)
}
