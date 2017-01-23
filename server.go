package main

import (
	"net/http"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	//"time"
	"net"
	"bufio"
	"bytes"
	"io"
)
type callback func (...string)

func failTheWebsocketConnection(msg string, w http.ResponseWriter, r *http.Request) {
	// websocket rfc 7.1.7
	fmt.Printf("xxx error %s\n", msg)
}

type packetData struct {
	fin bool
	rsv1 bool
	rsv2 bool
	rsv3 bool
	opcode int
	masked	bool
	payload_len int
	payload_len2 int64
	mask_key string
	payload_data []byte
}

func (pd *packetData)encode(w io.Writer) (err error) {
	b := bufio.NewWriter(w)
	defer func() {
		if err == nil {
			err = b.Flush()
		}
	}()
	// 写入 fin, rsv1 rsv2, rsv3, opcode
	firstByte :=byte(pd.fin)	<< 7	&
				byte(pd.rsv1)	<< 6	&
				byte(pd.rsv2)	<< 5	&
				byte(pd.rsv3)	<< 4	&
				byte(pd.opcode)
	err = b.WriteByte(firstByte)
	if err != nil {
		return err
	}

	// 写入 mask, payload len
	var secondByte byte
	secondByte = byte(pd.masked)	<< 7	&
					byte(pd.payload_len)
	err = b.WriteByte(secondByte)
	if err != nil {
		return err
	}
	if pd.payload_len == 126 {
		bs := []byte(int(pd.payload_len2))[2:4]
		_, err = b.Write(bs)
		if err != nil {
			return err
		}
	} else if pd.payload_len == 127 {
		bs := []byte(pd.payload_len2)
		_, err = b.Write(bs)
		if err != nil {
			return err
		}
	}
	if pd.masked {
		_, err = b.Write([]byte(pd.mask_key))
		if err != nil {
			return err
		}
	}
	_, err = b.Write(pd.payload_data)
	return err 
}

func (pd *packetData)decode(r io.Reader) (err error) {
	b := bufio.NewReader(r)
	firstByte, err := b.ReadByte()
	if err != nil {
		return err
	}
	pd.fin = (( 1 << 7 ) & firstByte ) > 0
	pd.rsv1 = (( 1 << 6 ) & firstByte ) > 0
	pd.rsv2 = (( 1 << 5 ) & firstByte ) > 0
	pd.rsv3 = (( 1 << 4 ) & firstByte ) > 0
	pd.opcode = ( ((1 << 3) | (1 << 2) | (1 << 1) | 1) & firstByte)
	secondByte, err := b.ReadByte()
	if err != nil {
		return err
	}
	pd.masked = ( (1 << 7) & secondByte ) > 0
	pd.payload_len = ( (1 << 6 | 1 << 5 | 1 << 4 | 1 << 3 | 1 << 2 | 1 << 1 | 1) & secondByte)
	lenp := pd.payload_len

	if pd.payload_len == 126 {
		var p [2]byte
		_, err := io.ReadFull(b, &p)
		if err != nil {
			return err
		}
		pd.payload_len2 = int64(p)
		lenp = pd.payload_len2
	} else if pd.payload_len == 127 {
		var p [8]byte
		_, err := io.ReadFull(b, &p)
		if err != nil {
			return err
		}
		pd.payload_len2 = int64(p)
		lenp = pd.payload_len2
	}
	if pd.masked {
		var p [4]byte
		_, err := io.ReadFull(b, &p)
		if err != nil {
			return err
		}
		pd.mask_key = string(p)
	}
	var p [lenp]byte
	_, err = io.ReadFull(b, &p)
	pd.payload_data = p[:]
}

type Conn struct {
	net.Conn
	rw *bufio.ReadWriter
}

func (conn *Conn)On(event string, cb callback)  {
	
}
func (conn *Conn)Emit(event string, args ...string)  {

}

type WebSocketServer struct {
	address string
	path string
	
	connChan chan *Conn
}

func (ws *WebSocketServer)Listen() {
	ws.connChan = make(chan *Conn)
	http.HandleFunc(ws.path, func(w http.ResponseWriter, r *http.Request) {
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
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		_conn, bufrw, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		conn := Conn{
			Conn: _conn,
			rw: bufrw,
		}
		ws.connChan <- &conn
	})

	go func() {
		http.ListenAndServe(ws.address, nil)
	}()
}

func (ws *WebSocketServer)Accept() *Conn {
	return <- ws.connChan
}

func main() {

	// 服务静态文件
	go func() {
		staticAddress := ":8080"
		http.Handle("/", http.FileServer(http.Dir("./test-client")))
		fmt.Printf("static file serve at %s\n", staticAddress)
		http.ListenAndServe(staticAddress, nil)
	}()

	address := ":8089"
	path := "/websocket"
	
	ws := WebSocketServer{
		path:path,
		address:address,
	}
	ws.Listen()
	fmt.Printf("websocket listen on %s\n", address)
	for {
		conn := ws.Accept()
		go handleConn(conn)
	}
}

func handleConn(conn *Conn) {
	//defer conn.Close()
	//defer fmt.Printf("%s will close\n", conn.Conn.RemoteAddr())
	fmt.Printf("%s connection established\n", conn.Conn.RemoteAddr())

	conn.On("helloworld", func(content ...string) {

	})
	conn.On("xxx", func(content ...string) {

	})
	conn.On("disconnect", func(content ...string) {
		conn.Close()
	})
}