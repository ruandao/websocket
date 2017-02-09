//package main
//
//import (
//	"net/http"
//	"crypto/sha1"
//	"encoding/base64"
//	"fmt"
//)
//type callback func (...string)
//
//func failTheWebsocketConnection(msg string, w http.ResponseWriter, r *http.Request) {
//	// websocket rfc 7.1.7
//	fmt.Printf("xxx error %s\n", msg)
//}
//
//type WebSocketServer struct {
//	address string
//	path string
//
//	connChan chan *Conn
//}
//
//func (ws *WebSocketServer)Listen() {
//	ws.connChan = make(chan *Conn)
//	http.HandleFunc(ws.path, func(w http.ResponseWriter, r *http.Request) {
//		if r.Method != "GET" {
//			failTheWebsocketConnection(fmt.Sprintf("method is not get(%s)\n", r.Method), w, r)
//			return
//		}
//		connection, upgrade := r.Header.Get("Connection"), r.Header.Get("Upgrade")
//		websocketVersion := r.Header.Get("Sec-Websocket-Version")
//		if connection != "Upgrade" {
//			failTheWebsocketConnection(fmt.Sprintf("connection is not Upgrade(%s)\n", connection), w, r)
//			return
//		}
//		if upgrade != "websocket" {
//			failTheWebsocketConnection(fmt.Sprintf("upgrade is not websocket(%s)\n", upgrade), w, r)
//			return
//		}
//		if websocketVersion != "13" {
//			failTheWebsocketConnection(fmt.Sprintf("websocket-version is not 13(%s)\n", websocketVersion), w, r)
//			return
//		}
//		websocketKey := r.Header.Get("Sec-Websocket-Key")
//		//origin := r.Header.Get("Origin")
//		websocketAccept :=  websocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
//		//hash := sha1.New()
//		hash := sha1.Sum([]byte(websocketAccept))
//		websocketAccept = base64.StdEncoding.EncodeToString(hash[:])
//
//		w.Header().Add("Upgrade", "websocket")
//		w.Header().Add("Connection", "Upgrade")
//		w.Header().Add("Sec-Websocket-Accept", websocketAccept)
//		w.WriteHeader(101)
//		hj, ok := w.(http.Hijacker)
//		if !ok {
//			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
//			return
//		}
//		_conn, bufrw, err := hj.Hijack()
//		if err != nil {
//			http.Error(w, err.Error(), http.StatusInternalServerError)
//			return
//		}
//		conn := Conn{
//			Conn: _conn,
//			rw: bufrw,
//		}
//		ws.connChan <- &conn
//	})
//
//	go func() {
//		http.ListenAndServe(ws.address, nil)
//	}()
//}
//
//func (ws *WebSocketServer)Accept() *Conn {
//	conn := <- ws.connChan
//	go readLoop(conn)
//	go handleConn(conn)
//	return conn
//}
//
//func main() {
//
//	// 服务静态文件
//	go func() {
//		staticAddress := ":8080"
//		http.Handle("/", http.FileServer(http.Dir("./test-client")))
//		fmt.Printf("static file serve at %s\n", staticAddress)
//		http.ListenAndServe(staticAddress, nil)
//	}()
//
//	address := ":8089"
//	path := "/websocket"
//
//	ws := WebSocketServer{
//		path:path,
//		address:address,
//	}
//	ws.Listen()
//	fmt.Printf("websocket listen on %s\n", address)
//	for {
//		conn := ws.Accept()
//
//		conn.On("data", func(args ...string) {
//			fmt.Println("receive data")
//		})
//		conn.On("hi", func(args ...string) {
//			fmt.Printf("receive hi \n")
//		})
//		conn.On("close", func(args ...string) {
//			conn.Emit("close")
//			conn.Close()
//		})
//	}
//}
//
//func readLoop(conn *Conn) {
//	var err error
//	for err == nil {
//		var pd packetData
//		err = pd.decode(conn)
//		if err != nil {
//			return
//		}
//		conn.processReceivePacket(&pd)
//	}
//}
//
//func handleConn(conn *Conn) {
//	//defer conn.Close()
//	//defer fmt.Printf("%s will close\n", conn.Conn.RemoteAddr())
//	fmt.Printf("%s connection established\n", conn.Conn.RemoteAddr())
//
//	conn.Emit("你好世界")
//	conn.Emit("世界你好")
//	conn.Emit("Hello world!")
//}