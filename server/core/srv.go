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

//Arduino msg type
const (
	UnKnown          = "unknown"
	Hbt              = "0"
	GameStartForward = "1"
	GameStart        = "2"
	GameEndForward   = "3"
	GameEnd          = "4"
	GameData         = "5"
	AuthorityCheck   = "6"
	TicketGet        = "7"
	BoxStatus        = "8"
	ResetGame        = "9"
)

var _ = log.Println

type Srv struct {
	inbox            *Inbox
	inboxMessageChan chan *InboxMessage
	mChan            chan MatchEvent
	httpResChan      chan *HttpResponse
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
	s.httpResChan = make(chan *HttpResponse, 1)
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
		case httpRes := <-s.httpResChan:
			s.handleHttpMessage(httpRes)
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

func (s *Srv) OnHttpRequest(msg *HttpResponse) {
	s.httpResChan <- msg
}

//http msg type
func (s *Srv) handleHttpMessage(httpRes *HttpResponse) {
	switch httpRes.Api {
	case AuthorityGet:
		arduinoId := httpRes.ArduinoId
		res := httpRes.Data
		log.Println("arduinoId:", arduinoId, "need feedback! and res :", res)
	case GameDataAdivinacionCreate:
	case GameDataAdivinacionModify:
	case GameDataBangCreate:
	case GameDataBangModify:
	case GameDataFollowCreate:
	case GameDataFollowModify:
	case GameDataGreetingCreate:
	case GameDataGreetingModify:
	case GameDataHighnoonCreate:
	case GameDataHighnoonModify:
	case GameDataHunterCreate:
	case GameDataHunterModify:
	case GameDataHunterBoxCreate:
	case GameDataHunterBoxModify:
	case GameDataMarksmanCreate:
	case GameDataMarksmanModify:
	case GameDataMinerCreate:
	case GameDataMinerModify:
	case GameDataPrivityCreate:
	case GameDataPrivityModify:
	case GameDataRussianCreate:
	case GameDataRussianModify:
	case TicketUse:
	case TicketCheck:
		arduinoId := httpRes.ArduinoId
		res := httpRes.Data
		log.Println("arduinoId:", arduinoId, "need feedback! and res :", res)
	}
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
	case InboxAddressTypeGameArduinoDevice:
		s.handleArduinoMessage(msg)
	case InboxAddressTypeDoorArduino:
		s.handleArduinoMessage(msg)
	case InboxAddressTypeMusicArduino:
		s.handleArduinoMessage(msg)
	}
}

func (s *Srv) handleArduinoMessage(msg *InboxMessage) {
	cmd := msg.GetCmd()
	switch cmd {
	case UnKnown:
		log.Println(msg.Get("ID"), "send UnKnown cmd!")
	case Hbt:
		log.Println("Receive htb:", msg.Get("ID"))
	case GameStartForward:
		admin := msg.Get("ADMIN")
		gameId := msg.Get("GAME")
		arduino := msg.Get("ARDUINO")
		log.Println("Game:", gameId, "start and forward to ", arduino, "! operator:", admin)
	case GameStart:
		admin := msg.Get("ADMIN")
		gameId := msg.Get("GAME")
		log.Println("Game:", gameId, "start! operator:", admin)
	case GameEndForward:
		admin := msg.Get("ADMIN")
		gameId := msg.Get("GAME")
		arduino := msg.Get("ARDUINO")
		log.Println("Game:", gameId, "end and forward to ", arduino, "! operator:", admin)
	case GameEnd:
		admin := msg.Get("ADMIN")
		gameId := msg.Get("GAME")
		log.Println("Game:", gameId, "end! operator:", admin)
	case GameData:
		gameId := msg.Get("GAME")
		log.Println("Receive Game:", gameId, "'s data!")
	case AuthorityCheck:
		//arduinoId := msg.Get("ID") //创建request的时候需要放入
		cardId := msg.Get("CARD_ID")
		authorityId := msg.Get("AR")
		log.Println("Get the card:", cardId, "  AuthrorityId:", authorityId)
	case TicketGet:
		//arduinoId := msg.Get("ID") //创建request的时候需要放入
		admin := msg.Get("ADMIN")
		gameId := msg.Get("GAME")
		cardId := msg.Get("CARD_ID")
		log.Println("Ticket Get：  CardId:", cardId, "GameId:", gameId, " Admin:", admin)
	case BoxStatus:
		admin := msg.Get("ADMIN")
		boxId := msg.Get("BOX_ID")
		boxStatus := msg.Get("ST")
		switch boxStatus {
		case "0":
			log.Println("Box:", boxId, "has'n been opened by player!")
		case "1":
			log.Println("Box:", boxId, "has been opened by player! Watting reset!")
		case "2":
			log.Println("Box:", boxId, "has been reset by admin:", admin, "!")
		}
	case ResetGame:
		admin := msg.Get("ADMIN")
		gameId := msg.Get("GAME")
		resetGame(gameId)
		log.Println("Admin:", admin, " reset the game:", gameId, "!")
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
	case "gameStart":
	case "queryGameInfo":
	case "nextStep":
	case "gameOver":
	case "completed":
	case "launch":
	case "lightOff":
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
		addr := InboxAddress{InboxAddressTypeGameArduinoDevice, roomArduino}
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

func resetGame(gameId string)  {
	switch gameId {
	
	}	
}
