package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"fmt"
)



func failTheWebsocketConnection(msg string, w http.ResponseWriter, r *http.Request) {
	// websocket rfc 7.1.7
	fmt.Printf("xxx error %s\n", msg)
}

type Server struct {
	address string
	path    string
	connChan chan *Conn
}

func New(address, path string) *Server {
	return &Server{
		address:address,
		path:path,
	}
}
func (ser *Server)Listen() {
	http.HandleFunc(ser.path, func(w http.ResponseWriter, r *http.Request) {
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
		websocketAccept :=  websocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
		//hash := sha1.New()
		hash := sha1.Sum([]byte(websocketAccept))
		websocketAccept = base64.StdEncoding.EncodeToString(hash[:])

		w.Header().Add("Upgrade", "websocket")
		w.Header().Add("Connection", "Upgrade")
		w.Header().Add("Sec-Websocket-Accept", websocketAccept)
		w.WriteHeader(101)
		// 怎么确定 client 已经收到上面的写入
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		_conn, _, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		conn := newConn(_conn)
		ser.connChan <- conn
	})

	go func() {
		http.ListenAndServe(ser.address, nil)
	}()
}

func (ser *Server)Accept() *Conn {
	conn := <- ser.connChan
	return conn
}
