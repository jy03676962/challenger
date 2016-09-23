package core

import (
	"fmt"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type AdminMode int

const (
	AdminModeNormal = iota
	AdminModeDebug  = iota
)

var _ = log.Println

type pendingMatch struct {
	ids  []string
	mode string
}

type Srv struct {
	inbox            *Inbox
	queue            *Queue
	db               *DB
	inboxMessageChan chan *InboxMessage
	mChan            chan MatchEvent
	pDict            map[string]*PlayerController
	aDict            map[string]*ArduinoController
	mDict            map[uint]*Match
	adminMode        AdminMode
	isSimulator      bool
	qc               *QuickChecker
}

func NewSrv(isSimulator bool) *Srv {
	s := Srv{}
	s.isSimulator = isSimulator
	s.inbox = NewInbox(&s)
	s.queue = NewQueue(&s)
	s.db = NewDb()
	s.inboxMessageChan = make(chan *InboxMessage, 1)
	s.mChan = make(chan MatchEvent)
	s.pDict = make(map[string]*PlayerController)
	s.aDict = make(map[string]*ArduinoController)
	s.mDict = make(map[uint]*Match)
	s.adminMode = AdminModeNormal
	s.initArduinoControllers()
	return &s
}

func (s *Srv) Run(tcpAddr string, udpAddr string, dbPath string) {
	e := s.db.connect(dbPath)
	if e != nil {
		log.Printf("open database error:%v\n", e.Error())
		os.Exit(1)
	}
	go s.listenTcp(tcpAddr)
	go s.listenUdp(udpAddr)
	s.mainLoop()
}

func (s *Srv) ListenWebSocket(conn *websocket.Conn) {
	log.Println("got new ws connection")
	s.inbox.ListenConnection(NewInboxWsConnection(conn))
}

// http interface

func (s *Srv) AddTeam(c echo.Context) error {
	count, _ := strconv.Atoi(c.FormValue("count"))
	mode := c.FormValue("mode")
	id := s.queue.AddTeamToQueue(count, mode)
	d := map[string]interface{}{"id": id}
	return c.JSON(http.StatusOK, d)
}

func (s *Srv) ResetQueue(c echo.Context) error {
	id := s.queue.ResetQueue()
	d := map[string]interface{}{"id": id}
	return c.JSON(http.StatusOK, d)
}

func (s *Srv) GetHistory(c echo.Context) error {
	d := s.db.getHistory(12)
	return c.JSON(http.StatusOK, d)
}

func (s *Srv) MatchStartAnswer(c echo.Context) error {
	mid, _ := strconv.Atoi(c.FormValue("mid"))
	d := s.db.startAnswer(mid, c.FormValue("eid"))
	s.sendMsgs("startAnswer", *d, InboxAddressTypePostgameDevice)
	return c.JSON(http.StatusOK, d)
}

func (s *Srv) MatchStopAnswer(c echo.Context) error {
	mid, _ := strconv.Atoi(c.FormValue("mid"))
	s.db.stopAnswer(mid)
	s.sendMsgs("stopAnswer", nil, InboxAddressTypePostgameDevice)
	return c.JSON(http.StatusOK, mid)
}

func (s *Srv) GetSurvey(c echo.Context) error {
	return c.JSON(http.StatusOK, GetSurvey())
}

func (s *Srv) UpdateQuestionInfo(c echo.Context) error {
	pid, _ := strconv.Atoi(c.FormValue("pid"))
	p := s.db.updateQuestionInfo(pid, c.FormValue("qid"), c.FormValue("aid"))
	s.sendMsgs("updatePlayerData", *p, InboxAddressTypeAdminDevice)
	return c.JSON(http.StatusOK, nil)
}

func (s *Srv) UpdatePlayerData(c echo.Context) error {
	pid, _ := strconv.Atoi(c.FormValue("pid"))
	p := s.db.updatePlayerData(pid, c.FormValue("name"), c.FormValue("eid"))
	s.sendMsgs("updatePlayerData", *p, InboxAddressTypeAdminDevice, InboxAddressTypePostgameDevice)
	return c.JSON(http.StatusOK, nil)
}

func (s *Srv) UpdateMatchData(c echo.Context) error {
	mid, _ := strconv.Atoi(c.FormValue("mid"))
	s.db.updateMatchData(mid, c.FormValue("eid"))
	return c.JSON(http.StatusOK, nil)
}

func (s *Srv) GetMainArduinoList(c echo.Context) error {
	return c.JSON(http.StatusOK, GetOptions().MainArduinoInfo)
}

func (s *Srv) GetAnsweringMatchData(c echo.Context) error {
	d := s.db.getAnsweringMatchData()
	ret := make(map[string]interface{})
	if d == nil {
		ret["code"] = 1
	} else {
		ret["code"] = 0
		ret["data"] = d
	}
	return c.JSON(http.StatusOK, ret)
}

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

func (s *Srv) listenUdp(address string) {
	udpAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Println("resolve udp address error:", err.Error())
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", udpAddress)
	if err != nil {
		log.Println("udp listen error: ", err.Error())
		os.Exit(1)
	}
	log.Println("listen udp:", address)
	s.inbox.ListenConnection(NewInboxUdpConnection(conn))
}

func (s *Srv) onInboxMessageArrived(msg *InboxMessage) {
	s.inboxMessageChan <- msg
}

func (s *Srv) onMatchEvent(evt MatchEvent) {
	s.mChan <- evt
}

func (s *Srv) onQueueUpdated(queueData []Team) {
	s.sendMsgs("HallData", queueData, InboxAddressTypeAdminDevice)
	history := s.db.getHistory(3)
	msg := NewInboxMessage()
	msg.SetCmd("matchData")
	data := make(map[string]interface{})
	data["queue"] = queueData
	data["history"] = history
	msg.Set("data", data)
	s.sends(msg, InboxAddressTypeQueueDevice)
}

func (s *Srv) handleMatchEvent(evt MatchEvent) {
	switch evt.Type {
	case MatchEventTypeEnd:
		delete(s.mDict, evt.ID)
		for _, p := range s.pDict {
			if p.MatchID == evt.ID {
				p.MatchID = 0
			}
		}
		d := evt.Data.(map[string]interface{})
		d["matchID"] = evt.ID
		s.queue.TeamFinishMatch(d["teamID"].(string))
		s.db.saveOrDelMatchData(d["matchData"].(*MatchData))
		s.sendMsgs("matchStop", d, InboxAddressTypeSimulatorDevice, InboxAddressTypeAdminDevice, InboxAddressTypeIngameDevice, InboxAddressTypeQueueDevice)
	case MatchEventTypeUpdate:
		s.sendMsgs("updateMatch", evt.Data, InboxAddressTypeSimulatorDevice, InboxAddressTypeAdminDevice, InboxAddressTypeIngameDevice, InboxAddressTypeQueueDevice)
	}
}

func (s *Srv) handleInboxMessage(msg *InboxMessage) {
	shouldUpdatePlayerController := false
	if msg.RemoveAddress != nil && msg.RemoveAddress.Type.IsPlayerControllerType() {
		cid := msg.RemoveAddress.String()
		if pc, ok := s.pDict[cid]; ok {
			pc.Online = false
			if pc.MatchID > 0 {
				s.mDict[pc.MatchID].OnMatchCmdArrived(msg)
			}
			shouldUpdatePlayerController = true
		}
	}
	if msg.AddAddress != nil && msg.AddAddress.Type.IsPlayerControllerType() {
		cid := msg.AddAddress.String()
		if pc, ok := s.pDict[cid]; ok {
			pc.Online = true
			if pc.MatchID > 0 {
				s.mDict[pc.MatchID].OnMatchCmdArrived(msg)
			}
		} else {
			pc := NewPlayerController(*msg.AddAddress)
			s.pDict[pc.ID] = pc
			if msg.AddAddress.Type == InboxAddressTypeWearableDevice {
				s.wearableControl("01", pc.ID)
			}
		}
		shouldUpdatePlayerController = true
	}
	if shouldUpdatePlayerController {
		s.sendMsgs("ControllerData", s.getControllerData(), InboxAddressTypeAdminDevice, InboxAddressTypeSimulatorDevice)
	}

	if msg.RemoveAddress != nil && msg.RemoveAddress.Type.IsArduinoControllerType() {
		id := msg.RemoveAddress.String()
		if controller := s.aDict[id]; controller != nil {
			controller.Online = false
			controller.ScoreUpdated = false
		}
		s.sendMsgs("removeTCP", msg.RemoveAddress, InboxAddressTypeArduinoTestDevice)
	}

	if msg.AddAddress != nil && msg.AddAddress.Type.IsArduinoControllerType() {
		if controller := s.aDict[msg.AddAddress.String()]; controller != nil {
			controller.Online = true
			if controller.NeedUpdateScore() {
				s.updateArduinoControllerScore(controller)
			}
		} else {
			log.Printf("Warning: get arduino connection not belong to list:%v\n", msg.AddAddress.String())
		}
		s.sendMsgs("addTCP", msg.AddAddress, InboxAddressTypeArduinoTestDevice)
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
	case InboxAddressTypeSimulatorDevice:
		s.handleSimulatorMessage(msg)
	case InboxAddressTypeArduinoTestDevice:
		s.handleArduinoTestMessage(msg)
	case InboxAddressTypeAdminDevice:
		s.handleAdminMessage(msg)
	case InboxAddressTypeMainArduinoDevice, InboxAddressTypeSubArduinoDevice:
		s.handleArduinoMessage(msg)
	case InboxAddressTypePostgameDevice:
		s.handlePostGameMessage(msg)
	case InboxAddressTypeWearableDevice:
		s.handleWearableMessage(msg)
	case InboxAddressTypeIngameDevice:
		s.handleIngameMessage(msg)
	case InboxAddressTypeQueueDevice:
		s.handleQueueMessage(msg)
	}
}

func (s *Srv) handleQueueMessage(msg *InboxMessage) {
	cmd := msg.GetCmd()
	if cmd == "init" {
		s.queue.TeamQueryData()
		s.sendToOne(msg, *msg.Address)
	}
}

func (s *Srv) handleWearableMessage(msg *InboxMessage) {
	msg.SetCmd("wearableLoc")
	for _, m := range s.mDict {
		m.OnMatchCmdArrived(msg)
	}
}

func (s *Srv) handleIngameMessage(msg *InboxMessage) {
	cmd := msg.GetCmd()
	if cmd == "init" {
		s.sendToOne(msg, *msg.Address)
	}
}

func (s *Srv) handleArduinoMessage(msg *InboxMessage) {
	cmd := msg.GetCmd()
	switch cmd {
	case "confirm_init_score":
		if controller := s.aDict[msg.Address.String()]; controller != nil {
			controller.ScoreUpdated = true
		}
	case "upload_score":
		for _, m := range s.mDict {
			m.OnMatchCmdArrived(msg)
		}
	case "hb":
		mode := msg.GetStr("MD")
		//log.Printf("got heartbeat:%v\n", msg)
		var mm ArduinoMode
		if mode != "" {
			intMode, err := strconv.Atoi(mode)
			if err == nil {
				mm = ArduinoMode(intMode)
			} else {
				mm = ArduinoModeUnknown
			}
		} else {
			mm = ArduinoModeUnknown
		}
		if controller := s.aDict[msg.Address.String()]; controller != nil {
			controller.Mode = mm
		}
		if s.qc != nil {
			s.qc.OnArduinoHeartBeat(msg)
		}
		switch s.adminMode {
		case AdminModeNormal:
			for _, m := range s.mDict {
				m.OnLaserInfoArrived(msg)
			}
		case AdminModeDebug:
			ur := msg.GetStr("UR")
			count := 0
			idx := 0
			for i, r := range ur {
				c := string(r)
				if c == "1" {
					count += 1
					idx = i
				}
			}
			if count > 0 {
				m := NewInboxMessage()
				m.SetCmd("laserInfo")
				m.Set("id", msg.Address.ID)
				m.Set("ur", ur)
				m.Set("idx", idx)
				if count > 1 {
					m.Set("error", 2)
				} else {
					m.Set("error", 0)
				}
				s.sends(m, InboxAddressTypeAdminDevice)
			}
		}
	}
	if msg.GetCmd() != "init" {
		s.sends(msg, InboxAddressTypeArduinoTestDevice)
	}
}

func (s *Srv) handleSimulatorMessage(msg *InboxMessage) {
	cmd := msg.GetCmd()
	switch cmd {
	case "init":
		d := map[string]interface{}{
			"options": GetOptions(),
			"ID":      msg.Address.String(),
		}
		s.sendMsgToAddresses("init", d, []InboxAddress{*msg.Address})
	case "startMatch":
		mode := msg.GetStr("mode")
		ids := make([]string, 0)
		for _, pc := range s.pDict {
			if pc.Address.Type == InboxAddressTypeSimulatorDevice {
				ids = append(ids, pc.ID)
			}
		}
		s.startNewMatch(ids, mode, "")
	case "stopMatch", "playerMove", "playerStop":
		mid := uint(msg.Get("matchID").(float64))
		if match := s.mDict[mid]; match != nil {
			match.OnMatchCmdArrived(msg)
		}
	}
}

func (s *Srv) handleArduinoTestMessage(msg *InboxMessage) {
	log.Printf("got test message:%v\n", msg)
	destID := msg.GetStr("addr")
	if len(destID) > 0 {
		mainAddr := InboxAddress{InboxAddressTypeMainArduinoDevice, destID}
		subAddr := InboxAddress{InboxAddressTypeSubArduinoDevice, destID}
		doorAddr := InboxAddress{InboxAddressTypeDoorArduino, destID}
		musicAddr := InboxAddress{InboxAddressTypeMusicArduino, destID}
		s.send(msg, []InboxAddress{mainAddr, subAddr, doorAddr, musicAddr})
	} else {
		s.sends(msg, InboxAddressTypeSubArduinoDevice, InboxAddressTypeMainArduinoDevice, InboxAddressTypeArduinoTestDevice, InboxAddressTypeDoorArduino, InboxAddressTypeMusicArduino)
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
		s.sendMsg("init", nil, msg.Address.ID, msg.Address.Type)
	case "queryHallData":
		s.queue.TeamQueryData()
	case "queryControllerData":
		s.sendMsg("ControllerData", s.getControllerData(), msg.Address.ID, msg.Address.Type)
	case "queryQuestionCount":
		s.sendMsg("QuestionCount", len(GetSurvey().Questions), msg.Address.ID, msg.Address.Type)
	case "teamCutLine":
		teamID := msg.GetStr("teamID")
		s.queue.TeamCutLine(teamID)
	case "teamRemove":
		teamID := msg.GetStr("teamID")
		s.queue.TeamRemove(teamID)
	case "teamChangeMode":
		teamID := msg.GetStr("teamID")
		mode := msg.GetStr("mode")
		s.queue.TeamChangeMode(teamID, mode)
	case "teamDelay":
		teamID := msg.GetStr("teamID")
		s.queue.TeamDelay(teamID)
	case "teamAddPlayer":
		teamID := msg.GetStr("teamID")
		s.queue.TeamAddPlayer(teamID)
	case "teamRemovePlayer":
		teamID := msg.GetStr("teamID")
		s.queue.TeamRemovePlayer(teamID)
	case "teamPrepare":
		teamID := msg.GetStr("teamID")
		s.queue.TeamPrepare(teamID)
	case "teamCancelPrepare":
		teamID := msg.GetStr("teamID")
		s.queue.TeamCancelPrepare(teamID)
	case "teamStart":
		teamID := msg.GetStr("teamID")
		mode := msg.GetStr("mode")
		ids := msg.Get("ids").(string)
		controllerIDs := strings.Split(ids, ",")
		s.queue.TeamStart(teamID)
		s.startNewMatch(controllerIDs, mode, teamID)
	case "teamCall":
		teamID := msg.GetStr("teamID")
		s.queue.TeamCall(teamID)
	case "arduinoModeChange":
		mode := strconv.Itoa(int(ArduinoMode(msg.Get("mode").(float64))))
		am := NewInboxMessage()
		am.SetCmd("mode_change")
		am.Set("mode", mode)
		log.Printf("send mode change:%v\n", mode)
		if mode == "3" {
			s.bgControl(GetOptions().BgIdle)
		}
		s.sends(am, InboxAddressTypeMainArduinoDevice, InboxAddressTypeSubArduinoDevice, InboxAddressTypeDoorArduino, InboxAddressTypeMusicArduino)
		s.sendMsgs("reset", nil, InboxAddressTypeIngameDevice)
	case "queryArduinoList":
		arduinolist := make([]ArduinoController, len(s.aDict))
		i := 0
		for _, controller := range s.aDict {
			arduinolist[i] = *controller
			i += 1
		}
		s.sendMsg("ArduinoList", arduinolist, msg.Address.ID, msg.Address.Type)
	case "stopMatch":
		mid := uint(msg.Get("matchID").(float64))
		if match := s.mDict[mid]; match != nil {
			match.OnMatchCmdArrived(msg)
		}
	case "laserOn":
		s.adminMode = AdminModeDebug
		id := msg.GetStr("id")
		idx := int(msg.Get("num").(float64))
		connected := false
		for _, ac := range s.aDict {
			if ac.Address.ID == id && ac.Online {
				connected = true
				break
			}
		}
		if !connected {
			dd := NewInboxMessage()
			dd.SetCmd("laserInfo")
			dd.Set("id", id)
			dd.Set("ur", "")
			dd.Set("error", 1)
			s.sends(dd, InboxAddressTypeAdminDevice)
			return
		}
		s.laserControl(id, idx, true)
	case "laserOff":
		s.adminMode = AdminModeNormal
		id := msg.GetStr("id")
		idx := int(msg.Get("num").(float64))
		s.laserControl(id, idx, false)
	case "stopListenLaser":
		s.adminMode = AdminModeNormal
		GetLaserPair().Save()
	case "recordLaser":
		key := msg.GetStr("from") + ":" + msg.GetStr("from_idx")
		GetLaserPair().Record(key, msg.GetStr("to"), msg.GetStr("to_idx"), 1)
	case "startQuickCheck":
		if s.qc == nil {
			s.qc = NewQuickChecker(s)
		}
	case "stopQuickCheck":
		if s.qc == nil {
			return
		}
		saveValue := msg.Get("save").(float64)
		save := false
		if saveValue > 0 {
			save = true
		}
		s.qc.Stop(save)
		s.qc = nil
	case "queryQuickCheck":
		if s.qc == nil {
			return
		}
		s.qc.Query()
	}
}

func (s *Srv) startNewMatch(controllerIDs []string, mode string, teamID string) {
	md := s.db.newMatch()
	mid := md.ID
	for _, id := range controllerIDs {
		if p, ok := s.pDict[id]; ok {
			p.MatchID = mid
		} else {
			s.db.saveOrDelMatchData(md)
			s.sends(NewErrorInboxMessage("无效的设备ID"), InboxAddressTypeAdminDevice)
			return
		}
	}
	m := NewMatch(s, controllerIDs, md, mode, teamID, s.isSimulator)
	s.mDict[mid] = m
	go m.Run()
	s.sendMsgs("newMatch", mid, InboxAddressTypeAdminDevice, InboxAddressTypeSimulatorDevice)
}

func (s *Srv) getControllerData() []PlayerController {
	r := make([]PlayerController, len(s.pDict))
	i := 0
	for _, pc := range s.pDict {
		r[i] = *pc
		i += 1
	}
	return r
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

func (s *Srv) wearableControl(status string, cid string) {
	if pc, ok := s.pDict[cid]; ok {
		id, _ := strconv.Atoi(pc.Address.ID)
		idStr := fmt.Sprintf("%03d", id)
		msg := NewInboxMessage()
		msg.SetCmd("STA")
		msg.Set("id", idStr)
		msg.Set("status", status)
		s.sendToOne(msg, pc.Address)
	}
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

func (s *Srv) lasersControl(large []int, small []int) {
	largeAddrs := make([]InboxAddress, 0)
	smallAddrs := make([]InboxAddress, 0)
	for _, info := range GetOptions().MainArduinoInfo {
		if info.LaserNum == 5 {
			largeAddrs = append(largeAddrs, InboxAddress{InboxAddressTypeMainArduinoDevice, info.ID})
		} else {
			smallAddrs = append(smallAddrs, InboxAddress{InboxAddressTypeMainArduinoDevice, info.ID})
		}
	}
	lm := NewInboxMessage()
	lm.SetCmd("laser_ctrl")
	ll := make([]map[string]string, len(large))
	for i, v := range large {
		idx := i + 1
		laser := make(map[string]string)
		laser["laser_n"] = strconv.Itoa(idx)
		if v > 0 {
			laser["laser_s"] = "1"
		} else {
			laser["laser_s"] = "0"
		}
		ll[i] = laser
	}
	lm.Set("laser", ll)
	sm := NewInboxMessage()
	sm.SetCmd("laser_ctrl")
	sl := make([]map[string]string, len(small))
	for i, v := range small {
		idx := i + 6
		laser := make(map[string]string)
		laser["laser_n"] = strconv.Itoa(idx)
		if v > 0 {
			laser["laser_s"] = "1"
		} else {
			laser["laser_s"] = "0"
		}
		sl[i] = laser
	}
	sm.Set("laser", ll)
	s.send(lm, largeAddrs)
	s.send(sm, smallAddrs)
}

func (s *Srv) laserControl(ID string, idx int, openOrClose bool) {
	valid := GetLaserPair().IsValid(ID, idx)
	if !valid && openOrClose {
		return
	}
	msg := NewInboxMessage()
	msg.SetCmd("laser_ctrl")
	laser := make(map[string]string)
	info := arduinoInfoFromID(ID)
	idx += 1
	if info.LaserNum == 5 {
		idx += 5
	}
	laser["laser_n"] = strconv.Itoa(idx)
	if openOrClose {
		laser["laser_s"] = "1"
	} else {
		laser["laser_s"] = "0"
	}
	lasers := []map[string]string{laser}
	msg.Set("laser", lasers)
	addr := InboxAddress{InboxAddressTypeMainArduinoDevice, ID}
	s.sendToOne(msg, addr)
}

// wall参数, 1主墙, 2小墙, 3二者同时
func (s *Srv) ledControl(wall int, mode string, ledT ...string) {
	mainLedList := make([]map[string]string, 0)
	subLedList := make([]map[string]string, 0)
	if wall&1 > 0 {
		if ledT == nil {
			mainLedList = append(mainLedList, map[string]string{"wall": "M", "led_t": "1", "mode": mode})
		} else {
			for _, t := range ledT {
				mainLedList = append(mainLedList, map[string]string{"wall": "M", "led_t": t, "mode": mode})
			}
		}
	}
	if wall&2 > 0 {
		for i := 1; i <= 3; i++ {
			wall := fmt.Sprintf("O%d", i)
			subLedList = append(subLedList, map[string]string{"wall": wall, "led_t": "1", "mode": mode})
		}
		mainLedList = append(mainLedList, map[string]string{"wall": "O1", "led_t": "1", "mode": mode})
	}
	if len(mainLedList) > 0 {
		mainMsg := NewInboxMessage()
		mainMsg.SetCmd("led_ctrl")
		mainMsg.Set("led", mainLedList)
		s.sends(mainMsg, InboxAddressTypeMainArduinoDevice)
	}
	if len(subLedList) > 0 {
		subMsg := NewInboxMessage()
		subMsg.SetCmd("led_ctrl")
		subMsg.Set("led", subLedList)
		s.sends(subMsg, InboxAddressTypeSubArduinoDevice)
	}
}

func (s *Srv) setWallM2M3Auto(isAuto bool) {
	msg := NewInboxMessage()
	msg.SetCmd("btn_ctrl")
	if isAuto {
		msg.Set("useful", "0")
	} else {
		msg.Set("useful", "2")
	}
	msg.Set("mode", "0")
	msg.Set("stage", "0")
	s.sends(msg, InboxAddressTypeMainArduinoDevice)
}

func (s *Srv) ledControlByAddresses(mode string, addrs []InboxAddress) {
	m := NewInboxMessage()
	m.SetCmd("led_ctrl")
	li := make([]map[string]string, 1)
	li[0] = map[string]string{"wall": "M", "led_t": "1", "mode": mode}
	m.Set("led", li)
	s.send(m, addrs)
}

func (s *Srv) ledControlByCell(x int, y int, mode string) {
	ids := GetOptions().mainArduinosByPos(x, y)
	if len(ids) == 0 {
		return
	}
	m := NewInboxMessage()
	m.SetCmd("led_ctrl")
	li := make([]map[string]string, 1)
	li[0] = map[string]string{"wall": "M", "led_t": "1", "mode": mode}
	m.Set("led", li)
	addrs := make([]InboxAddress, len(ids))
	for i, id := range ids {
		addrs[i] = InboxAddress{InboxAddressTypeMainArduinoDevice, id}
	}
	s.send(m, addrs)
}

func (s *Srv) lightControl(mode string) {
	msg := NewInboxMessage()
	msg.SetCmd("light_ctrl")
	msg.Set("light_mode", mode)
	s.sends(msg, InboxAddressTypeMainArduinoDevice)
}

func (s *Srv) bgControl(music string) {
	msg := NewInboxMessage()
	msg.SetCmd("mp3_ctrl")
	msg.Set("music", music)
	s.sends(msg, InboxAddressTypeMusicArduino)
}

func (s *Srv) musicControlByCell(x int, y int, music string) {
	ids := GetOptions().mainArduinosByPos(x, y)
	if len(ids) == 0 {
		return
	}
	msg := NewInboxMessage()
	msg.SetCmd("mp3_ctrl")
	msg.Set("music", music)
	addrs := make([]InboxAddress, len(ids))
	for i, id := range ids {
		addrs[i] = InboxAddress{InboxAddressTypeMainArduinoDevice, id}
	}
	s.send(msg, addrs)
}

func (s *Srv) ledFlowEffect() {
	opt := GetOptions()
	ledList := make([]map[string]string, 3)
	ledList[0] = map[string]string{"wall": "M", "led_t": "2", "mode": "1"}  // M2常亮
	ledList[1] = map[string]string{"wall": "M", "led_t": "3", "mode": "1"}  // M3常亮
	ledList[2] = map[string]string{"wall": "O1", "led_t": "1", "mode": "1"} // O1常亮
	leftList := make([]map[string]string, 4)
	rightList := make([]map[string]string, 4)
	copy(leftList, ledList)
	copy(rightList, ledList)
	leftList[3] = map[string]string{"wall": "M", "led_t": "1", "mode": "49"}
	rightList[3] = map[string]string{"wall": "M", "led_t": "1", "mode": "48"}
	leftArduinos := make([]InboxAddress, 0)
	rightArduinos := make([]InboxAddress, 0)
	for _, info := range opt.MainArduinoInfo {
		addr := InboxAddress{InboxAddressTypeMainArduinoDevice, info.ID}
		if info.LaserDir == "L" {
			leftArduinos = append(leftArduinos, addr)
		} else {
			rightArduinos = append(rightArduinos, addr)
		}
	}
	leftMsg := NewInboxMessage()
	leftMsg.SetCmd("led_ctrl")
	leftMsg.Set("led", leftList)
	s.send(leftMsg, leftArduinos)
	rightMsg := NewInboxMessage()
	rightMsg.SetCmd("led_ctrl")
	rightMsg.Set("led", rightList)
	s.send(rightMsg, rightArduinos)
	s.ledControl(2, "1")
}

func (s *Srv) ledRampageEffect(offs map[string]bool) {
	s.ledControl(2, "21")
	addrsA := make([]InboxAddress, 0)
	addrsB := make([]InboxAddress, 0)
	addrsOff := make([]InboxAddress, len(offs))
	idx := 0
	for _, info := range GetOptions().MainArduinoInfo {
		if _, ok := offs[info.ID]; ok {
			addrsOff[idx] = InboxAddress{InboxAddressTypeMainArduinoDevice, info.ID}
			idx += 1
		} else if info.Type == "A" {
			addrsA = append(addrsA, InboxAddress{InboxAddressTypeMainArduinoDevice, info.ID})
		} else {
			addrsB = append(addrsB, InboxAddress{InboxAddressTypeMainArduinoDevice, info.ID})
		}
	}
	liA := make([]map[string]string, 1)
	liA[0] = map[string]string{"wall": "M", "led_t": "1", "mode": "21"}
	ma := NewInboxMessage()
	ma.SetCmd("led_ctrl")
	ma.Set("led", liA)
	s.send(ma, addrsA)
	liB := make([]map[string]string, 1)
	liB[0] = map[string]string{"wall": "M", "led_t": "1", "mode": "22"}
	mb := NewInboxMessage()
	mb.SetCmd("led_ctrl")
	mb.Set("led", liA)
	s.send(mb, addrsB)
	liO := make([]map[string]string, 1)
	liO[0] = map[string]string{"wall": "M", "led_t": "1", "mode": "24"}
	mo := NewInboxMessage()
	mo.SetCmd("led_ctrl")
	mo.Set("led", liO)
	s.send(mo, addrsOff)
}

func (s *Srv) send(msg *InboxMessage, addrs []InboxAddress) {
	s.inbox.Send(msg, addrs)
}

func (s *Srv) sendToOne(msg *InboxMessage, addr InboxAddress) {
	s.send(msg, []InboxAddress{addr})
}

func (s *Srv) initArduinoControllers() {
	for _, main := range GetOptions().MainArduino {
		addr := InboxAddress{InboxAddressTypeMainArduinoDevice, main}
		controller := NewArduinoController(addr)
		s.aDict[addr.String()] = controller
	}
	for _, sub := range GetOptions().SubArduino {
		addr := InboxAddress{InboxAddressTypeSubArduinoDevice, sub}
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

func (s *Srv) updateArduinoControllerScore(controller *ArduinoController) {
	if !controller.NeedUpdateScore() {
		return
	}
	scoreInfo := GetScoreInfo()
	msg := NewInboxMessage()
	msg.SetCmd("init_score")
	msg.Set("score", scoreInfo)
	msg.Set("upload_time", strconv.Itoa(GetOptions().UploadTime))
	msg.Set("heartbeat_time", strconv.Itoa(GetOptions().HeartbeatTime))
	msg.Set("s_upload_time", strconv.Itoa(GetOptions().SubUploadTime))
	msg.Set("s_heartbeat_time", strconv.Itoa(GetOptions().SubHeartbeatTime))
	s.send(msg, []InboxAddress{controller.Address})
}
