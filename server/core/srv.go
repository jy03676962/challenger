package core

import (
	"log"
	"net"
	"os"

	"golang.org/x/net/websocket"
	"strconv"
	//"golang.org/x/net/html/atom"
	//"regexp"
	//"math"
	"fmt"
	"time"
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
	//match            *Match
	adminMode   AdminMode
	isSimulator bool
	//--------game info------------
	adivainacion *Adivainacion
	bang         *Bang
	follow       *Follow
	greeting     *Greeting
	highnoon     *Highnoon
	hunter       *Hunter
	marksman     *Marksman
	miner        *Miner
	privity      *Privity
	russian      *Russian
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
	s.initGameInfo()
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
	case GameDataAdivinacionCreate:
		fallthrough
	case GameDataAdivinacionModify:
		fallthrough
	case GameDataBangCreate:
		fallthrough
	case GameDataBangModify:
		fallthrough
	case GameDataFollowCreate:
		fallthrough
	case GameDataFollowModify:
		fallthrough
	case GameDataGreetingCreate:
		fallthrough
	case GameDataGreetingModify:
		fallthrough
	case GameDataHighnoonCreate:
		fallthrough
	case GameDataHighnoonModify:
		fallthrough
	case GameDataHunterCreate:
		fallthrough
	case GameDataHunterModify:
		fallthrough
	case GameDataHunterBoxCreate:
		fallthrough
	case GameDataHunterBoxModify:
		fallthrough
	case GameDataMarksmanCreate:
		fallthrough
	case GameDataMarksmanModify:
		fallthrough
	case GameDataMinerCreate:
		fallthrough
	case GameDataMinerModify:
		fallthrough
	case GameDataPrivityCreate:
		fallthrough
	case GameDataPrivityModify:
		fallthrough
	case GameDataRussianCreate:
		fallthrough
	case GameDataRussianModify:
		log.Println("data server res:", httpRes.JsonData)
		if res, ok := httpRes.Get("return").(bool); ok {
			gameId, _ := strconv.Atoi(httpRes.Msg.GetStr("GAME"))
			if !res {
				s.uploadGameInfo(httpRes.Msg, gameId)
			}
		}
	case AuthorityGet:
		if res, ok := httpRes.Get("return").(bool); ok {
			fmt.Println("authority:", res)
			arduinoId := httpRes.Msg.GetStr("ID")
			addr := InboxAddress{InboxAddressTypeGameArduinoDevice, arduinoId}
			msg := NewInboxMessage()
			msg.SetCmd("authority_check")
			msg.Set("return", res)
			s.sendToOne(msg, addr)
		}
	case TicketUse:
		if res, ok := httpRes.Get("return").(bool); ok {
			gameId, _ := strconv.Atoi(httpRes.Msg.GetStr("GAME"))
			if !res {
				log.Println("Modify Ticket failed!")
				s.gameStart(gameId, httpRes.Msg)
			}
		}
	case TicketCheck:
		arduinoId := httpRes.Msg.GetStr("ID")
		gameId, _ := strconv.Atoi(httpRes.Msg.GetStr("GAME"))
		ticketId := httpRes.Get("ticket_game_id").(int)
		if ticketId != -1 {
			s.loginGame(strconv.Itoa(ticketId), gameId, httpRes.Msg)
		}
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
		log.Println(msg.GetStr("ID"), "send UnKnown cmd!")
	case Hbt:
		//log.Println("Receive htb:", msg.GetStr("ID"))
	case GameStartForward:
		admin := msg.GetStr("ADMIN")
		//gameId := msg.GetStr("GAME")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		arduino := msg.GetStr("ARDUINO")
		log.Println("Game:", gameId, "start and forward to ", arduino, "! operator:", admin)
		s.gameStart(gameId, msg)
	case GameStart:
		admin := msg.GetStr("ADMIN")
		//gameId := msg.GetStr("GAME")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		log.Println("Game:", gameId, "start! operator:", admin)
		s.gameStart(gameId, msg)
	case GameEndForward:
		admin := msg.GetStr("ADMIN")
		//gameId := msg.GetStr("GAME")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		arduino := msg.GetStr("ARDUINO")
		log.Println("Game:", gameId, "end and forward to ", arduino, "! operator:", admin)
		s.gameEnd(msg, gameId)
	case GameEnd:
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		s.gameEnd(msg, gameId)
		log.Println("Game:", gameId, "end!")
	case GameData:
		//gameId := msg.GetStr("GAME")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		log.Println("Receive Game:", gameId, "'s data!")
		s.upldateGameInfo(msg, gameId)
	case AuthorityCheck:
		arduinoId := msg.GetStr("ID") //创建request的时候需要放入
		cardId := msg.GetStr("CARD_ID")
		authorityId := msg.GetStr("AR")
		log.Println("Get the card:", cardId, "  AuthrorityId:", authorityId, " ArduinoId:", arduinoId)
		request := NewHttpRequest(s)
		request.SetApi(AuthorityGet)
		//request.SetCardId(cardId)
		//request.SetArduinoId(arduinoId)
		params := make(map[string]string)
		//params["card_Uid"] = "00FF0FF000FFCF4D54B110484EBAF95B4EB0"
		params["card_Uid"] = cardId
		params["authority_ID"] = authorityId
		//params["op"] = "get_user_authid"
		params["op"] = "validate_authid"
		request.SetParams(params)
		request.SetMsg(msg)
		request.DoGet()
	case TicketGet:
		//arduinoId := msg.GetStr("ID") //创建request的时候需要放入
		admin := msg.GetStr("ADMIN")
		gameId := msg.GetStr("GAME")
		cardId := msg.GetStr("CARD_ID")
		log.Println("Ticket Get：  CardId:", cardId, "GameId:", gameId, " Admin:", admin)
		request := NewHttpRequest(s)
		request.SetApi(TicketCheck)
		//request.SetCardId(cardId)
		//request.SetArduinoId(arduinoId)
		params := make(map[string]string)
		params["card_Uid"] = cardId
		params["game_ID"] = gameId
		params["op"] = "get_ticket_game_id"
		request.SetParams(params)
		request.SetMsg(msg)
		request.DoGet()
	case BoxStatus:
		admin := msg.GetStr("ADMIN")
		boxId := msg.GetStr("BOX_ID")
		boxStatus := msg.GetStr("ST")
		switch boxStatus {
		case "0":
			log.Println("Box:", boxId, "has'n been opened by player!")
		case "1":
			log.Println("Box:", boxId, "has been opened by player! Watting reset!")
		case "2":
			log.Println("Box:", boxId, "has been reset by admin:", admin, "!")
		}
	case ResetGame:
		admin := msg.GetStr("ADMIN")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		s.resetGame(gameId)
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

//func (s *Srv) startNewMatch() {
//	m := NewMatch(s)
//	s.match = m
//	go m.Run()
//}

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

func (s *Srv) initGameInfo() {
	s.adivainacion = NewAdivainacion()
	s.bang = NewBang()
	s.follow = NewFollow()
	s.greeting = NewGreeting()
	s.highnoon = NewHighnoon()
	s.hunter = NewHunter()
	s.marksman = NewMarksman()
	s.miner = NewMiner()
	s.privity = NewPrivity()
	s.russian = NewRussian()
}

func (s *Srv) bgmControl(music string) {
	msg := NewInboxMessage()
	msg.SetCmd("mp3_ctrl")
	msg.Set("music", music)
	s.sends(msg, InboxAddressTypeMusicArduino)
}

func (s *Srv) loginGame(ticketId string, gameId int, msg *InboxMessage) {
	cardId := msg.GetStr("CARD_ID")
	switch gameId {
	case ID_Russian:
		if s.russian.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.russian.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.russian.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.russian.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.russian.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Adivainacion:
		if s.adivainacion.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.adivainacion.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.adivainacion.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.adivainacion.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.adivainacion.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Bang:
		if s.bang.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.bang.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.bang.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.bang.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.bang.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Follow:
		if s.follow.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.follow.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.follow.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.follow.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.follow.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Greeting:
		if s.greeting.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.greeting.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.greeting.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.greeting.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.greeting.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Highnoon:
		if s.highnoon.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.highnoon.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.highnoon.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.highnoon.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.highnoon.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Hunter:
		if s.hunter.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.hunter.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.hunter.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.hunter.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.hunter.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Marksman:
		if s.marksman.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.marksman.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.marksman.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.marksman.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.marksman.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Miner:
		if s.miner.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.miner.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.miner.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.miner.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.miner.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	case ID_Privity:
		if s.privity.LoginInfo.PlayerCardInfo["1p"] != "" {
			s.privity.LoginInfo.PlayerCardInfo["1p"] = cardId
			s.privity.LoginInfo.CardTicketInfo[cardId] = ticketId
		} else {
			s.privity.LoginInfo.PlayerCardInfo["2p"] = cardId
			s.privity.LoginInfo.CardTicketInfo[cardId] = ticketId
		}
	}
}

func (s *Srv) gameStart(gameId int, msg *InboxMessage) {
	admin := msg.GetStr("ADMIN")
	request := NewHttpRequest(s)
	request.SetMsg(msg)
	request.SetApi(TicketUse)
	params := make(map[string]string)
	switch gameId {
	case ID_Russian:
		s.russian.Time_start = currentTime()
		s.russian.LoginInfo.IsUploadInfo = true
		cardId_1p := s.russian.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.russian.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.russian.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.russian.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Adivainacion:
		s.adivainacion.Time_start = currentTime()
		s.adivainacion.LoginInfo.IsUploadInfo = true
		cardId_1p := s.adivainacion.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.adivainacion.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.adivainacion.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.adivainacion.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Bang:
		s.bang.Time_start = currentTime()
		s.bang.LoginInfo.IsUploadInfo = true
		cardId_1p := s.bang.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.bang.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.bang.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.bang.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Follow:
		s.follow.Time_start = currentTime()
		s.follow.LoginInfo.IsUploadInfo = true
		cardId_1p := s.follow.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.follow.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.follow.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.follow.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Greeting:
		s.greeting.Time_start = currentTime()
		s.greeting.LoginInfo.IsUploadInfo = true
		cardId_1p := s.greeting.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.greeting.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.greeting.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.greeting.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Highnoon:
		s.highnoon.Time_start = currentTime()
		s.highnoon.LoginInfo.IsUploadInfo = true
		cardId_1p := s.highnoon.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.highnoon.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.highnoon.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.highnoon.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Hunter:
		s.hunter.Time_start = currentTime()
		s.hunter.LoginInfo.IsUploadInfo = true
		cardId_1p := s.hunter.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.hunter.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.hunter.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.hunter.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Marksman:
		s.marksman.Time_start = currentTime()
		s.marksman.LoginInfo.IsUploadInfo = true
		cardId_1p := s.marksman.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.marksman.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.marksman.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.marksman.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Miner:
		s.miner.Time_start = currentTime()
		s.miner.LoginInfo.IsUploadInfo = true
		cardId_1p := s.miner.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.miner.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.miner.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.miner.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	case ID_Privity:
		s.privity.Time_start = currentTime()
		s.privity.LoginInfo.IsUploadInfo = true
		cardId_1p := s.privity.LoginInfo.PlayerCardInfo["1p"]
		cardId_2p := s.privity.LoginInfo.PlayerCardInfo["2p"]
		params["game_ID"] = strconv.Itoa(gameId)
		params["exchanger_ID"] = admin
		params["id"] = s.privity.LoginInfo.CardTicketInfo[cardId_1p]
		request.SetParams(params)
		request.DoPost()
		if cardId_2p != "" {
			params["id"] = s.privity.LoginInfo.CardTicketInfo[cardId_2p]
			request.SetParams(params)
			request.DoPost()
		}
	}
}

func (s *Srv) gameEnd(msg *InboxMessage, gameId int) {
	switch gameId {
	case ID_Russian:
		s.russian.LoginInfo.IsUploadInfo = true
		s.russian.Time_start = currentTime()
		s.russian.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.russian.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.russian.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.russian.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
	case ID_Adivainacion:
		s.adivainacion.LoginInfo.IsUploadInfo = true
		s.adivainacion.Time_start = currentTime()
		s.adivainacion.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.adivainacion.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.adivainacion.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.adivainacion.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
	case ID_Bang:
		s.bang.LoginInfo.IsUploadInfo = true
		s.bang.Time_start = currentTime()
		s.bang.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.bang.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.bang.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.bang.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
		s.bang.Point_round[1] = "6"
		s.bang.Point_round[2] = "5"
		s.bang.Point_round[3] = "4"
	case ID_Follow:
		s.follow.LoginInfo.IsUploadInfo = true
		s.follow.Time_start = currentTime()
		s.follow.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.follow.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.follow.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.follow.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
		s.follow.Last_round = "10"
	case ID_Greeting:
		s.greeting.LoginInfo.IsUploadInfo = true
		s.greeting.Time_start = currentTime()
		s.greeting.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.greeting.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.greeting.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.greeting.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
	case ID_Highnoon:
		s.highnoon.LoginInfo.IsUploadInfo = true
		s.highnoon.Time_start = currentTime()
		s.highnoon.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.highnoon.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.highnoon.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.highnoon.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
		s.highnoon.Result_round_1p[1] = "0.11"
		s.highnoon.Result_round_2p[1] = "0.12"
		s.highnoon.Result_round_1p[2] = "0.21"
		s.highnoon.Result_round_2p[2] = "0.22"
		s.highnoon.Result_round_1p[3] = "0.31"
		s.highnoon.Result_round_2p[3] = "0.32"
		s.highnoon.Result_round_1p[4] = "0.41"
		s.highnoon.Result_round_2p[4] = "0.42"
		s.highnoon.Result_round_1p[5] = "0.51"
		s.highnoon.Result_round_2p[5] = "0.52"
		s.highnoon.Result_round_1p[6] = "0.61"
		s.highnoon.Result_round_2p[6] = "0.62"
		s.highnoon.Result_round_1p[7] = ""
		s.highnoon.Result_round_2p[7] = ""
	case ID_Hunter:
		s.hunter.LoginInfo.IsUploadInfo = true
		s.hunter.Time_start = currentTime()
		s.hunter.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.hunter.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.hunter.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.hunter.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
		s.hunter.Time_firstButton = "5"
	case ID_Marksman:
		s.marksman.LoginInfo.IsUploadInfo = true
		s.marksman.Time_start = currentTime()
		s.marksman.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.marksman.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.marksman.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.marksman.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
		s.marksman.Point_right = "10"
		s.marksman.Point_left = "20"
	case ID_Miner:
		s.miner.LoginInfo.IsUploadInfo = true
		s.miner.Time_start = currentTime()
		s.miner.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.miner.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.miner.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.miner.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
	case ID_Privity:
		s.privity.LoginInfo.IsUploadInfo = true
		s.privity.Time_start = currentTime()
		s.privity.LoginInfo.PlayerCardInfo["1p"] = "cardId1"
		s.privity.LoginInfo.CardTicketInfo["cardId1"] = "ticketId1"
		s.privity.LoginInfo.PlayerCardInfo["2p"] = "cardId2"
		s.privity.LoginInfo.CardTicketInfo["cardId2"] = "ticketId2"
		s.privity.Num_right = "10"
		s.privity.Num_question = "20"
	}

	switch gameId {
	case ID_Russian:
		s.russian.Time_end = currentTime()
	case ID_Adivainacion:
		s.adivainacion.Time_end = currentTime()
	case ID_Bang:
		s.bang.Time_end = currentTime()
	case ID_Follow:
		s.follow.Time_end = currentTime()
	case ID_Greeting:
		s.greeting.Time_end = currentTime()
	case ID_Highnoon:
		s.highnoon.Time_end = currentTime()
	case ID_Hunter:
		s.hunter.Time_end = currentTime()
		//choose box
	case ID_Marksman:
		s.marksman.Time_end = currentTime()
	case ID_Miner:
		s.miner.Time_end = currentTime()
	case ID_Privity:
		s.privity.Time_end = currentTime()
	}
	s.uploadGameInfo(msg, gameId)
}

func (s *Srv) resetGame(gameId int) {
	switch gameId {
	case ID_Russian:
		s.russian.Reset()
	case ID_Adivainacion:
		s.adivainacion.Reset()
	case ID_Bang:
		s.bang.Reset()
	case ID_Follow:
		s.follow.Reset()
	case ID_Greeting:
		s.greeting.Reset()
	case ID_Highnoon:
		s.highnoon.Reset()
	case ID_Hunter:
		s.hunter.Reset()
	case ID_Marksman:
		s.marksman.Reset()
	case ID_Miner:
		s.miner.Reset()
	case ID_Privity:
		s.privity.Reset()
	}
}

func (s *Srv) upldateGameInfo(msg *InboxMessage, gameId int) {
	switch gameId {
	case ID_Russian:
	case ID_Adivainacion:
	case ID_Bang:
		s.bang.Point_round[1] = msg.GetStr("PR1")
		s.bang.Point_round[2] = msg.GetStr("PR2")
		s.bang.Point_round[3] = msg.GetStr("PR3")
	case ID_Follow:
		s.follow.Last_round = msg.GetStr("LR")
	case ID_Greeting:
	case ID_Highnoon:
		s.highnoon.Result_round_1p[1] = msg.GetStr("R1P1")
		s.highnoon.Result_round_2p[1] = msg.GetStr("R1P2")
		s.highnoon.Result_round_1p[2] = msg.GetStr("R2P1")
		s.highnoon.Result_round_2p[2] = msg.GetStr("R2P2")
		s.highnoon.Result_round_1p[3] = msg.GetStr("R3P1")
		s.highnoon.Result_round_2p[3] = msg.GetStr("R3P2")
		s.highnoon.Result_round_1p[4] = msg.GetStr("R4P1")
		s.highnoon.Result_round_2p[4] = msg.GetStr("R4P2")
		s.highnoon.Result_round_1p[5] = msg.GetStr("R5P1")
		s.highnoon.Result_round_2p[5] = msg.GetStr("R5P2")
		s.highnoon.Result_round_1p[6] = msg.GetStr("R6P1")
		s.highnoon.Result_round_2p[6] = msg.GetStr("R6P2")
		s.highnoon.Result_round_1p[7] = msg.GetStr("R7P1")
		s.highnoon.Result_round_2p[7] = msg.GetStr("R7P2")
	case ID_Hunter:
		s.hunter.Time_firstButton = msg.GetStr("FB")
	case ID_Marksman:
		s.marksman.Point_right = msg.GetStr("PR")
		s.marksman.Point_left = msg.GetStr("PL")
	case ID_Miner:
	case ID_Privity:
		s.privity.Num_right = msg.GetStr("NR")
		s.privity.Num_question = msg.GetStr("NQ")
	}
}

func (s *Srv) uploadGameInfo(msg *InboxMessage, gameId int) {
	switch gameId {
	case ID_Russian:
		if !s.russian.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Adivainacion:
		if !s.adivainacion.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Bang:
		if !s.bang.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Follow:
		if !s.follow.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Greeting:
		if !s.greeting.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Highnoon:
		if !s.highnoon.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Hunter:
		if !s.hunter.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Marksman:
		if !s.marksman.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Miner:
		if !s.miner.LoginInfo.IsUploadInfo {
			return
		}
	case ID_Privity:
		if !s.privity.LoginInfo.IsUploadInfo {
			return
		}
	}
	//arduinoId := msg.GetStr("ID")
	params := make(map[string]string)
	request := NewHttpRequest(s)
	switch gameId {
	case ID_Russian:
		request.SetApi(GameDataRussianCreate)
		params["card_ID1"] = s.russian.LoginInfo.PlayerCardInfo["1p"]
		params["card_ID2"] = s.russian.LoginInfo.PlayerCardInfo["2p"]
		params["time_start"] = s.russian.Time_start
		params["time_end"] = s.russian.Time_end
		params["op"] = "set_russian"
	case ID_Adivainacion:
		request.SetApi(GameDataAdivinacionCreate)
		params["card_ID"] = s.adivainacion.LoginInfo.PlayerCardInfo["1p"]
		params["time_start"] = s.adivainacion.Time_start
		params["time_end"] = s.adivainacion.Time_end
		params["op"] = "set_adivinacion"
	case ID_Bang:
		request.SetApi(GameDataBangCreate)
		params["card_ID"] = s.bang.LoginInfo.PlayerCardInfo["1p"]
		params["time_start"] = s.bang.Time_start
		params["time_end"] = s.bang.Time_end
		params["point_round1"] = s.bang.Point_round[1]
		params["point_round2"] = s.bang.Point_round[2]
		params["point_round3"] = s.bang.Point_round[3]
		params["op"] = "set_bang"
	case ID_Follow:
		request.SetApi(GameDataFollowCreate)
		params["card_ID1"] = s.follow.LoginInfo.PlayerCardInfo["1p"]
		params["card_ID2"] = s.follow.LoginInfo.PlayerCardInfo["2p"]
		params["time_start"] = s.follow.Time_start
		params["time_end"] = s.follow.Time_end
		params["last_round"] = s.follow.Last_round
		params["op"] = "set_follow"
	case ID_Greeting:
		request.SetApi(GameDataGreetingCreate)
		params["card_ID1"] = s.greeting.LoginInfo.PlayerCardInfo["1p"]
		params["card_ID2"] = s.greeting.LoginInfo.PlayerCardInfo["2p"]
		params["time_start"] = s.greeting.Time_start
		params["time_end"] = s.greeting.Time_end
		params["op"] = "set_greeting"
	case ID_Highnoon:
		request.SetApi(GameDataHighnoonCreate)
		params["card_ID1"] = s.highnoon.LoginInfo.PlayerCardInfo["1p"]
		params["card_ID2"] = s.highnoon.LoginInfo.PlayerCardInfo["2p"]
		params["time_start"] = s.highnoon.Time_start
		params["time_end"] = s.highnoon.Time_end
		params["1p_result_round1"] = s.highnoon.Result_round_1p[1]
		params["2p_result_round1"] = s.highnoon.Result_round_2p[1]
		params["1p_result_round2"] = s.highnoon.Result_round_1p[2]
		params["2p_result_round2"] = s.highnoon.Result_round_2p[2]
		params["1p_result_round3"] = s.highnoon.Result_round_1p[3]
		params["2p_result_round3"] = s.highnoon.Result_round_2p[3]
		params["1p_result_round4"] = s.highnoon.Result_round_1p[4]
		params["2p_result_round4"] = s.highnoon.Result_round_2p[4]
		params["1p_result_round5"] = s.highnoon.Result_round_1p[5]
		params["2p_result_round5"] = s.highnoon.Result_round_2p[5]
		params["1p_result_round6"] = s.highnoon.Result_round_1p[6]
		params["2p_result_round6"] = s.highnoon.Result_round_2p[6]
		params["1p_result_round7"] = s.highnoon.Result_round_1p[7]
		params["2p_result_round7"] = s.highnoon.Result_round_2p[7]
		params["op"] = "set_highnoon"
	case ID_Hunter:
		request.SetApi(GameDataHunterCreate)
		params["card_ID1"] = s.hunter.LoginInfo.PlayerCardInfo["1p"]
		params["card_ID2"] = s.hunter.LoginInfo.PlayerCardInfo["2p"]
		params["time_start"] = s.hunter.Time_start
		params["time_end"] = s.hunter.Time_end
		params["time_firstbutton"] = s.hunter.Time_firstButton
		params["box_ID"] = strconv.Itoa(s.hunter.Box_ID)
		params["op"] = "set_hunter"
	case ID_Marksman:
		request.SetApi(GameDataMarksmanCreate)
		params["card_ID1"] = s.marksman.LoginInfo.PlayerCardInfo["1p"]
		params["card_ID2"] = s.marksman.LoginInfo.PlayerCardInfo["2p"]
		params["time_start"] = s.marksman.Time_start
		params["time_end"] = s.marksman.Time_end
		params["point_left"] = s.marksman.Point_left
		params["point_right"] = s.marksman.Point_right
		params["op"] = "set_marksman"
	case ID_Miner:
		request.SetApi(GameDataMinerCreate)
		params["card_ID1"] = s.miner.LoginInfo.PlayerCardInfo["1p"]
		params["card_ID2"] = s.miner.LoginInfo.PlayerCardInfo["2p"]
		params["time_start"] = s.miner.Time_start
		params["time_end"] = s.miner.Time_end
		params["op"] = "set_miner"
	case ID_Privity:
		request.SetApi(GameDataPrivityCreate)
		params["card_ID1"] = s.privity.LoginInfo.PlayerCardInfo["1p"]
		params["card_ID2"] = s.privity.LoginInfo.PlayerCardInfo["2p"]
		params["time_start"] = s.privity.Time_start
		params["time_end"] = s.privity.Time_end
		params["number_question"] = s.privity.Num_question
		params["number_right"] = s.privity.Num_right
		params["op"] = "set_privity"
	}
	//request.SetArduinoId(arduinoId)
	request.SetMsg(msg)
	request.SetParams(params)
	request.DoPost()
}

func currentTime() string {
	tm := time.Now().Format("2006-01-02 15:04:05")
	return tm
}
