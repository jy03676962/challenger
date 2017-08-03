package core

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

var _ = log.Printf

const tcpSendMinInterval = 100

type InboxConnection interface {
	ReadJSON(v *InboxMessage) error
	WriteJSON(v *InboxMessage) error
	Close() error
	Accept(addr InboxAddress) bool
}

type InboxTcpConnection struct {
	conn    *net.TCPConn
	r       *bufio.Reader
	id      string
	ch      chan []byte
	closeCh chan struct{}
}

func NewInboxTcpConnection(conn *net.TCPConn) *InboxTcpConnection {
	tcp := InboxTcpConnection{conn: conn}
	tcp.r = bufio.NewReader(conn)
	tcp.ch = make(chan []byte, 1000)
	tcp.closeCh = make(chan struct{})
	go tcp.doWrite()
	return &tcp
}

func (tcp *InboxTcpConnection) Close() error {
	close(tcp.closeCh)
	return tcp.conn.Close()
}

func (tcp *InboxTcpConnection) ReadJSON(v *InboxMessage) error {
	tcp.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	b, e := tcp.r.ReadBytes(60) // tcp message frame start with '<'
	if e != nil {
		if tcp.id != "" {
			v.RemoveAddress = &InboxAddress{at(tcp.id), tcp.id}
			v.ShouldCloseConnection = true
		} else if e == io.EOF {
			v.ShouldCloseConnection = true
		}
		return e
	}
	tcp.conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	b, e = tcp.r.ReadBytes(62) // tcp message frame end with '>'
	if e != nil {
		if tcp.id != "" {
			v.RemoveAddress = &InboxAddress{at(tcp.id), tcp.id}
			v.ShouldCloseConnection = true
		} else if e == io.EOF {
			v.ShouldCloseConnection = true
		}
		return e
	}
	if tcp.id != "" {
		v.Address = &InboxAddress{at(tcp.id), tcp.id}
	}
	if len(b) == 1 { // only has '>' delimiter
		return nil
	}
	if b[0] == 123 { // first byte is '{', json encoding frame
		json.Unmarshal(b[:len(b)-1], &v.Data)
	} else { // parse heart beat frame
		parseTcpHB(string(b[:len(b)-1]), v)
		//v.SetCmd("hb")
		infoType := v.GetStr("TYPE")
		if infoType != "" {
			v.SetCmd(infoType)
		} else {
			v.SetCmd(UnKnown)
		}
		//v.SetCmd("hb")
		if id := v.GetStr("ID"); id != "" && tcp.id != id {
			v.AddAddress = &InboxAddress{at(id), id}
			v.Address = v.AddAddress
			if tcp.id != "" {
				v.RemoveAddress = &InboxAddress{at(tcp.id), tcp.id}
			}
			tcp.id = id
		}
	}
	return nil
}

func at(id string) InboxAddressType {
	if id == "" {
		return InboxAddressTypeUnknown
	} else if strings.HasPrefix(id, "G") {
		return InboxAddressTypeGameArduinoDevice
	} else if strings.HasPrefix(id, "B") {
		return InboxAddressTypeBoxArduinoDevice
	} else if strings.HasPrefix(id, "N") {
		return InboxAddressTypeNightArduino
	} else if strings.HasPrefix(id, "D") {
		return InboxAddressTypeDjArduino
	} else if strings.HasPrefix(id, "A") {
		return InboxAddressTypeAdminDevice
	}
	return InboxAddressTypeUnknown
}

// Tcp HB format is [key1]value1[key2]value2
func parseTcpHB(hb string, v *InboxMessage) {
	kvs := strings.Split(hb, "[")
	for _, s := range kvs {
		kv := strings.Split(s, "]")
		if len(kv) == 2 {
			v.Set(kv[0], kv[1])
		}
	}
}

func (tcp *InboxTcpConnection) WriteJSON(v *InboxMessage) error {
	b, e := v.Marshal()
	if e != nil {
		return e
	}

	buf := make([]byte, len(b)+2)
	for i := 1; i < len(buf)-1; i++ {
		buf[i] = b[i-1]
	}
	buf[0] = 60
	buf[len(buf)-1] = 62
	select {
	case tcp.ch <- buf:
	default:
	}
	return nil
}

func (tcp *InboxTcpConnection) doWrite() {
	for {
		select {
		case <-tcp.closeCh:
			return
		case bytes := <-tcp.ch:
			_, err := tcp.conn.Write(bytes)
			if err != nil {
				log.Printf("tcp written:%v, error:%v\n", string(bytes), err.Error())
			} else {
				//log.Printf("tcp written:%v\n", string(bytes))
			}
		}
		time.Sleep(tcpSendMinInterval * time.Millisecond)
	}
}

func (tcp *InboxTcpConnection) Accept(addr InboxAddress) bool {
	if addr.Type != at(tcp.id) {
		return false
	}
	return addr.ID == "" || addr.ID == tcp.id
}

type InboxWsConnection struct {
	conn *websocket.Conn
	t    InboxAddressType
	id   string
	l    *sync.RWMutex
}

func NewInboxWsConnection(conn *websocket.Conn) *InboxWsConnection {
	return &InboxWsConnection{conn: conn, l: new(sync.RWMutex)}
}

func (ws *InboxWsConnection) Close() error {
	return ws.conn.Close()
}

func (ws *InboxWsConnection) ReadJSON(v *InboxMessage) error {
	e := websocket.JSON.Receive(ws.conn, &v.Data)
	if e != nil {
		id, t := ws.getAddressInfo()
		if id != "" {
			v.RemoveAddress = &InboxAddress{t, id}
		}
		v.ShouldCloseConnection = true
		return e
	}
	if v.GetCmd() == "init" {
		tt, _ := strconv.Atoi(v.Get("TYPE").(string))
		t := InboxAddressType(tt)
		id := v.GetStr("ID")
		oldid, oldt := ws.getAddressInfo()
		if oldid != id {
			v.AddAddress = &InboxAddress{t, id}
			v.Address = v.AddAddress
			if oldid != "" {
				v.RemoveAddress = &InboxAddress{oldt, oldid}
			}
			ws.setAddressInfo(id, t)
		}
	} else {
		id, t := ws.getAddressInfo()
		if id != "" {
			v.Address = &InboxAddress{t, id}
		}
	}
	return nil
}

func (ws *InboxWsConnection) WriteJSON(v *InboxMessage) error {
	return websocket.JSON.Send(ws.conn, v.Data)
}

func (ws *InboxWsConnection) Accept(addr InboxAddress) bool {
	if id, t := ws.getAddressInfo(); id != "" {
		if t != addr.Type {
			return false
		}
		return addr.ID == "" || id == addr.ID
	}
	return false
}

func (ws *InboxWsConnection) getAddressInfo() (string, InboxAddressType) {
	ws.l.RLock()
	defer ws.l.RUnlock()
	return ws.id, ws.t
}

func (ws *InboxWsConnection) setAddressInfo(id string, t InboxAddressType) {
	ws.l.Lock()
	defer ws.l.Unlock()
	ws.id, ws.t = id, t
}
