//package main
//
//import (
//	"fmt"
//	"net"
//	"bufio"
//	"sync"
//)
//
//type Conn struct {
//	net.Conn
//	rw *bufio.ReadWriter
//	cb map[string]callback
//	cbLocker sync.RWMutex
//}
//
//func (conn *Conn)On(event string, cb callback)  {
//	conn.cbLocker.Lock()
//	defer conn.cbLocker.Unlock()
//	if conn.cb == nil {
//		conn.cb = make(map[string]callback)
//	}
//	conn.cb[event] = cb
//}
//func (conn *Conn)Emit(event string, args ...string) error {
//	var pd packetData
//	pd.fin = 1
//	pd.opcode = 1
//	lenp := len([]byte(event))
//	if lenp <= 125 {
//		pd.payload_len = lenp
//	} else if lenp >= 126 && lenp < 65536 {
//		pd.payload_len = 126
//		pd.payload_len2 = int64(lenp)
//	} else {
//		pd.payload_len = 127
//		pd.payload_len2 = int64(lenp)
//	}
//	pd.payload_data = []byte(event)
//
//	err := pd.encode(conn)
//	return err
//	//return nil
//}
//
//func (conn *Conn)processReceivePacket(pd *packetData) {
//	content := string(pd.payload_data)
//	for event,cb := range conn.cb {
//		fmt.Printf("event :%s content: %s match: %s\n", event, content, event == content)
//		if event == content {
//			cb()
//		}
//	}
//	fmt.Printf("read packet: \n%s\n", pd)
//}
//
