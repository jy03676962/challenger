package core

import (
	"log"
	"net"
	"os"

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
	if msg.RemoveAddress != nil && msg.RemoveAddress.Type.IsArduinoControllerType() {
		id := msg.RemoveAddress.String()
		if controller := s.aDict[id]; controller != nil {
			controller.Online = false
		}
		//s.sendMsgs("removeTCP", msg.RemoveAddress, InboxAddressTypeArduinoTestDevice)
	}

	if msg.AddAddress != nil && msg.AddAddress.Type.IsArduinoControllerType() {
		if controller := s.aDict[msg.AddAddress.String()]; controller != nil {
			controller.Online = true
		} else {
			log.Printf("Warning: get arduino connection not belong to list:%v\n", msg.AddAddress.String())
		}
		//s.sendMsgs("addTCP", msg.AddAddress, InboxAddressTypeArduinoTestDevice)
	}
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
		sendMsg1 := NewInboxMessage()
		if s.match != nil {
			s.match.reset()
			sendMsg1.SetCmd("reset success!")
		} else {
			sendMsg1.SetCmd("game has'n started,can't reset the game")
		}
		addr := InboxAddress{msg.Address.Type, msg.Address.ID}
		s.sendToOne(sendMsg1, addr)
		//s.sendMsg("init", nil, msg.Address.ID, msg.Address.Type)
	case "gameStart":
		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("start game")
		addr := InboxAddress{msg.Address.Type, msg.Address.ID}
		s.sendToOne(sendMsg1, addr)
		if s.match == nil {
			s.startNewMatch()
			s.match.setStage(StageRoom1)
			log.Println("create new match")
		} else {
			s.match.setStage(StageRoom1)
			log.Println("start game")
		}
	case "queryGameInfo":
		msg1 := NewInboxMessage()
		msg1.SetCmd("GameInfo")
		arduinolist := make([]ArduinoController, len(s.aDict))
		i := 0
		for _, controller := range s.aDict {
			arduinolist[i] = *controller
			i += 1
		}
		msg1.Set("ArduinoList", arduinolist)
		if s.match != nil {
			msg1.Set("CurrentTime", s.match.TotalTime)
			if s.match.Stage == StageRoom6 && s.match.endRoom.Step == 3 {
				if s.match.endRoom.Ending == 1 {
					msg1.Set("CurrentRoom", "Good Ending!")
				} else {
					msg1.Set("CurrentRoom", "Bad Ending!")
				}
			} else {
				msg1.Set("CurrentRoom", s.match.Stage)
				msg1.Set("CurrentStep", s.match.Step)
			}
		} else {
			msg1.Set("CurrentTime", 0.00)
			msg1.Set("CurrentRoom", "has'n started!")
			msg1.Set("CurrentStep", 0)
		}
		msg1.Set("TotalTime", GetOptions().TotalTime)
		//log.Println(msg)
		addr := InboxAddress{msg.Address.Type, msg.Address.ID}
		s.sendToOne(msg1, addr)
		//s.sendMsg("ArduinoList", arduinolist, msg.Address.ID, msg.Address.Type)
	case "nextStep":
		sendMsg1 := NewInboxMessage()
		if s.match != nil {
			s.match.OnMatchCmdArrived(msg)
			sendMsg1.SetCmd("next success!")
		} else {
			sendMsg1.SetCmd("game has'n started,can't reset the game")
		}
		addr := InboxAddress{msg.Address.Type, msg.Address.ID}
		s.sendToOne(sendMsg1, addr)
	case "gameOver":
		sendMsg1 := NewInboxMessage()
		addr := InboxAddress{msg.Address.Type, msg.Address.ID}
		s.sendToOne(sendMsg1, addr)

		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("mode_change")
		sendMsg.Set("mode", "0")
		s.sends(sendMsg, InboxAddressTypeRoomArduinoDevice)

		openDoor := NewInboxMessage()
		openDoor.SetCmd("door_ctrl")
		openDoor.Set("status", "1")
		openDoor.Set("time", "200")
		s.sends(openDoor, InboxAddressTypeDoorArduino)

		closeMusic := NewInboxMessage()
		closeMusic.SetCmd("mp3_ctrl")
		closeMusic.Set("music", "0")
		s.sends(closeMusic, InboxAddressTypeMusicArduino)
		if s.match != nil {
			sendMsg1.SetCmd("game over successs")
			s.match.setStage(StageEnd)
		} else {
			sendMsg1.SetCmd("game has'n started!")
		}
	case "completed":
		sendMsg1 := NewInboxMessage()
		log.Println("completed!")
		if s.match != nil {
			if s.match.Stage == StageRoom6 && s.match.endRoom.Step == 4 {
				s.match.setStage(StageEnd)
				sendMsg1.SetCmd("game over!")
			} else {
				sendMsg1.SetCmd("condition mismatch!")
			}
		} else {
			sendMsg1.SetCmd("game has'n started!")
		}
		addr := InboxAddress{msg.Address.Type, msg.Address.ID}
		s.sendToOne(sendMsg1, addr)
	case "launch":
		sendMsg1 := NewInboxMessage()
		log.Println("launch!")
		if s.match != nil {
			if s.match.Stage == StageRoom6 && s.match.endRoom.Step == 1 {
				s.match.endRoom.Step = 2
				sendMsg1.SetCmd("launch success!")
			} else {
				sendMsg1.SetCmd("condition mismatch!")
			}
		} else {
			sendMsg1.SetCmd("game has'n started!")
		}
		addr := InboxAddress{msg.Address.Type, msg.Address.ID}
		s.sendToOne(sendMsg1, addr)
	case "lightOff":
		sendMsg1 := NewInboxMessage()
		log.Println("lightOff!")
		if s.match != nil {
			if s.match.Stage == StageRoom6 && s.match.endRoom.Step == 3 {
				s.match.endRoom.Step = 4
				sendMsg1.SetCmd("light off success!")
			} else {
				sendMsg1.SetCmd("condition mismatch!")
			}
		} else {
			sendMsg1.SetCmd("game has'n started!")
		}
		addr := InboxAddress{msg.Address.Type, msg.Address.ID}
		s.sendToOne(sendMsg1, addr)
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
