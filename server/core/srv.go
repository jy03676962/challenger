package core

import (
	"log"
	"net"
	"os"

	"golang.org/x/net/websocket"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"time"
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
	Event            = "10"
	DJControl        = "11"
	MineControl      = "12"
	BoxStatusGet     = "13"
)

var _ = log.Println

type Srv struct {
	inbox            *Inbox
	inboxMessageChan chan *InboxMessage
	mChan            chan MatchEvent
	httpResChan      chan *HttpResponse
	aDict            map[string]*ArduinoController
	match            *Match
	isSimulator      bool
	//--------game info------------
	boxes        []HunterBox
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
	s.initArduinoControllers()
	s.initGameInfo()
	return &s
}

func (s *Srv) Run(tcpAddr string, adminAddr string, dbPath string) {
	go s.listenTcp(tcpAddr)
	go s.listenTcp(adminAddr)
	go s.watchBoxStatus()
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
	log.Println("data server res:", httpRes.JsonData)
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
		if res, ok := httpRes.Get("return").(bool); ok {
			gameId, _ := strconv.Atoi(httpRes.Msg.GetStr("GAME"))
			if !res {
				//s.uploadGameInfo(httpRes.Msg, gameId)
				log.Println("upload game ", gameId, " failed!")
			} else {
				s.resetGame(gameId)
			}
		}
	case AuthorityGet:
		if res, ok := httpRes.Get("return").(bool); ok {
			arduinoId := httpRes.Msg.GetStr("ID")
			arduinoType := at(arduinoId)
			addr := InboxAddress{arduinoType, arduinoId}
			msg := NewInboxMessage()
			msg.SetCmd("authority_check")
			if res {
				msg.Set("return", "true")
			} else {
				msg.Set("return", "false")
			}
			s.sendToOne(msg, addr)
		}
	case TicketUse:
		if res, ok := httpRes.Get("return").(bool); ok {
			gameId, _ := strconv.Atoi(httpRes.Msg.GetStr("GAME"))
			if !res {
				log.Println("Modify Ticket failed!Game ", gameId, "start failed!")
				//s.gameStart(gameId, httpRes.Msg)
			}
		}
	case TicketCheck:
		arduinoId := httpRes.Msg.GetStr("ID")
		gameId, _ := strconv.Atoi(httpRes.Msg.GetStr("GAME"))
		addr := InboxAddress{InboxAddressTypeGameArduinoDevice, arduinoId}
		msg := NewInboxMessage()
		msg.SetCmd("ticket_check")
		if ticketId, ok := httpRes.Get("id").(float64); ok {
			if ticketId != -1 {
				s.loginGame(strconv.FormatFloat(ticketId, 'f', 0, 64), gameId, httpRes.Msg)
				msg.Set("return", "true")
				log.Println("it has ticket:", ticketId, " gameId:", gameId)
			} else {
				msg.Set("return", "false")
				log.Println("it has'n ticket!")
			}
		} else {
			log.Println("ticketId is'n int!", reflect.TypeOf(ticketId), "ticketId:", ticketId)
		}
		res := httpRes.Data
		log.Println("arduinoId:", arduinoId, "need feedback! and res :", res)
		s.sendToOne(msg, addr)
	case BoxUpload:
		if res, ok := httpRes.Get("return").(bool); ok {
			if !res {
				log.Println("Modify BoxStatus failed!")
				//s.uploadBoxStatus(boxId)
			} else {
				log.Println("box has been upload!")
			}
		}
	}
}

func (s *Srv) handleMatchEvent(evt MatchEvent) {
	switch evt.Type {
	//case MatchEventTypeEnd:
	//case MatchEventTypeUpdate:
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
	case InboxAddressTypeBoxArduinoDevice:
		s.handleArduinoMessage(msg)
	case InboxAddressTypeNightArduino:
		s.handleArduinoMessage(msg)
	case InboxAddressTypeDjArduino:
		s.handleArduinoMessage(msg)
	}
}

