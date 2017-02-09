package websocket

import (
	"io"
	"strconv"
	"bufio"
	"bytes"
	"encoding/binary"
	"strings"
	"fmt"
)


type opCode int

func (oc opCode)String() string {
	switch oc {
	case 0:
		return "ContinuationFrame"
	case 1:
		return "TextFrame"
	case 2:
		return "BinaryFrame"
	case 3,4,5,6,7:
		return "FurtherFrame(non-control)"
	case 8:
		return "ConnectionClose"
	case 9:
		return "Ping"
	case 10:
		return "Pong"
	default:
		return "FurtherFrame(control)"
	}
}

type PacketData struct {
	frameType int
	payload []byte
	fin int
	rsv1 int
	rsv2 int
	rsv3 int
	opcode opCode
	masked	int
	payload_len int
	payload_len2 int64
	mask_key [4]byte
	payload_data []byte
}

const (
	ContinuationFrame = 0
	TextFrame         = 1
	BinaryFrame       = 2
	CloseFrame        = 8
	PingFrame         = 9
	PongFrame         = 10
)

func newTextFrame(event string, args ...string) []*PacketData {
	data := []string{}
	data = append(data, event)
	data = append(data, args...)
	var pds []*PacketData
	for idx, arg := range data {
		var pd PacketData
		pd.fin = 0
		pd.opcode = TextFrame
		if idx == len(data) - 1 {
			pd.fin = 1
			if len(data) > 1 {
				pd.opcode = 0
			}
		}

		lenp := len([]byte(arg))
		if lenp <= 125 {
			pd.payload_len = lenp
		} else if lenp >= 126 && lenp < 65536 {
			pd.payload_len = 126
			pd.payload_len2 = int64(lenp)
		} else {
			pd.payload_len = 127
			pd.payload_len2 = int64(lenp)
		}
		pd.payload_data = []byte(arg)

		pds = append(pds, &pd)
	}
	return pds
}
func newCloseFrame() *PacketData {
	var pd PacketData
	pd.fin = 1
	pd.opcode = CloseFrame
	lenp := 0
	if lenp <= 125 {
		pd.payload_len = lenp
	} else if lenp >= 126 && lenp < 65536 {
		pd.payload_len = 126
		pd.payload_len2 = int64(lenp)
	} else {
		pd.payload_len = 127
		pd.payload_len2 = int64(lenp)
	}
	//pd.payload_data = []byte(event)
	return &pd
}
func newPingFrame() *PacketData {
	var pd PacketData
	pd.fin = 1
	pd.opcode = PingFrame
	lenp := 0
	if lenp <= 125 {
		pd.payload_len = lenp
	} else if lenp >= 126 && lenp < 65536 {
		pd.payload_len = 126
		pd.payload_len2 = int64(lenp)
	} else {
		pd.payload_len = 127
		pd.payload_len2 = int64(lenp)
	}
	//pd.payload_data = []byte(event)
	return &pd
}

func newPongFrame() *PacketData {
	var pd PacketData
	pd.fin = 1
	pd.opcode = PongFrame
	lenp := 0
	if lenp <= 125 {
		pd.payload_len = lenp
	} else if lenp >= 126 && lenp < 65536 {
		pd.payload_len = 126
		pd.payload_len2 = int64(lenp)
	} else {
		pd.payload_len = 127
		pd.payload_len2 = int64(lenp)
	}
	//pd.payload_data = []byte(event)
	return &pd
}

