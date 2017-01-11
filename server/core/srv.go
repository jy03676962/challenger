package core

import (
	"log"
	"net"
	"os"
	"strconv"

	"golang.org/x/net/websocket"
)

type AdminMode int

const (
	AdminModeNormal = iota
	AdminModeDebug  = iota
)

var _ = log.Println

type Srv struct {
	inbox            *Inbox
	inboxMessageChan chan *InboxMessage
	mChan            chan MatchEvent
	aDict            map[string]*ArduinoController
	match            *Match
	adminMode        AdminMode
	isSimulator      bool
}

func NewSrv(isSimulator bool) *Srv {
	s := Srv{}
	s.isSimulator = isSimulator
	s.inbox = NewInbox(&s)
	s.inboxMessageChan = make(chan *InboxMessage, 1)
	s.mChan = make(chan MatchEvent)
	s.aDict = make(map[string]*ArduinoController)
	s.adminMode = AdminModeNormal
	s.initArduinoControllers()
	return &s
}

func (s *Srv) Run(tcpAddr string, adminAddr string, dbPath string) {
	go s.listenTcp(tcpAddr)
	go s.listenTcp(adminAddr)
	s.mainLoop()
}

func (s *Srv) ListenWebSocket(conn *websocket.Conn) {
	log.Println("got new ws connection")
	s.inbox.ListenConnection(NewInboxWsConnection(conn))
}

// http interface

func (s *Srv) mainLoop() {
	for {
		select {
		case msg := <-s.inboxMessageChan:
			s.handleInboxMessage(msg)
		case evt := <-s.mChan:
			s.handleMatchEvent(evt)
		}
	}
}

func (s *Srv) listenTcp(address string) {
	tcpAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Println("resolve tcp address error:", err.Error())
		os.Exit(1)
	}
	lr, err := net.ListenTCP("tcp", tcpAddress)
	if err != nil {
		log.Println("listen tcp error:", err.Error())
		os.Exit(1)
	}
	defer lr.Close()
	log.Println("listen tcp:", address)
	for {
		conn, err := lr.AcceptTCP()
		//conn.SetKeepAlive(true)
		if err != nil {
			log.Println("tcp listen error: ", err.Error())
		} else {
			log.Printf("got new tcp connection:%v\n", conn.RemoteAddr())
			go s.inbox.ListenConnection(NewInboxTcpConnection(conn))
		}
	}
}

func (s *Srv) listenAdmin(address string) {
	tcpAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Println("resolve tcp address error:", err.Error())
		os.Exit(1)
	}
	lr, err := net.ListenTCP("tcp", tcpAddress)
	if err != nil {
		log.Println("listen tcp error:", err.Error())
		os.Exit(1)
	}
	defer lr.Close()
	log.Println("listen tcp:", address)
	for {
		conn, err := lr.AcceptTCP()
		//conn.SetKeepAlive(true)
		if err != nil {
			log.Println("tcp listen error: ", err.Error())
		} else {
			log.Printf("got new tcp connection:%v\n", conn.RemoteAddr())
			go s.inbox.ListenConnection(NewInboxTcpConnection(conn))
		}
	}
}

func (s *Srv) onInboxMessageArrived(msg *InboxMessage) {
	s.inboxMessageChan <- msg
}

func (s *Srv) onMatchEvent(evt MatchEvent) {
	s.mChan <- evt
}

func (s *Srv) handleMatchEvent(evt MatchEvent) {
	switch evt.Type {
	case MatchEventTypeEnd:
	case MatchEventTypeUpdate:
	}
}

func (s *Srv) handleInboxMessage(msg *InboxMessage) {
	if msg.Address == nil {
		log.Printf("message has no address:%v\n", msg.Data)
		return
	}
	cmd := msg.GetCmd()
	if len(cmd) == 0 {
		log.Printf("message has no cmd:%v\n", msg.Data)
		return
	}
	switch msg.Address.Type {
	case InboxAddressTypeAdminDevice:
		s.handleAdminMessage(msg)
	case InboxAddressTypeRoomArduinoDevice:
		s.handleArduinoMessage(msg)
	case InboxAddressTypeDoorArduino:
		s.handleArduinoMessage(msg)
	case InboxAddressTypeMusicArduino:
		s.handleArduinoMessage(msg)
	}
}

func (s *Srv) handleArduinoMessage(msg *InboxMessage) {
	if s.match != nil {
		s.match.OnMatchCmdArrived(msg)
	}
}

func (s *Srv) handlePostGameMessage(msg *InboxMessage) {
	switch msg.GetCmd() {
	case "init":
		s.sendMsg("init", nil, msg.Address.ID, msg.Address.Type)
	}
}

func (s *Srv) handleAdminMessage(msg *InboxMessage) {
	switch msg.GetCmd() {
	case "init":
		if s.match != nil {
			s.match.reset()
		}
		s.sendMsg("init", nil, msg.Address.ID, msg.Address.Type)
	case "gameStart":
		s.startNewMatch()
	case "arduinoModeChange":
		mode := strconv.Itoa(int(ArduinoMode(msg.Get("mode").(float64))))
		am := NewInboxMessage()
		am.SetCmd("mode_change")
		am.Set("mode", mode)
		log.Printf("send mode change:%v\n", mode)
	case "queryArduinoList":
		arduinolist := make([]ArduinoController, len(s.aDict))
		i := 0
		for _, controller := range s.aDict {
			arduinolist[i] = *controller
			i += 1
		}
		s.sendMsg("ArduinoList", arduinolist, msg.Address.ID, msg.Address.Type)
	case "stopMatch":
		s.match.OnMatchCmdArrived(msg)
	case "nextStep":
		if s.match != nil {
			s.match.OnMatchCmdArrived(msg)
		}
	}
}