func (s *Srv) handleArduinoMessage(msg *InboxMessage) {
	cmd := msg.GetCmd()
	switch cmd {
	case UnKnown:
		log.Println(msg.GetStr("ID"), "send UnKnown cmd!")
	case Hbt:
		//log.Println("Receive htb:", msg.GetStr("ID"),msg.GetStr("CARD_ID"))
	case GameStartForward:
		admin := msg.GetStr("ADMIN")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		arduino := msg.GetStr("ARDUINO")
		log.Println("Game:", gameId, "start and forward to ", arduino, "! operator:", admin)
		s.gameControl("1", arduino)
		s.gameStart(gameId, msg)
	case GameStart:
		admin := msg.GetStr("ADMIN")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		log.Println("Game:", gameId, "start! operator:", admin)
		s.gameStart(gameId, msg)
	case GameEndForward:
		admin := msg.GetStr("ADMIN")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		arduino := msg.GetStr("ARDUINO")
		log.Println("Game:", gameId, "end and forward to ", arduino, "! operator:", admin)
		//不处理数据，只进行转发
		s.gameControl("0", arduino)
	case GameEnd:
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		s.gameEnd(msg, gameId)
		log.Println("Game:", gameId, "end!")
	case GameData:
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		log.Println("Receive Game:", gameId, "'s data!")
		s.updateGameInfo(msg, gameId)
	case AuthorityCheck:
		arduinoId := msg.GetStr("ID") //创建request的时候需要放入
		cardId := msg.GetStr("CARD_ID")
		authorityId := msg.GetStr("AR")
		log.Println("Get the card:", cardId, "  AuthrorityId:", authorityId, " ArduinoId:", arduinoId)
		request := NewHttpRequest(s)
		request.SetApi(AuthorityGet)
		params := make(map[string]string)
		params["card_Uid"] = cardId
		params["authority_ID"] = authorityId
		params["op"] = "validate_authid"
		request.SetParams(params)
		request.SetMsg(msg)
		request.DoGet()
	case TicketGet:
		admin := msg.GetStr("ADMIN")
		gameId := msg.GetStr("GAME")
		cardId := msg.GetStr("CARD_ID")
		log.Println("Ticket Get：  CardId:", cardId, "GameId:", gameId, " Admin:", admin)
		request := NewHttpRequest(s)
		request.SetApi(TicketCheck)
		params := make(map[string]string)
		params["card_Uid"] = cardId
		params["game_ID"] = gameId
		params["op"] = "get_ticket_game_id"
		request.SetParams(params)
		request.SetMsg(msg)
		request.DoGet()
	case BoxStatus:
		boxId, _ := strconv.Atoi(msg.GetStr("BOX_ID"))
		boxStatus, _ := strconv.Atoi(msg.GetStr("ST"))
		for k := range s.boxes {
			if s.boxes[k].Box_ID == boxId {
				s.boxes[k].Box_status = boxStatus
				switch boxStatus {
				case 0:
					s.uploadBoxStatus(k)
					s.boxes[k].Reset()
					log.Println("Box:", boxId, "has'n been opened by player，and reset the box！")
				case 1:
					s.uploadBoxStatus(k)
					log.Println("Box:", boxId, "has been opened by player! Watting reset!")
				case 2:
					s.boxes[k].Reset()
					log.Println("Box:", boxId, "has been reset by admin!")
				}
				break
			}
		}
	case ResetGame:
		admin := msg.GetStr("ADMIN")
		gameId, _ := strconv.Atoi(msg.GetStr("GAME"))
		s.resetGame(gameId)
		log.Println("Admin:", admin, " reset the game:", gameId, "!")
	case Event:
		event, _ := strconv.Atoi(msg.GetStr("EVENT"))
		s.startNewMatch(event)
	case DJControl:
		dj, _ := strconv.Atoi(msg.GetStr("DJ"))
		s.startNewMatch(dj)
		log.Println("DJ:", dj)
	case MineControl:
		arduinoId := msg.GetStr("ID")
		mineNum := msg.GetStr("M")
		control := msg.GetStr("CTRL")

		addr := InboxAddress{InboxAddressTypeGameArduinoDevice, arduinoId}
		msg := NewInboxMessage()
		msg.SetCmd("mine_ctrl")
		msg.Set("num", mineNum)
		msg.Set("ctrl", control)
		s.sendToOne(msg, addr)
	case BoxStatusGet:
		arduinoId := msg.GetStr("ID")
		box := make([]map[string]string, 0)
		for i := range s.boxes {
			var status int
			if s.boxes[i].IsAssigned {
				if s.boxes[i].Box_status == 1 {
					status = 2
				} else {
					status = 1
				}
			} else {
				status = 0
			}
			box = append(box,
				map[string]string{"box_n": strconv.Itoa(i), "box_s": strconv.Itoa(status)},
			)
		}
		addr := InboxAddress{InboxAddressTypeGameArduinoDevice, arduinoId}
		msg := NewInboxMessage()
		msg.SetCmd("box_status")
		msg.Set("box", box)
		s.sendToOne(msg, addr)
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

func (s *Srv) startNewMatch(event int) {
	if s.match != nil {
		if s.match.IsGoing {
			return
		} else if event != s.match.Event {
			s.stopMatch()
			m := NewMatch(s, event)
			s.match = m
			go m.Run()
		}
	} else {
		m := NewMatch(s, event)
		s.match = m
		go m.Run()
	}
}

func (s *Srv) stopMatch() {
	s.match.Stop()
	s.match = nil
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

func (s *Srv) gameControl(value string, arduinoId string) {
	addr := InboxAddress{InboxAddressTypeGameArduinoDevice, arduinoId}
	msg := NewInboxMessage()
	msg.SetCmd("game_ctrl")
	msg.Set("value", value)
	s.sendToOne(msg, addr)
}

func (s *Srv) doorControl(IL string, OL string, ID string) {
	addr := InboxAddress{InboxAddressTypeDjArduino, ID}
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
	for _, gameArdunio := range GetOptions().GameArduino {
		addr := InboxAddress{InboxAddressTypeGameArduinoDevice, gameArdunio}
		controller := NewArduinoController(addr)
		s.aDict[addr.String()] = controller
	}
	for _, boxArduino := range GetOptions().BoxArduino {
		addr := InboxAddress{InboxAddressTypeBoxArduinoDevice, boxArduino}
		controller := NewArduinoController(addr)
		s.aDict[addr.String()] = controller
	}
	for _, trashArduino := range GetOptions().NightArduino {
		addr := InboxAddress{InboxAddressTypeNightArduino, trashArduino}
		controller := NewArduinoController(addr)
		s.aDict[addr.String()] = controller
	}
	for _, djArduino := range GetOptions().DjArduino {
		addr := InboxAddress{InboxAddressTypeDjArduino, djArduino}
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
	s.boxes = make([]HunterBox, GetOptions().BoxNum)
	for i := range s.boxes {
		s.boxes[i].Box_ID = i
	}
}

func (s *Srv) bgmControl(music string) {
	msg := NewInboxMessage()
	msg.SetCmd("mp3_ctrl")
	msg.Set("music", music)
	s.sends(msg, InboxAddressTypeDjArduino)
}

func (s *Srv) loginGame(ticketId string, gameId int, msg *InboxMessage) {
	//log.Println("login:", msg)
	cardId := msg.GetStr("CARD_ID")
	switch gameId {
	case ID_Russian:
		s.russian.LoginInfo.setCardId(cardId)
		s.russian.LoginInfo.setTicket(cardId, ticketId)
	case ID_Adivainacion:
		s.adivainacion.LoginInfo.setCardId(cardId)
		s.adivainacion.LoginInfo.setTicket(cardId, ticketId)
	case ID_Bang:
		s.bang.LoginInfo.setCardId(cardId)
		s.bang.LoginInfo.setTicket(cardId, ticketId)
	case ID_Follow:
		s.follow.LoginInfo.setCardId(cardId)
		s.follow.LoginInfo.setTicket(cardId, ticketId)
	case ID_Greeting:
		s.greeting.LoginInfo.setCardId(cardId)
		s.greeting.LoginInfo.setTicket(cardId, ticketId)
	case ID_Highnoon:
		s.highnoon.LoginInfo.setCardId(cardId)
		s.highnoon.LoginInfo.setTicket(cardId, ticketId)
	case ID_Hunter:
		s.hunter.LoginInfo.setCardId(cardId)
		s.hunter.LoginInfo.setTicket(cardId, ticketId)
	case ID_Marksman:
		s.marksman.LoginInfo.setCardId(cardId)
		s.marksman.LoginInfo.setTicket(cardId, ticketId)
	case ID_Miner:
		s.miner.LoginInfo.setCardId(cardId)
		s.miner.LoginInfo.setTicket(cardId, ticketId)
	case ID_Privity:
		s.privity.LoginInfo.setCardId(cardId)
		s.privity.LoginInfo.setTicket(cardId, ticketId)
	}
}

func (s *Srv) gameStart(gameId int, msg *InboxMessage) {
	admin := msg.GetStr("ADMIN")
	request := NewHttpRequest(s)
	request.SetMsg(msg)
	request.SetApi(TicketUse)
	params := make(map[string]string)
	params["op"] = "set_exchanger_id"
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.russian.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.adivainacion.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.bang.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.follow.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.greeting.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.highnoon.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.hunter.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.marksman.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.miner.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
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
			request2 := NewHttpRequest(s)
			request2.SetMsg(msg)
			request2.SetApi(TicketUse)
			params2 := make(map[string]string)
			params2["op"] = "set_exchanger_id"
			params2["game_ID"] = strconv.Itoa(gameId)
			params2["exchanger_ID"] = admin
			params2["id"] = s.privity.LoginInfo.CardTicketInfo[cardId_2p]
			request2.SetParams(params2)
			request2.DoPost()
		}
	}
}

func (s *Srv) gameEnd(msg *InboxMessage, gameId int) {
	//s.hunter.LoginInfo.IsUploadInfo = true
	//s.hunter.Time_start = currentTime()
	//s.hunter.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
	//s.hunter.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
	//s.hunter.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
	//s.hunter.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
	//s.hunter.Time_firstButton = "5"
	//s.hunter.Box_ID = 1
	s.updateGameInfo(msg, gameId)
	/*            test code
		switch gameId {
		case ID_Russian:
			s.russian.LoginInfo.IsUploadInfo = true
			s.russian.Time_start = currentTime()
			//s.russian.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.russian.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.russian.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.russian.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
			//s.russian.Bullet_trigger = "3"
			//s.russian.Desk_num = "5"
		case ID_Adivainacion:
			s.adivainacion.LoginInfo.IsUploadInfo = true
			s.adivainacion.Time_start = currentTime()
			//s.adivainacion.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.adivainacion.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.adivainacion.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.adivainacion.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
		case ID_Bang:
			s.bang.LoginInfo.IsUploadInfo = true
			s.bang.Time_start = currentTime()
			//s.bang.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.bang.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.bang.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.bang.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
			//s.bang.Point_round[1] = "6"
			//s.bang.Point_round[2] = "5"
			//s.bang.Point_round[3] = "4"
		case ID_Follow:
			s.follow.LoginInfo.IsUploadInfo = true
			s.follow.Time_start = currentTime()
			//s.follow.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.follow.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.follow.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.follow.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
			//s.follow.Last_round = "10"
		case ID_Greeting:
			s.greeting.LoginInfo.IsUploadInfo = true
			s.greeting.Time_start = currentTime()
			//s.greeting.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.greeting.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.greeting.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.greeting.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
		case ID_Highnoon:
			s.highnoon.LoginInfo.IsUploadInfo = true
			s.highnoon.Time_start = currentTime()
			//s.highnoon.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.highnoon.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.highnoon.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.highnoon.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
			//s.highnoon.Result_round_1p[1] = "0.11"
			//s.highnoon.Result_round_2p[1] = "0.12"
			//s.highnoon.Result_round_1p[2] = "0.21"
			//s.highnoon.Result_round_2p[2] = "0.22"
			//s.highnoon.Result_round_1p[3] = "0.31"
			//s.highnoon.Result_round_2p[3] = "0.32"
			//s.highnoon.Result_round_1p[4] = "0.41"
			//s.highnoon.Result_round_2p[4] = "0.42"
			//s.highnoon.Result_round_1p[5] = "0.51"
			//s.highnoon.Result_round_2p[5] = "0.52"
			//s.highnoon.Result_round_1p[6] = "0.61"
			//s.highnoon.Result_round_2p[6] = "0.62"
			//log.Println(s.highnoon.Result_round_1p)
		case ID_Hunter:
			s.hunter.LoginInfo.IsUploadInfo = true
			s.hunter.Time_start = currentTime()
			s.hunter.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			s.hunter.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			s.hunter.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			s.hunter.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
			s.hunter.Time_firstButton = "5"
			s.hunter.Box_ID = 1
		case ID_Marksman:
			s.marksman.LoginInfo.IsUploadInfo = true
			s.marksman.Time_start = currentTime()
			//s.marksman.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.marksman.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.marksman.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.marksman.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
			//s.marksman.Point_right = "10"
			//s.marksman.Point_left = "20"
		case ID_Miner:
			s.miner.LoginInfo.IsUploadInfo = true
			s.miner.Time_start = currentTime()
			//s.miner.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.miner.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.miner.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.miner.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
		case ID_Privity:
			s.privity.LoginInfo.IsUploadInfo = true
			s.privity.Time_start = currentTime()
			//s.privity.LoginInfo.PlayerCardInfo["1p"] = "00FF0FF000FFCF4D54B110484DBDBBB104D0"
			//s.privity.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B110484DBDBBB104D0"] = "ticketId1"
			//s.privity.LoginInfo.PlayerCardInfo["2p"] = "00FF0FF000FFCF4D54B1104846B4FBC10480"
			//s.privity.LoginInfo.CardTicketInfo["00FF0FF000FFCF4D54B1104846B4FBC10480"] = "ticketId2"
			//s.privity.Num_right = "10"
			//s.privity.Num_question = "20"
		}
	                test      code                */
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

func (s *Srv) updateGameInfo(msg *InboxMessage, gameId int) {
	switch gameId {
	case ID_Russian:
		if msg.GetStr("BT") != "" {
			s.russian.Bullet_trigger = msg.GetStr("BT")
		} else {
			s.russian.Bullet_trigger = "0"
		}

		if msg.GetStr("DN") != "" {
			s.russian.Desk_num = msg.GetStr("DN")
		} else {
			s.russian.Desk_num = "0"
		}
	case ID_Adivainacion:
	case ID_Bang:
		if msg.GetStr("PR1") != "" {
			s.bang.Point_round[1] = msg.GetStr("PR1")
		} else {
			s.bang.Point_round[1] = "0"
		}

		if msg.GetStr("PR2") != "" {
			s.bang.Point_round[2] = msg.GetStr("PR2")
		} else {
			s.bang.Point_round[2] = "0"
		}
		if msg.GetStr("PR3") != "" {
			s.bang.Point_round[3] = msg.GetStr("PR3")
		} else {
			s.bang.Point_round[3] = "0"
		}
	case ID_Follow:
		if msg.GetStr("LR") != "" {
			s.follow.Last_round = msg.GetStr("LR")
		} else {
			s.follow.Last_round = "0"
		}
	case ID_Greeting:
	case ID_Highnoon:
		if msg.GetStr("R1P1") != "" {
			s.highnoon.Result_round_1p[1] = msg.GetStr("R1P1")
		} else {
			s.highnoon.Result_round_1p[1] = "0"
		}

		if msg.GetStr("R1P2") != "" {
			s.highnoon.Result_round_2p[1] = msg.GetStr("R1P2")
		} else {
			s.highnoon.Result_round_2p[1] = "0"
		}

		if msg.GetStr("R2P1") != "" {
			s.highnoon.Result_round_1p[2] = msg.GetStr("R2P1")
		} else {
			s.highnoon.Result_round_1p[2] = "0"
		}

		if msg.GetStr("R2P2") != "" {
			s.highnoon.Result_round_2p[2] = msg.GetStr("R2P2")
		} else {
			s.highnoon.Result_round_2p[2] = "0"
		}

		if msg.GetStr("R3P1") != "" {
			s.highnoon.Result_round_1p[3] = msg.GetStr("R3P1")
		} else {
			s.highnoon.Result_round_1p[3] = "0"
		}

		if msg.GetStr("R3P2") != "" {
			s.highnoon.Result_round_2p[3] = msg.GetStr("R3P2")
		} else {
			s.highnoon.Result_round_2p[3] = "0"
		}

		if msg.GetStr("R4P1") != "" {
			s.highnoon.Result_round_1p[4] = msg.GetStr("R4P1")
		} else {
			s.highnoon.Result_round_1p[4] = "0"
		}

		if msg.GetStr("R4P2") != "" {
			s.highnoon.Result_round_2p[4] = msg.GetStr("R4P2")
		} else {
			s.highnoon.Result_round_2p[4] = "0"
		}

		if msg.GetStr("R5P1") != "" {
			s.highnoon.Result_round_1p[5] = msg.GetStr("R5P1")
		} else {
			s.highnoon.Result_round_1p[5] = "0"
		}

		if msg.GetStr("R5P2") != "" {
			s.highnoon.Result_round_2p[5] = msg.GetStr("R5P2")
		} else {
			s.highnoon.Result_round_2p[5] = "0"
		}

		if msg.GetStr("R6P1") != "" {
			s.highnoon.Result_round_1p[6] = msg.GetStr("R6P1")
		} else {
			s.highnoon.Result_round_1p[6] = "0"
		}

		if msg.GetStr("R6P2") != "" {
			s.highnoon.Result_round_2p[6] = msg.GetStr("R6P2")
		} else {
			s.highnoon.Result_round_2p[6] = "0"
		}

		if msg.GetStr("R7P1") != "" {
			s.highnoon.Result_round_1p[7] = msg.GetStr("R7P1")
		} else {
			s.highnoon.Result_round_1p[7] = "0"
		}

		if msg.GetStr("R7P2") != "" {
			s.highnoon.Result_round_2p[7] = msg.GetStr("R7P2")
		} else {
			s.highnoon.Result_round_2p[7] = "0"
		}
	case ID_Hunter:
		s.hunter.Time_firstButton = msg.GetStr("FB")
		if s.hunter.Time_firstButton != "0" && s.hunter.Time_firstButton != "" {
			//choose box
			cardId1 := s.hunter.LoginInfo.PlayerCardInfo["1p"]
			cardId2 := s.hunter.LoginInfo.PlayerCardInfo["2p"]
			s.hunter.Box_ID = s.boxes[s.setBox(cardId1, cardId2)].Box_ID
			if s.hunter.Box_ID != -1 {
				arduinoId := returnBox(s.hunter.Box_ID)
				if arduinoId != "" {
					addr := InboxAddress{InboxAddressTypeBoxArduinoDevice, arduinoId}
					msg := NewInboxMessage()
					msg.SetCmd("box_set")
					msg.Set("cardId1", cardId1)
					if cardId2 != "" {
						msg.Set("cardId2", cardId2)
					} else {
						log.Println("none cardId2")
					}
					s.sendToOne(msg, addr)
				}
				log.Println(s.boxes)
			}
			log.Println("assigned box ~ cardId1:", cardId1, " cardId2:", cardId2)
		} else {
			s.hunter.Time_firstButton = "0"
		}
	case ID_Marksman:
		if msg.GetStr("PR") != "" {
			s.marksman.Point_right = msg.GetStr("PR")
		} else {
			s.marksman.Point_right = "0"
		}

		if msg.GetStr("PL") != "" {
			s.marksman.Point_left = msg.GetStr("PL")
		} else {
			s.marksman.Point_left = "0"
		}
	case ID_Miner:
	case ID_Privity:
		if msg.GetStr("NR") != "" {
			s.privity.Num_right = msg.GetStr("NR")
		} else {
			s.privity.Num_right = "0"
		}

		if msg.GetStr("NQ") != "" {
			s.privity.Num_question = msg.GetStr("NQ")
		} else {
			s.privity.Num_question = "0"
		}
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
		params["desk_no"] = s.russian.Desk_num
		params["bullet_trigger"] = s.russian.Bullet_trigger
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

func (s *Srv) uploadBoxStatus(boxNum int) {
	params := make(map[string]string)
	params["box_ID"] = strconv.Itoa(s.boxes[boxNum].Box_ID)
	params["time_build"] = s.boxes[boxNum].Time_build
	params["time_validity"] = s.boxes[boxNum].Time_validity
	params["card_ID1"] = s.boxes[boxNum].Card_ID1
	params["card_ID2"] = s.boxes[boxNum].Card_ID2
	params["box_status"] = strconv.Itoa(s.boxes[boxNum].Box_status)
	params["op"] = "set_hunter_box"
	request := NewHttpRequest(s)
	request.SetApi(BoxUpload)
	request.SetParams(params)
	request.DoPost()
}

//根据num 返回boxId
func returnBox(boxNum int) string {
	switch boxNum {
	case 0:
		return "B-1"
	case 1:
		return "B-2"
	case 2:
		return "B-3"
	case 3:
		return "B-4"
	case 4:
		return "B-5"
	case 5:
		return "B-6"
	case 6:
		return "B-7"
	case 7:
		return "B-8"
	case 8:
		return "B-9"
	case 9:
		return "B-10"
	}
	return ""
}

func (s *Srv) setBox(cardId1, cardId2 string) int {
	boxId := s.getRandomBoxId()
	//for k := range s.boxes {
	//	if s.boxes[k].Box_ID == 0 && !s.boxes[k].IsAssigned {
	//		boxId = k
	//	}
	//}
	if boxId == -1 {
		log.Println("all box has been assigned!")
	} else {
		s.boxes[boxId].IsAssigned = true
		s.boxes[boxId].Box_status = -1
		s.boxes[boxId].Card_ID1 = cardId1
		s.boxes[boxId].Card_ID2 = cardId2
		s.boxes[boxId].Time_build = currentTime()
		s.boxes[boxId].Time_validity = boxLastTime()
		log.Println("BoxId:", s.boxes[boxId].Box_ID, " is assigned!")
		log.Println("BoxInfo:", s.boxes[boxId])
	}
	return boxId
}

func (s *Srv) getRandomBoxId() int {
	totalNum := s.getNotAssignedBoxTotalNum()
	if totalNum == 0 {
		return -1
	} else {
		return s.generateRandomNumber(totalNum)
	}
}

//获得没有被分配出去的宝箱数量
func (s *Srv) getNotAssignedBoxTotalNum() int {
	var totalNum int = 0
	for i := range s.boxes {
		if !s.boxes[i].IsAssigned {
			totalNum++
		}
	}
	log.Println(totalNum, " boxes is not assigned!")
	return totalNum
}

//从没有被分配出去的宝箱中挑选一个
func (s *Srv) generateRandomNumber(boxTotalNum int) int {
	//首先对宝箱进行牌序
	sort.Sort(HunterBoxSlice(s.boxes))
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//生成随机数
	num := r.Intn(boxTotalNum)
	return num
}

func (s *Srv) watchBoxStatus() {
	dt := 1000 * time.Millisecond
	tickChan := time.Tick(dt)
	for {
		<-tickChan
		for i := range s.boxes {
			if s.boxes[i].IsAssigned && s.boxes[i].Box_status == -1 {
				loc, _ := time.LoadLocation("Local")
				validityTime, err := time.ParseInLocation("2006-01-02 15:04:05", s.boxes[i].Time_validity, loc)
				//log.Println("validityTime:", validityTime)
				if err == nil {
					lastTime := validityTime.Unix()
					timeNow := time.Now().Unix()
					//log.Println("now:", timeNow, " endTime", lastTime)
					if lastTime <= timeNow {
						var arduinoId string
						arduinoId = returnBox(i)
						addr := InboxAddress{InboxAddressTypeBoxArduinoDevice, arduinoId}
						msg := NewInboxMessage()
						msg.SetCmd("box_reset")
						msg.Set("num", strconv.Itoa(i))
						s.sendToOne(msg, addr)
					}
				}
			}

		}
	}
}

func currentTime() string {
	tm := time.Now().Format("2006-01-02 15:04:05")
	return tm
}

func boxLastTime() string {
	nextTime := time.Now().Unix() + int64(GetOptions().BoxLastTime)
	tm := time.Unix(nextTime, 0).Format("2006-01-02 15:04:05")
	log.Println("UTC:", tm)
	return tm
}