func (pd *PacketData)decode(r io.Reader) (err error) {
	b := bufio.NewReader(r)
	firstByte, err := b.ReadByte()
	if err != nil {
		return err
	}
	pd.fin = int((( 1 << 7 ) & firstByte ) >> 7)
	pd.rsv1 = int((( 1 << 6 ) & firstByte ) >> 6)
	pd.rsv2 = int((( 1 << 5 ) & firstByte ) >> 5)
	pd.rsv3 = int((( 1 << 4 ) & firstByte ) >> 4)
	pd.opcode = opCode( ((1 << 3) | (1 << 2) | (1 << 1) | 1) & firstByte)
	pd.frameType = int(pd.opcode)
	secondByte, err := b.ReadByte()
	if err != nil {
		return err
	}
	pd.masked = int(( (1 << 7) & secondByte ) >> 7)
	pd.payload_len = int( (1 << 6 | 1 << 5 | 1 << 4 | 1 << 3 | 1 << 2 | 1 << 1 | 1) & secondByte)
	lenp := int64(pd.payload_len)

	if pd.payload_len == 126 {
		var p [2]byte
		_, err := io.ReadFull(b, p[:])
		if err != nil {
			return err
		}
		r := bytes.NewReader(p[:])
		binary.Read(r, binary.LittleEndian, &pd.payload_len2)
		lenp = pd.payload_len2
	} else if pd.payload_len == 127 {
		var p [8]byte
		_, err := io.ReadFull(b, p[:])
		if err != nil {
			return err
		}
		r := bytes.NewReader(p[:])
		binary.Read(r, binary.LittleEndian, &pd.payload_len2)
		lenp = pd.payload_len2
	}
	if pd.masked > 0 {
		var p [4]byte
		_, err := io.ReadFull(b, p[:])
		if err != nil {
			return err
		}
		pd.mask_key = p
	}
	// 没考虑，解码的时候，需要使用mask key
	p := make([]byte, lenp)
	_, err = io.ReadFull(b, p)
	if pd.masked > 0 {
		for idx, data := range p {
			pd.payload_data = append(pd.payload_data, data ^ pd.mask_key[idx % 4])
		}
	} else {
		pd.payload_data = p[:]
	}
	return nil
}

func (pd *PacketData)encode(w io.Writer) (err error)  {
	b := bufio.NewWriter(w)
	defer func() {
		if err == nil {
			err = b.Flush()
		}
	}()
	// 写入 fin, rsv1 rsv2, rsv3, opcode
	firstByte :=byte(pd.fin)	<< 7	|
		byte(pd.rsv1)	<< 6	|
		byte(pd.rsv2)	<< 5	|
		byte(pd.rsv3)	<< 4	|
		byte(pd.opcode)
	err = b.WriteByte(firstByte)
	if err != nil {
		return err
	}

	// 写入 mask, payload len
	var secondByte byte
	secondByte = byte(pd.masked)	<< 7	|
		byte(pd.payload_len)
	err = b.WriteByte(secondByte)
	if err != nil {
		return err
	}
	if pd.payload_len == 126 {
		bs := []byte(strconv.Itoa(int(pd.payload_len2)))[2:4]
		_, err = b.Write(bs)
		if err != nil {
			return err
		}
	} else if pd.payload_len == 127 {
		bs := []byte(strconv.Itoa(int(pd.payload_len2)))
		_, err = b.Write(bs)
		if err != nil {
			return err
		}
	}
	if pd.masked == 1 {
		_, err = b.Write(pd.mask_key[:])
		if err != nil {
			return err
		}
	}
	_, err = b.Write(pd.payload_data)
	return err
}


func (pd *PacketData)String() string {
	return strings.Join([]string{
		fmt.Sprintf("fin:%v", pd.fin),
		fmt.Sprintf("rsv1:%v", pd.rsv1),
		fmt.Sprintf("rsv2:%v", pd.rsv2),
		fmt.Sprintf("rsv3:%v", pd.rsv3),
		fmt.Sprintf("opcode:%v", pd.opcode),
		fmt.Sprintf("frameType: %v", pd.frameType),
		fmt.Sprintf("masked:%v", pd.masked),
		fmt.Sprintf("payload_len:%v", pd.payload_len),
		fmt.Sprintf("payload_len2:%v", pd.payload_len2),
		fmt.Sprintf("mask_key:%s", string(pd.mask_key[:])),
		fmt.Sprintf("payload_data:%v", string(pd.payload_data)),
	},"\n")
}