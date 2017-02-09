package main

import (
	"github.com/ruandao/websocket"
	"fmt"
	"net/http"
)

func main() {

	// 服务静态文件
	go func() {
		staticAddress := ":8080"
		http.Handle("/", http.FileServer(http.Dir("./test-client")))
		fmt.Printf("static file serve at %s\n", staticAddress)
		http.ListenAndServe(staticAddress, nil)
	}()

	port := ":8089"
	address := "/websocket"
	ser := websocket.New(port, address)
	ser.Listen()
	fmt.Printf("listen on %s\n", port)
	for {
		conn := ser.Accept()
		conn.On("hi", func(args ...string) {
			fmt.Printf("你说 hi\n")
			conn.Emit(fmt.Sprintf("你说 hi %s", args))
		})
		conn.On("hello", func(args ...string) {
			conn.Emit(fmt.Sprintf("你说 hello %s", args))
		})
		conn.On("close", func(args ...string) {
			conn.Close()
		})
		conn.On("error", func(args ...string) {

		})
	}
}