func (s *Srv) startNewMatch() {
	m := NewMatch(s)
	s.match = m
	go m.Run()
}

func (s *Srv) sendMsg(cmd string, data interface{}, id string, t InboxAddressType) {
	addr := InboxAddress{t, id}
	s.sendMsgToAddresses(cmd, data, []InboxAddress{addr})
}

func (s *Srv) sendMsgs(cmd string, data interface{}, types ...InboxAddressType) {
	addrs := make([]InboxAddress, len(types))
	for i, t := range types {
		addrs[i] = InboxAddress{t, ""}
	}
	s.sendMsgToAddresses(cmd, data, addrs)
}

func (s *Srv) sendMsgToAddresses(cmd string, data interface{}, addrs []InboxAddress) {
	msg := NewInboxMessage()
	msg.SetCmd(cmd)
	if data != nil {
		msg.Set("data", data)
	}
	s.send(msg, addrs)
}

func (s *Srv) sends(msg *InboxMessage, types ...InboxAddressType) {
	addrs := make([]InboxAddress, len(types))
	for i, t := range types {
		addrs[i] = InboxAddress{t, ""}
	}
	s.send(msg, addrs)
}

func (s *Srv) doorControl(IL string, OL string, ID string) {
	addr := InboxAddress{InboxAddressTypeDoorArduino, ID}
	msg := NewInboxMessage()
	msg.SetCmd("led_ctrl")
	controls := make([]map[string]string, 0)
	if len(IL) > 0 {
		controls = append(controls, map[string]string{
			"wall":  "IL",
			"led_t": "1",
			"mode":  IL,
		})
	}
	if len(OL) > 0 {
		controls = append(controls, map[string]string{
			"wall":  "OL",
			"led_t": "1",
			"mode":  OL,
		})
	}
	msg.Set("led", controls)
	s.sendToOne(msg, addr)
}

func (s *Srv) send(msg *InboxMessage, addrs []InboxAddress) {
	s.inbox.Send(msg, addrs)
}

func (s *Srv) sendToOne(msg *InboxMessage, addr InboxAddress) {
	s.send(msg, []InboxAddress{addr})
}

func (s *Srv) initArduinoControllers() {
	for _, roomArduino := range GetOptions().RoomArduino {
		addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, roomArduino}
		controller := NewArduinoController(addr)
		s.aDict[addr.String()] = controller
	}
	for _, lightArduino := range GetOptions().LightArduino {
		addr := InboxAddress{InboxAddressTypeLightArduinoDevice, lightArduino}
		controller := NewArduinoController(addr)
		s.aDict[addr.String()] = controller
	}
	for _, music := range GetOptions().MusicArduino {
		addr := InboxAddress{InboxAddressTypeMusicArduino, music}
		controller := NewArduinoController(addr)
		s.aDict[addr.String()] = controller
	}
	for _, door := range GetOptions().DoorArduino {
		addr := InboxAddress{InboxAddressTypeDoorArduino, door}
		controller := NewArduinoController(addr)
		s.aDict[addr.String()] = controller
	}
}

func (s *Srv) bgmControl(music string) {
	msg := NewInboxMessage()
	msg.SetCmd("mp3_ctrl")
	msg.Set("music", music)
	s.sends(msg, InboxAddressTypeMusicArduino)
}

func (s *Srv) fakeBooksControl(mode string, music string, id string) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("fake_book")
	sendMsg.Set("mode", mode)
	sendMsg.Set("time", strconv.FormatFloat(GetOptions().FakeAnimationTime, 'f', 0, 64))
	sendMsg.Set("music", music)
	addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, id}
	s.sendToOne(sendMsg, addr)
}

func (s *Srv) candlesControl(candles []map[string]string, id string) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("led_candle")
	sendMsg.Set("candles", candles)
	addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, id}
	s.sendToOne(sendMsg, addr)
}

func (s *Srv) stairRoomCandlesCtrl(status string, id string) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("candle_ctrl")
	sendMsg.Set("status", status)
	addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, id}
	s.sendToOne(sendMsg, addr)

}

func (s *Srv) powerStatus(power []map[string]string) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("magic_table")
	sendMsg.Set("power", power)
	addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-5"}
	s.sendToOne(sendMsg, addr)
}

func (s *Srv) starControl(lights []map[string]string, leds []map[string]string) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("star_led")
	sendMsg.Set("light", lights)
	sendMsg.Set("led", leds)
	addrs := []InboxAddress{
		{InboxAddressTypeRoomArduinoDevice, "R-5-1"},
		{InboxAddressTypeRoomArduinoDevice, "R-5-2"},
		{InboxAddressTypeRoomArduinoDevice, "R-5-3"},
		{InboxAddressTypeRoomArduinoDevice, "R-5-4"},
		{InboxAddressTypeRoomArduinoDevice, "R-5-5"}}
	s.send(sendMsg, addrs)

}
