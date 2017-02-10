a practice project implement websocket protocol in server

install:`go get github.com/ruandao/websocket`

usage:
```
    port := ":8089"
	address := "/websocket"
	ser := websocket.New(port, address)
	ser.Listen()
	fmt.Printf("listen on %s\n", port)
	for {
		conn := ser.Accept()
		fmt.Printf("will run on\n")
		conn.On("hi", func() {
			fmt.Printf("你说 hi\n")
			conn.Emit(fmt.Sprintf("你说 hi %s", args))
		})
		conn.On("hello", func() {
			conn.Emit(fmt.Sprintf("你说 hello %s", args))
		})
		conn.On("close", func() {
			conn.Close()
		})
		conn.On("error", func() {

		})
	}
```