package websocket

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
	"fmt"
)

type Conn struct {
	net.Conn
	cbMapping 				map[string]callback
	cbMappingLocker 		sync.RWMutex

	writeChan 				chan *PacketData
	closeStatus 			int32
	hadSendCloseFrame 		bool
	hadReceiveCloseFrame	bool
	pingLoopChan			chan int
	pingInterval			int
	writeChanCloseNotify	chan int
}

type callback func(...string)

func newConn(conn net.Conn) *Conn {
	_conn := &Conn{
		Conn: conn,
		cbMapping:make(map[string]callback),

		writeChan:make(chan *PacketData),
		pingLoopChan:make(chan int),
		pingInterval:25,
		writeChanCloseNotify: make(chan int),
	}
	go _conn.readLoop()
	go _conn.writeLoop()
	go _conn.pingLoop()
	return _conn
}

func (conn *Conn)On(event string, cb callback)  {
	conn.cbMappingLocker.Lock()
	defer conn.cbMappingLocker.Unlock()
	fmt.Printf("register on %s callback: %v\n", event, cb)
	conn.cbMapping[event] = cb
}

func (conn *Conn)Emit(event string, args ...string) {
	pds := newTextFrame(event, args...)
	for _, pd := range pds {
		conn.writeChan <- pd
	}
}

func (conn *Conn)Close() {
	pd := newCloseFrame()
	conn.writeChan <- pd
	conn.hadSendCloseFrame = true
	atomic.AddInt32(&conn.closeStatus, 1)
	conn.closeConn()
}
func (conn *Conn)closeConn()  {
	if conn.closeStatus == 2 {
		close(conn.writeChanCloseNotify)
		// 延迟300毫秒，使得不会和PingLoop发生竞争
		//（即，保证writeChanCloseNotify成功通知）
		<- time.After(300 * 1000 * 1000)
		close(conn.writeChan)
		conn.Conn.Close()
	}
}

func (conn *Conn)readLoop() {
	var pds []*PacketData
	for {
		var pd PacketData
		err := pd.decode(conn)
		if err != nil {
			fmt.Printf("err: %s\n", err)
			return
		}

		switch pd.frameType {
		case ContinuationFrame:
			pds = append(pds, &pd)
		case TextFrame, BinaryFrame:
			//fmt.Printf("frame: %s\n", pd.String())
			pds = append(pds, &pd)
			conn.processDataFrame(pds)
			pds = []*PacketData{}
		case CloseFrame:
			fmt.Printf("close frame\n")
			conn.hadReceiveCloseFrame = true
			atomic.AddInt32(&conn.closeStatus, 1)
			conn.closeConn()
			conn.processCloseFrame()
			//  结束read
			return
		case PingFrame, PongFrame:
			fmt.Printf("frameType: %s\n", pd.opcode)
			conn.pingLoopChan <- pd.frameType
		default:
			fmt.Printf("unknown frameType")
			return
		}
	}
}
func (conn *Conn)writeLoop() {
	for {
		pd, ok := <- conn.writeChan
		if !ok {
			return
		}
		pd.encode(conn)
	}
}
func (conn *Conn)pingLoop() {
	//var lastReceivePing int
	var lastReceivePong int
	for {
		select {
		case <- time.After(time.Duration(conn.pingInterval) * time.Second):
			fmt.Printf("write ping\n")
			pd := newPingFrame()
			conn.writeChan <- pd
		case frameType := <- conn.pingLoopChan:
			switch frameType {
			case PingFrame:
				//lastReceivePing = time.Now()
				fmt.Printf("receive ping Frame\n")
				pd := newPongFrame()
				conn.writeChan <- pd
			case PongFrame:
				fmt.Printf("receive pong Frame\n")
				lastReceivePong = time.Now().Nanosecond()
			}
		case <-time.After(time.Duration(time.Duration(conn.pingInterval * 3) * time.Second - time.Duration(lastReceivePong))):
			// 如果三个ping 间隔，仍然没有收到pong响应，则认为已经掉线了
			// 报错
			fmt.Printf("out of ping lifecycle\n")
			conn.processError()
			return
		case <-conn.writeChanCloseNotify:
			return
		}

	}
}

func (conn *Conn)processDataFrame(pds []*PacketData)  {
	_event := string(pds[0].payload_data)
	var args []string
	for _, pd := range pds[1:] {
		args = append(args, string(pd.payload_data))
	}
	conn.cbMappingLocker.RLock()
	defer conn.cbMappingLocker.RUnlock()

	for event, cb := range conn.cbMapping {
		fmt.Printf("client event: %s server event: %s equal: %v\n", _event, event, _event == event)
		if event == _event {
			cb(args...)
		}
	}
}

func (conn *Conn)processCloseFrame()  {
	callback, exist := conn.cbMapping["close"]
	fmt.Printf("close callback exist: %v\n", exist)
	if !exist {
		return
	}

	callback()
}
func (conn *Conn)processError() {
	callback, exist := conn.cbMapping["error"]
	if !exist {
		return
	}
	callback()
}