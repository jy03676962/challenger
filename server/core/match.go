package core

import (
	"golang.org/x/net/html/atom"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

var _ = log.Printf

type MatchEventType int

const (
	MatchEventTypeEnd = iota
	MatchEventTypeUpdate
)

const (
	DoorClose = iota
	DoorOpen
)

const (
	READY      = "ready"
	StageRoom1 = "room1"
	StageRoom2 = "room2"
	StageRoom3 = "room3"
	StageRoom4 = "room4"
	StageRoom5 = "room5"
	StageRoom6 = "room6"
	StageEnd   = "end"
)

type MatchEvent struct {
	Type MatchEventType
	ID   uint
	Data interface{}
}

type Match struct {
	opt *MatchOptions
	srv *Srv

	Stage             string
	Step              int
	TotalTime         float64
	OpenDoorDelayTime float64
	CurrentBgm        int

	msgCh   chan *InboxMessage
	closeCh chan bool

	entranceRoom *EntranceRoom
	livingRoom   *Room1
	library      *Room2
	stairRoom    *Room3
	magicLab     *Room4
	starTower    *Room5
	endRoom      *Room6
	exitRoom     *ExitRoom
}

func NewMatch(s *Srv) *Match {
	m := Match{}
	m.Stage = READY
	m.CurrentBgm = 0
	m.srv = s
	m.opt = GetOptions()
	m.TotalTime = 0 //config
	m.msgCh = make(chan *InboxMessage, 1000)
	m.closeCh = make(chan bool)
	m.initHardwareData()
	return &m

}

func (m *Match) initHardwareData() {
	m.livingRoom = NewRoom1()
	m.library = NewRoom2()
	m.stairRoom = NewRoom3()
	m.magicLab = NewRoom4()
	m.starTower = NewRoom5()
	m.endRoom = NewRoom6()
	m.entranceRoom = NewEntranceRoom()
	m.exitRoom = NewExitRoom()
	m.initHardware()
}

func (m *Match) Run() {
	dt := 10 * time.Millisecond
	tickChan := time.Tick(dt)
	for {
		<-tickChan
		m.handleInputs()
		m.gameStage(dt)
	}
}

func (m *Match) OnMatchCmdArrived(cmd *InboxMessage) {
	go func() {
		select {
		case m.msgCh <- cmd:
		case <-m.closeCh:
		}
	}()
}

func (m *Match) handleInputs() bool {
	hasInputs := false
	for {
		select {
		case msg := <-m.msgCh:
			hasInputs = true
			m.handleInput(msg)
		default:
			return hasInputs
		}
	}
}

func (m *Match) handleInput(msg *InboxMessage) { //处理arduino的信息，来改变服务器变量
	cmd := msg.GetCmd()
	if cmd == "hb" {
		id := msg.GetStr("ID")
		addr := InboxAddress{att(id), id}
		switch id {
		case "B-1":
			mp3, _ := strconv.Atoi(msg.GetStr("MP3"))
			if mp3 != m.CurrentBgm {
				m.bgmPlay(m.CurrentBgm)
			}
		case "L-1":
			isChange := false
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("light_ctrl")
			c := []rune(msg.GetStr("RL"))
			var open bool
			for k, v := range c {
				if v == '1' {
					open = true
				} else if v == '0' {
					open = false
				}
				switch k {
				case 1:
					if m.livingRoom.LightStatus != open {
						if m.livingRoom.LightStatus {
							sendMsg.Set("room1", "1")
						} else {
							sendMsg.Set("room1", "0")
						}
						isChange = true
					}
				case 2:
					if m.library.LightStatus != open {
						if m.library.LightStatus {
							sendMsg.Set("room2", "1")
						} else {
							sendMsg.Set("room2", "0")
						}
						isChange = true
					}
				case 3:
					if m.stairRoom.LightStatus != open {
						if m.stairRoom.LightStatus {
							sendMsg.Set("room3", "1")
						} else {
							sendMsg.Set("room3", "0")
						}
						isChange = true
					}
				case 4:
					if m.magicLab.LightStatus != open {
						if m.magicLab.LightStatus {
							sendMsg.Set("room4", "1")
						} else {
							sendMsg.Set("room4", "0")
						}
						isChange = true
					}
				case 5:
					if m.starTower.LightStatus != open {
						if m.starTower.LightStatus {
							sendMsg.Set("room5", "1")
						} else {
							sendMsg.Set("room5", "0")
						}
						isChange = true
					}
				case 6:
					if m.endRoom.LightStatus != open {
						if m.endRoom.LightStatus {
							sendMsg.Set("room6", "1")
						} else {
							sendMsg.Set("room6", "0")
						}
						isChange = true
					}
				}
				if isChange {
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "D-0":
			if msg.GetStr("ST") == "1" {
				m.entranceRoom.DoorStatus = DoorOpen
			} else {
				m.entranceRoom.DoorStatus = DoorClose
			}
		case "R-1-1":
			if msg.GetStr("ST") == "1" {
				m.livingRoom.DoorWardrobe = DoorOpen
			} else {
				m.livingRoom.DoorWardrobe = DoorClose
			}
		case "R-1-3":
			if msg.GetStr("ST") == "1" {
				m.livingRoom.CrystalStatus = 1
			} else {
				m.livingRoom.CrystalStatus = 0
			}
		case "D-1":
			if msg.GetStr("ST") == "1" {
				m.livingRoom.DoorMirror = DoorOpen
			} else {
				m.livingRoom.DoorMirror = DoorClose
			}
		case "R-2-6":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_table")
			if msg.GetStr("U") == "1" {
				if m.library.Table.IsUseful != true {
					sendMsg.Set("useful", "0")
				}
			} else {
				if m.library.Table.IsUseful != false {
					sendMsg.Set("useful", "1")
				}
			}
			if msg.GetStr("F") == "1" {
				m.library.Table.IsFinish = true
			} else {
				m.library.Table.IsFinish = false
			}
			if len(sendMsg.Data) > 2 {
				m.srv.sendToOne(sendMsg, addr)
			}
			m.library.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			angle, _ := strconv.ParseFloat(msg.GetStr("A"), 64)
			if angle != m.library.Table.CurrentAngle {
				m.dealAngle(angle)
			}
			if !m.library.InAnimation && m.library.MagicWords != 0 {
				m.dealMagicWords(m.library, m.library.MagicWords)
			}
		case "R-2-7":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_book")
			if msg.GetStr("ST") == "1" {
				if !m.library.MagicBooksLightStatus[0] {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.library.MagicBooksLightStatus[0] {
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "R-2-8":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_book")
			if msg.GetStr("ST") == "1" {
				if !m.library.MagicBooksLightStatus[1] {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.library.MagicBooksLightStatus[1] {
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "D-2":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("door_ctrl")
			if msg.GetStr("ST") == "1" {
				if m.library.DoorExit != DoorOpen {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.library.DoorExit != DoorClose {
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "R-3-1":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[0] = 0
			} else {
				color := msg.GetStr("C")
				if color != "0" {
					if color != strconv.Itoa(m.stairRoom.Candles[0]) {
						m.stairRoom.Candles[0], _ = strconv.Atoi(color)
					}
				} else {
					m.stairRoom.Candles[0] = 1
				}
			}
		case "R-3-2":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[1] = 0
			} else {
				color := msg.GetStr("C")
				if color != "0" {
					if color != strconv.Itoa(m.stairRoom.Candles[1]) {
						m.stairRoom.Candles[1], _ = strconv.Atoi(color)
					}
				} else {
					m.stairRoom.Candles[1] = 1
				}
			}
		case "R-3-3":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[2] = 0
			} else {
				color := msg.GetStr("C")
				if color != "0" {
					if color != strconv.Itoa(m.stairRoom.Candles[2]) {
						m.stairRoom.Candles[2], _ = strconv.Atoi(color)
					}
				} else {
					m.stairRoom.Candles[2] = 1
				}
			}
		case "R-3-4":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[3] = 0
			} else {
				color := msg.GetStr("C")
				if color != "0" {
					if color != strconv.Itoa(m.stairRoom.Candles[3]) {
						m.stairRoom.Candles[3], _ = strconv.Atoi(color)
					}
				} else {
					m.stairRoom.Candles[3] = 1
				}
			}
		case "R-3-5":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[4] = 0
			} else {
				color := msg.GetStr("C")
				if color != "0" {
					if color != strconv.Itoa(m.stairRoom.Candles[4]) {
						m.stairRoom.Candles[4], _ = strconv.Atoi(color)
					}
				} else {
					m.stairRoom.Candles[4] = 1
				}
			}
		case "R-3-6":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[5] = 0
			} else {
				color := msg.GetStr("C")
				if color != "0" {
					if color != strconv.Itoa(m.stairRoom.Candles[5]) {
						m.stairRoom.Candles[5], _ = strconv.Atoi(color)
					}
				} else {
					m.stairRoom.Candles[5] = 1
				}
			}
		case "R-3-7":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_table")
			if msg.GetStr("U") == "1" {
				if m.stairRoom.Table.IsUseful != true {
					sendMsg.Set("useful", "0")
				}
			} else {
				if m.stairRoom.Table.IsUseful != false {
					sendMsg.Set("useful", "1")
				}
			}
			if msg.GetStr("F") == "1" {
				if m.stairRoom.Table.IsFinish != true {
					sendMsg.Set("finish", "0")
				}
			} else {
				if m.stairRoom.Table.IsFinish != false {
					sendMsg.Set("finish", "1")
				}
			}
			if len(sendMsg.Data) > 2 {
				m.srv.sendToOne(sendMsg, addr)
			}
			m.stairRoom.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			if !m.stairRoom.InAnimation && m.stairRoom.MagicWords != 0 {
				m.dealMagicWords(m.stairRoom, m.stairRoom.MagicWords)
			}
		case "D-3":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("door_ctrl")
			if msg.GetStr("ST") == "1" {
				if m.stairRoom.DoorExit != DoorOpen {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.stairRoom.DoorExit != DoorClose {
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "R-4-1":
			if !m.magicLab.Stands[0].IsPowerOn || m.magicLab.Stands[0].IsPowerful {
				return
			}
			status := msg.GetStr("P")
			if status != m.magicLab.Stands[0].Power {
				m.magicLab.Stands[0].Power = status
				power := make([]map[string]string, 0)
				power = append(power, map[string]string{"power_type": "1", "status": status})
				m.srv.powerStatus(power)
				if status == "2" {
					m.magicLab.Stands[0].IsPowerful = true
				}
			}
		case "R-4-2":
			if !m.magicLab.Stands[1].IsPowerOn || m.magicLab.Stands[1].IsPowerful {
				return
			}
			status := msg.GetStr("P")
			if status != m.magicLab.Stands[1].Power {
				m.magicLab.Stands[1].Power = status
				power := make([]map[string]string, 0)
				power = append(power, map[string]string{"power_type": "2", "status": status})
				m.srv.powerStatus(power)
				if status == "2" {
					m.magicLab.Stands[1].IsPowerful = true
				}
			}
		case "R-4-3":
			if !m.magicLab.Stands[2].IsPowerOn || m.magicLab.Stands[2].IsPowerful {
				return
			}
			status := msg.GetStr("P")
			if status != m.magicLab.Stands[2].Power {
				m.magicLab.Stands[2].Power = status
				power := make([]map[string]string, 0)
				power = append(power, map[string]string{"power_type": "3", "status": status})
				m.srv.powerStatus(power)
				if status == "2" {
					m.magicLab.Stands[2].IsPowerful = true
				}
			}
		case "R-4-4":
			if !m.magicLab.Stands[3].IsPowerOn || m.magicLab.Stands[3].IsPowerful {
				return
			}
			status := msg.GetStr("P")
			if status != m.magicLab.Stands[3].Power {
				m.magicLab.Stands[3].Power = status
				power := make([]map[string]string, 0)
				power = append(power, map[string]string{"power_type": "4", "status": status})
				m.srv.powerStatus(power)
				if status == "2" {
					m.magicLab.Stands[3].IsPowerful = true
				}
			}
		case "R-4-5":
			if msg.GetStr("U") == "1" {
				m.magicLab.Table.IsUseful = true
			} else {
				m.magicLab.Table.IsUseful = false
			}
			if msg.GetStr("F") == "1" {
				m.magicLab.Table.IsFinish = true
			} else {
				m.magicLab.Table.IsFinish = false
			}
			m.magicLab.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			if m.magicLab.MagicWords != 0 {
				m.dealMagicWords(m.magicLab, m.magicLab.MagicWords)
			}
		case "R-4-6":
			sendMsg := NewInboxMessage()
			if msg.GetStr("ST") == "1" {
				if !m.magicLab.DeskLight {
					sendMsg.SetCmd("book_desk")
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.magicLab.DeskLight {
					sendMsg.SetCmd("book_desk")
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "D-4":
			sendMsg := NewInboxMessage()
			log.Println("server:", m.magicLab.DoorExit)
			log.Println("arudino:", msg.GetStr("ST"))
			if msg.GetStr("ST") == "1" {
				if m.magicLab.DoorExit != DoorOpen {
					sendMsg.SetCmd("door_ctrl")
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.magicLab.DoorExit != DoorClose {
					sendMsg.SetCmd("door_ctrl")
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "R-5-6":
			if msg.GetStr("U") == "1" {
				m.starTower.Table.IsUseful = true
			} else {
				m.starTower.Table.IsUseful = false
			}
			if msg.GetStr("F") == "1" {
				m.starTower.Table.IsFinish = true
			} else {
				m.starTower.Table.IsFinish = false
			}
			m.starTower.CurrentConstellationLight, _ = strconv.Atoi(msg.GetStr("S"))
			if m.starTower.CurrentConstellationLight != 0 {
				m.dealStar(m.starTower.CurrentConstellationLight)
			}
			m.starTower.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			if m.starTower.MagicWords != 0 {
				m.dealMagicWords(m.starTower, m.starTower.MagicWords)
			}
		case "R-5-7":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_rob")
			if msg.GetStr("ST") == "1" {
				if m.starTower.DoorMagicRod != DoorOpen {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.starTower.DoorMagicRod != DoorClose {
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "R-5-8":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("light_ctrl")
			if msg.GetStr("ST") == "1" {
				if m.starTower.LightWall != true {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.starTower.LightWall != false {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "D-5":
			sendMsg := NewInboxMessage()
			if msg.GetStr("ST") == "1" {
				if m.starTower.DoorExit != DoorOpen {
					sendMsg.SetCmd("door_ctrl")
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)

				}
			} else {
				if m.starTower.DoorExit != DoorClose {
					sendMsg.SetCmd("door_ctrl")
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)

				}
			}
		case "R-6-1":
			status := msg.GetStr("ST")
			if status == "3" && !m.endRoom.PowerPointFull[0] {
				ty, _ := strconv.Atoi(msg.GetStr("TY"))
				m.endRoom.PowerPoint[0] = ty
				m.endRoom.PowerPointFull[0] = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg, addr)
				m.broadSymbolToArduino(0)
				m.endRoom.CurrentSymbol = 0
			} else if status == "1" {
				m.endRoom.PowerPointFull[0] = false
			}
		case "R-6-2":
			status := msg.GetStr("ST")
			if status == "3" && !m.endRoom.PowerPointFull[1] {
				ty, _ := strconv.Atoi(msg.GetStr("TY"))
				m.endRoom.PowerPoint[1] = ty
				m.endRoom.PowerPointFull[1] = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg, addr)
				m.broadSymbolToArduino(0)
				m.endRoom.CurrentSymbol = 0
			} else if status == "1" {
				m.endRoom.PowerPointFull[1] = false
			}
		case "R-6-3":
			status := msg.GetStr("ST")
			if status == "3" && !m.endRoom.PowerPointFull[2] {
				ty, _ := strconv.Atoi(msg.GetStr("TY"))
				m.endRoom.PowerPoint[2] = ty
				m.endRoom.PowerPointFull[2] = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg, addr)
				m.broadSymbolToArduino(0)
				m.endRoom.CurrentSymbol = 0
			} else if status == "1" {
				m.endRoom.PowerPointFull[2] = false
			}
		case "R-6-4":
			status := msg.GetStr("ST")
			if status == "3" && !m.endRoom.PowerPointFull[3] {
				ty, _ := strconv.Atoi(msg.GetStr("TY"))
				m.endRoom.PowerPoint[3] = ty
				m.endRoom.PowerPointFull[3] = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg, addr)
				m.broadSymbolToArduino(0)
				m.endRoom.CurrentSymbol = 0
			} else if status == "1" {
				m.endRoom.PowerPointFull[3] = false
			}
		case "R-6-5":
			status := msg.GetStr("ST")
			if status == "3" && !m.endRoom.PowerPointFull[4] {
				ty, _ := strconv.Atoi(msg.GetStr("TY"))
				m.endRoom.PowerPoint[4] = ty
				m.endRoom.PowerPointFull[4] = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg, addr)
				m.broadSymbolToArduino(0)
				m.endRoom.CurrentSymbol = 0
			} else if status == "1" {
				m.endRoom.PowerPointFull[4] = false
			}
		case "R-6-6":
			status := msg.GetStr("ST")
			if status == "3" && !m.endRoom.PowerPointFull[5] {
				ty, _ := strconv.Atoi(msg.GetStr("TY"))
				m.endRoom.PowerPoint[5] = ty
				m.endRoom.PowerPointFull[5] = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg, addr)
				m.broadSymbolToArduino(0)
				m.endRoom.CurrentSymbol = 0
			} else if status == "1" {
				m.endRoom.PowerPointFull[5] = false
			}
		case "R-6-7":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_table")
			if msg.GetStr("U") == "1" {
				if m.endRoom.Table.IsUseful != true {
					sendMsg.Set("useful", "0")
				}
			} else {
				if m.endRoom.Table.IsUseful != false {
					sendMsg.Set("useful", "1")
				}
			}
			if msg.GetStr("F") == "1" {
				if m.endRoom.Table.IsFinish != true {
					sendMsg.Set("finish", "0")
				}
			} else {
				if m.endRoom.Table.IsFinish != false {
					sendMsg.Set("finish", "1")
				}
			}
			if msg.GetStr("D") == "1" {
				if m.endRoom.Table.IsDestroyed != true {
					sendMsg.Set("destroyed", "0")
				}
			} else {
				if m.endRoom.Table.IsDestroyed != false {
					sendMsg.Set("destroyed", "1")
				}
			}
			if len(sendMsg.Data) > 2 {
				m.srv.sendToOne(sendMsg, addr)
			}
			if m.endRoom.Table.IsUseful {
				cs, _ := strconv.Atoi(msg.GetStr("TY"))
				if cs != m.endRoom.CurrentSymbol && cs != 0 {
					m.endRoom.CurrentSymbol = cs
					m.broadSymbolToArduino(m.endRoom.CurrentSymbol)
				}
			}
			m.endRoom.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.endRoom, m.endRoom.MagicWords)
		case "R-6-9":
			return
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("water_light")
			if msg.GetStr("ST") == "1" {
				if !m.endRoom.WaterLight {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.endRoom.WaterLight {
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "D-6":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("door_ctrl")
			if msg.GetStr("ST") == "1" {
				if m.endRoom.DoorExit != DoorOpen {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.endRoom.DoorExit != DoorClose {
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		}
	} else if cmd == "nextStep" {
		switch m.Stage {
		case READY:
			m.setStage(StageRoom1)
		case StageRoom1:
			m.setStage(StageRoom2)
		case StageRoom2:
			//m.library.Step++
			// if m.library.Step < 4 {
			//log.Println("jump 1 step,current step library :", m.library.Step)
			//} else {
			//m.setStage(StageRoom3)
			//}
			m.library.Step = 3
		case StageRoom3:
			//m.stairRoom.Step++
			// if m.stairRoom.Step < 4 {
			//log.Println("jump 1 step,current step: stariRoom ", m.stairRoom.Step)
			//} else {
			//m.setStage(StageRoom4)
			//}
			m.stairRoom.Step = 4
		case StageRoom4:
			// m.magicLab.Step++
			//if m.magicLab.Step < 5 {
			//log.Println("jump 1 step,current step magicLab :", m.magicLab.Step)
			//} else {
			//m.setStage(StageRoom5)
			//}
			m.magicLab.Step = 4
		case StageRoom5:
			//m.starTower.Step++
			//if m.starTower.Step < 4 {
			//log.Println("jump 1 step,current step starTower :", m.starTower.Step)
			//} else {
			//m.setStage(StageRoom6)
			//}
			m.starTower.Table.IsFinish = true
			m.starTower.Table.IsDestroyed = true
			m.starTower.Step = 3
		case StageRoom6:
			//m.endRoom.Step++
			//if m.endRoom.Step < 5 {
			//log.Println("jump 1 step,current step endRoom :", m.endRoom.Step)
			//} else {
			//m.setStage(StageEnd)
			//}
			m.setStage(StageEnd)
		}
	} else if cmd == "init" {
		m.reset()
	}
}

func (m *Match) setStage(s string) {
	if m.Stage == s {
		return
	}
	switch s {
	case StageRoom1:
		if m.CurrentBgm != 2 {
			m.CurrentBgm = 2
			m.bgmPlay(m.CurrentBgm)
		}
	case StageRoom2:
		if m.CurrentBgm != 3 {
			m.CurrentBgm = 3
			m.bgmPlay(m.CurrentBgm)
		}
	case StageRoom3:
		if m.CurrentBgm != 4 {
			m.CurrentBgm = 4
			m.bgmPlay(m.CurrentBgm)
		}
	case StageRoom4:
		if m.CurrentBgm != 5 {
			m.CurrentBgm = 5
			m.bgmPlay(m.CurrentBgm)
		}
	case StageRoom5:
		if m.CurrentBgm != 6 {
			m.CurrentBgm = 6
			m.bgmPlay(m.CurrentBgm)
		}
	case StageRoom6:
		if m.CurrentBgm != 7 {
			m.CurrentBgm = 7
			m.bgmPlay(m.CurrentBgm)
		}
	case StageEnd:
		if m.endRoom.Ending == 0 {
			m.CurrentBgm = 0
			m.bgmPlay(m.CurrentBgm)
		}
	}
	log.Printf("game stage:%v\n", s)
	m.Stage = s
}

func (m *Match) gameStage(dt time.Duration) {
	if m.Stage == "" {
		log.Println("game stage error!")
		return
	}
	if m.Stage != READY {
		if m.Stage == StageRoom6 && m.endRoom.Step != 1 {
			m.TotalTime = m.endRoom.LastTime
		} else if m.Stage != StageEnd {
			m.TotalTime += dt.Seconds()
		}
	}
	switch m.Stage {
	case READY:
		m.Step = 0
		if m.CurrentBgm != 1 {
			m.CurrentBgm = 1
			m.bgmPlay(m.CurrentBgm)
		}
	case StageRoom1:
		m.Step = 0
		if m.livingRoom.DoorMirror == DoorOpen {
			m.room1Animation()
			log.Println("room 1 finish!")
		}
	case StageRoom2:
		if m.library.Step == 1 {
			if m.fakeActNum() >= 5 {
				if m.ensureFakeBooks() {
					m.fakeBooksAnimation(dt)
				} else {
					m.fakeBooksErrorAnimation(dt)
				}
			}
			m.Step = 1
		} else if m.library.Step == 2 {
			if m.library.Table.IsFinish {
				m.tableFinish(m.library)
			}
			if m.library.Table.IsDestroyed {
				m.library.Step = 3
			}
			m.Step = 2
		} else if m.library.Step == 3 {
			m.library.DoorExit = DoorOpen
			m.endingAnimation(StageRoom2, dt)
			m.Step = 3
		}
	case StageRoom3:
		if m.stairRoom.Step == 1 {
			if m.ensureCandlesPoweron() {
				m.magicTableAnimation(StageRoom3, dt)
				m.stairRoom.Step = 2
				log.Println("all candles on!")
			}
			m.Step = 1
		} else if m.stairRoom.Step == 2 {
			if m.ensureCandlesColorNum() {
				m.tableFinish(m.stairRoom)
				m.stairRoom.Step = 3
				log.Println("all candles color turn!")
			}
			m.Step = 2
		} else if m.stairRoom.Step == 3 {
			if m.ensureCandlesColor() {
				if !m.stairRoom.Table.IsFinish {
					m.stairRoom.Table.IsFinish = true
					sendMsg := NewInboxMessage()
					sendMsg.SetCmd("magic_table")
					sendMsg.Set("useful", "1")
					sendMsg.Set("finish", "1")
					addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-3-7"}
					m.srv.sendToOne(sendMsg, addr)
					log.Println("all candles color right!")
				}
				if m.stairRoom.Table.IsDestroyed {
					m.stairRoom.Step = 4
				}
			} else {
				if m.stairRoom.Table.IsFinish {
					m.stairRoom.Table.IsFinish = false
					sendMsg := NewInboxMessage()
					sendMsg.SetCmd("magic_table")
					sendMsg.Set("useful", "1")
					sendMsg.Set("finish", "0")
					addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-3-7"}
					m.srv.sendToOne(sendMsg, addr)
					log.Println("some candles color wrong!")
				}
			}
			m.Step = 3
		} else if m.stairRoom.Step == 4 {
			m.stairRoom.DoorExit = DoorOpen
			m.endingAnimation(StageRoom3, dt)
			m.Step = 4
		}
	case StageRoom4:
		if m.magicLab.Step == 1 {
			m.Step = 1
			if m.ensureMagicStandsPowerOn() {
				m.magicTableAnimation(StageRoom4, dt)
				m.magicLab.Step = 2
				log.Println("all magic stands launch!")
				m.magicLab.Table.IsUseful = true

			}
		} else if m.magicLab.Step == 2 {
			m.Step = 2
			if m.ensureMagicStandsPowerFul() {
				m.magicLab.Table.IsFinish = true
				m.magicLab.Step = 3
				log.Println("all magic stands poweful!")
			}
		} else if m.magicLab.Step == 3 {
			m.Step = 3
			if m.magicLab.Table.IsDestroyed {
				m.magicLab.Step = 4
				log.Println("magic table destroyed!")
			}
		} else if m.magicLab.Step == 4 {
			m.Step = 4
			m.magicLab.DoorExit = DoorOpen
			m.endingAnimation(StageRoom4, dt)
		}
	case StageRoom5:
		if m.starTower.Step == 1 {
			m.Step = 1
			if m.starTower.Table.IsUseful {
				m.starTower.Step = 2
				log.Println("magic table is useful!!")
			}
		} else if m.starTower.Step == 2 {
			m.Step = 2
			if m.ensureConstellationSymbol() || m.starTower.Table.IsFinish {
				m.starTower.Table.IsFinish = true
				m.starTower.Step = 3
				log.Println("constellation right!")
			}
		} else if m.starTower.Step == 3 {
			m.Step = 3
			if m.starTower.Table.IsDestroyed && m.starTower.Table.IsFinish {
				m.starTower.DoorExit = DoorOpen
				m.starTower.DoorMagicRod = DoorOpen
				m.endingAnimation(StageRoom5, dt)
			}

		}
	case StageRoom6:
		if m.endRoom.Step == 1 {
			m.Step = 1
			//if m.endRoom.NextStep == 2 {
			//m.amMagicAnimation()
			//m.endRoom.Step = 2
			//log.Println("room 6 step 1 finish!")
			//}
		} else if m.endRoom.Step == 2 {
			m.Step = 2
			sec := dt.Seconds()
			//if m.exitRoom.ButtonNextStage { //endroom 数据维护需要锁
			if m.CurrentBgm != 8 {
				m.bgmPlay(8) //bgm
				m.CurrentBgm = 8
			}
			if !m.endRoom.Table.IsUseful {
				m.magicTableAnimation(StageRoom6, dt)
				m.endRoom.LastTime = m.opt.Room6LastTime / 1000
				m.endRoom.CandleTime = m.opt.Room6LastTime / 7
				m.endRoom.CurrentCandle = 0
			} else {
				m.endRoom.LastTime = math.Max(m.endRoom.LastTime-sec, 0)
				m.endRoom.CandleTime -= sec * 1000
				leftTime := int(m.endRoom.CandleTime) / 1000
				if m.endRoom.CurrentCandle < 7 {
					if leftTime <= 0 {
						log.Println("current candle off", m.endRoom.CurrentCandle)
						sendMsg := NewInboxMessage()
						sendMsg.SetCmd("led_candle")
						sendMsg.Set("mode", "1")
						candles := make([]map[string]string, 0)
						candles = append(candles, map[string]string{"candle_n": strconv.Itoa(m.endRoom.CurrentCandle), "color": "0"})
						sendMsg.Set("candles", candles)
						addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-8"}
						m.srv.sendToOne(sendMsg, addr)
						m.endRoom.CandleTime = m.opt.Room6LastTime / 7
						m.endRoom.CurrentCandle++
					}
				} else {
					m.endRoom.Ending = 2
					m.endRoom.Step = 3
					log.Println("bad ending!")
				}
				if m.ensureElementSymbol() {
					sendMsg := NewInboxMessage()
					sendMsg.SetCmd("power_point")
					sendMsg.Set("mode", "1")
					addrs := []InboxAddress{
						{InboxAddressTypeRoomArduinoDevice, "R-6-1"},
						{InboxAddressTypeRoomArduinoDevice, "R-6-2"},
						{InboxAddressTypeRoomArduinoDevice, "R-6-3"},
						{InboxAddressTypeRoomArduinoDevice, "R-6-4"},
						{InboxAddressTypeRoomArduinoDevice, "R-6-5"},
						{InboxAddressTypeRoomArduinoDevice, "R-6-6"},
					}
					m.srv.send(sendMsg, addrs)
					m.endRoom.Step = 3
					m.endRoom.Ending = 1
					log.Println("good ending!")
				}
			}
		} else if m.endRoom.Step == 3 {
			m.Step = 3
			if m.endRoom.Ending == 1 { //goodending
				if m.CurrentBgm != 9 {
					m.bgmPlay(9) //bgm
					m.CurrentBgm = 9
					m.endingAnimation(StageRoom6, dt)
				}
			} else if m.endRoom.Ending == 2 { //badending
				if m.CurrentBgm != 10 {
					m.bgmPlay(10) //bgm
					m.CurrentBgm = 10
					m.endingAnimation(StageRoom6, dt)
				}
			}
		} else if m.endRoom.Step == 4 {
			m.Step = 4
			if m.CurrentBgm != 11 {
				m.CurrentBgm = 11
				m.bgmPlay(m.CurrentBgm)
				m.endRoom.WaterLight = false

				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("led_candle")
				sendMsg.Set("light_status", "0")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-8"}

				m.srv.sendToOne(sendMsg, addr)
				sendMsg2 := NewInboxMessage()
				sendMsg2.SetCmd("water_light")
				sendMsg2.Set("status", "0")
				addr2 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-9"}
				m.srv.sendToOne(sendMsg2, addr2)
			}
		}
	case StageEnd:
		m.Step = 0
		if m.endRoom.DoorExit != DoorOpen {
			m.endRoom.DoorExit = DoorOpen
			sendMsg3 := NewInboxMessage()
			sendMsg3.SetCmd("door_ctrl")
			sendMsg3.Set("status", "1")
			addr3 := InboxAddress{InboxAddressTypeDoorArduino, "D-6"}
			m.srv.sendToOne(sendMsg3, addr3)
			log.Println("game over!")
		}
	}
	m.updateStage()
}

func (m *Match) updateStage() {
	if m.Stage == "" {
		log.Println("game stage error!")
		return
	}
	switch m.Stage {
	case StageRoom1:
		if m.livingRoom.DoorMirror == DoorOpen {
			m.setStage(StageRoom2)
		}
	case StageRoom2:
		if m.library.DoorExit == DoorOpen {
			m.setStage(StageRoom3)
		}
	case StageRoom3:
		if m.stairRoom.DoorExit == DoorOpen {
			m.setStage(StageRoom4)
		}
	case StageRoom4:
		if m.magicLab.DoorExit == DoorOpen {
			m.setStage(StageRoom5)
		}
	case StageRoom5:
		if m.starTower.DoorExit == DoorOpen {
			m.setStage(StageRoom6)
		}
	case StageRoom6:
		if m.endRoom.DoorExit == DoorOpen {
			m.setStage(StageEnd)
		}
	case StageEnd:
	}
}

func (m *Match) reset() {
	m.initHardwareData()
	m.Stage = READY
	m.Step = 0
	m.TotalTime = 0
	m.CurrentBgm = 1
	m.bgmPlay(m.CurrentBgm)

	doorMsg := NewInboxMessage()
	doorMsg.SetCmd("door_ctrl")
	doorMsg.Set("useful", "0")
	addr := InboxAddress{InboxAddressTypeDoorArduino, "D-0"}
	m.srv.sendToOne(doorMsg, addr)
	log.Println("game reset success!")
}

func (m *Match) initHardware() {
	sendMsg1 := NewInboxMessage()
	sendMsg1.SetCmd("mode_change")
	sendMsg1.Set("mode", "1")
	m.srv.sends(sendMsg1, InboxAddressTypeDoorArduino, InboxAddressTypeRoomArduinoDevice, InboxAddressTypeMusicArduino)

	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("reset")
	m.srv.sends(sendMsg, InboxAddressTypeDoorArduino, InboxAddressTypeRoomArduinoDevice, InboxAddressTypeMusicArduino)

	sendMsg2 := NewInboxMessage()
	m.magicLab.DeskLight = true
	sendMsg2.SetCmd("book_desk")
	sendMsg2.Set("status", "1")
	addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-6"}
	m.srv.sendToOne(sendMsg2, addr)
}

func (m *Match) destoryFailed(room interface{}) {
	sendMsg := NewInboxMessage()
	switch room.(type) {
	case *Room3:
		m.bgmPlay(4)
		addrs := []InboxAddress{
			{InboxAddressTypeRoomArduinoDevice, "R-3-1"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-2"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-3"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-4"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-5"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-6"}}
		sendMsg.SetCmd("candle_ctrl")
		sendMsg.Set("mode", "1")
		m.srv.send(sendMsg, addrs)
	case *Room5:
		m.bgmPlay(6)
	}
}

func (m *Match) tableFinish(room interface{}) {
	m.bgmPlay(12)
	sendMsg := NewInboxMessage()
	switch room.(type) {
	case *Room2:
		addrs := []InboxAddress{
			{InboxAddressTypeRoomArduinoDevice, "R-2-2"},
			{InboxAddressTypeRoomArduinoDevice, "R-2-3"},
			{InboxAddressTypeRoomArduinoDevice, "R-2-4"},
			{InboxAddressTypeRoomArduinoDevice, "R-2-5"}}
		sendMsg.SetCmd("led_candle")
		sendMsg.Set("mode", "1")
		m.srv.send(sendMsg, addrs)
	case *Room3:
		addrs := []InboxAddress{
			{InboxAddressTypeRoomArduinoDevice, "R-3-1"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-2"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-3"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-4"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-5"},
			{InboxAddressTypeRoomArduinoDevice, "R-3-6"}}
		sendMsg.SetCmd("candle_ctrl")
		sendMsg.Set("mode", "2")
		m.srv.send(sendMsg, addrs)
	case *Room4:
		addrs := []InboxAddress{
			{InboxAddressTypeRoomArduinoDevice, "R-4-1"},
			{InboxAddressTypeRoomArduinoDevice, "R-4-2"},
			{InboxAddressTypeRoomArduinoDevice, "R-4-3"},
			{InboxAddressTypeRoomArduinoDevice, "R-4-4"}}
		sendMsg.SetCmd("magic_desk")
		sendMsg.Set("mode", "2")
		m.srv.send(sendMsg, addrs)
	case *Room5:
	}
}

//room1
func (m *Match) room1Animation() {

}

//room2
func (m *Match) fakeActNum() int {
	num := 0
	for _, v := range m.library.FakeBooks {
		if v {
			num++
		}
	}
	return num
}
func (m *Match) fakeBooksAnimation(dt time.Duration) {
	sec := dt.Seconds()
	if !m.library.InAnimation {
		m.library.FakeAnimationTime = m.opt.FakeAnimationTime / 1000
		m.library.InAnimation = true
		m.library.FakeAnimationStep = 1
		m.library.CandleDelay = 0
		m.library.CandleMode = 0
	} else {
		//需要delay一个时间，为了方便第5本书亮起来
		m.library.FakeAnimationTime = math.Max(m.library.FakeAnimationTime-sec, 0)
		if m.library.FakeAnimationTime == 0 {
			switch m.library.FakeAnimationStep {
			case 1:
				if m.library.FakeAnimationTime == 0 {
					addrs := []InboxAddress{
						{InboxAddressTypeRoomArduinoDevice, "R-2-7"},
						{InboxAddressTypeRoomArduinoDevice, "R-2-8"},
					}
					m.library.MagicBooksLightStatus[0] = false
					m.library.MagicBooksLightStatus[1] = false
					sendMsg := NewInboxMessage()
					sendMsg.SetCmd("magic_book")
					sendMsg.Set("status", "0")
					m.srv.send(sendMsg, addrs)
					m.srv.fakeBooksControl("4", "0", "R-2-9")
					m.srv.fakeBooksControl("4", "0", "R-2-10")
					m.srv.fakeBooksControl("4", "0", "R-2-11")
					m.srv.fakeBooksControl("4", "0", "R-2-12")
					m.srv.fakeBooksControl("4", "0", "R-2-13")
					m.srv.fakeBooksControl("4", "0", "R-2-14")
					m.srv.fakeBooksControl("4", "0", "R-2-15")
					m.srv.fakeBooksControl("4", "0", "R-2-16")
					m.srv.fakeBooksControl("4", "0", "R-2-17")
					m.srv.fakeBooksControl("4", "0", "R-2-18")
					m.srv.fakeBooksControl("4", "0", "R-2-19")
					m.srv.fakeBooksControl("4", "0", "R-2-20")
					m.srv.fakeBooksControl("4", "0", "R-2-21")
					m.srv.fakeBooksControl("4", "0", "R-2-22")
					m.srv.fakeBooksControl("4", "0", "R-2-23")
					m.library.FakeAnimationTime = m.opt.FakeAnimationTime / 1000
					m.library.FakeAnimationStep++
					log.Println("step 1,light off and mode 4 start!")
				}
				//send step 1 灯光全灭
			case 2: //第二组 2,3 并且蜡烛、灯箱开始
				//TODO
				m.library.CandleDelay = math.Max(m.library.CandleDelay-sec, 0)
				if m.library.CandleDelay == 0 {
					candlesN := make([]map[string]string, 1)
					candlesS := make([]map[string]string, 1)
					if m.library.CandleMode == 0 {
						candlesN[0] = map[string]string{"candle": "0", "color": "1"}
						m.srv.candlesControl(candlesN, "R-2-2")
						candlesS[0] = map[string]string{"candle": "0", "color": "1"}
						m.srv.candlesControl(candlesS, "R-2-4")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 1 {
						candlesN[0] = map[string]string{"candle": "1", "color": "1"}
						m.srv.candlesControl(candlesN, "R-2-2")
						candlesS[0] = map[string]string{"candle": "1", "color": "1"}
						m.srv.candlesControl(candlesS, "R-2-4")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 2 {
						candlesN[0] = map[string]string{"candle": "2", "color": "1"}
						m.srv.candlesControl(candlesN, "R-2-2")
						candlesS[0] = map[string]string{"candle": "2", "color": "1"}
						m.srv.candlesControl(candlesS, "R-2-4")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 3 {
						candlesN[0] = map[string]string{"candle": "0", "color": "1"}
						m.srv.candlesControl(candlesN, "R-2-3")
						candlesS[0] = map[string]string{"candle": "0", "color": "1"}
						m.srv.candlesControl(candlesS, "R-2-5")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 4 {
						candlesN[0] = map[string]string{"candle": "1", "color": "1"}
						m.srv.candlesControl(candlesN, "R-2-3")
						candlesS[0] = map[string]string{"candle": "1", "color": "1"}
						m.srv.candlesControl(candlesS, "R-2-5")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 5 {
						candlesN[0] = map[string]string{"candle": "2", "color": "1"}
						m.srv.candlesControl(candlesN, "R-2-3")
						candlesS[0] = map[string]string{"candle": "2", "color": "1"}
						m.srv.candlesControl(candlesS, "R-2-5")
						m.library.CandleDelay = 0
						m.library.CandleMode = 0
						m.library.FakeAnimationTime = m.opt.FakeAnimationTime / 1000
						m.library.FakeAnimationStep++
					}
				}
			case 3: //第三组 4,5,6
				m.library.Table.IsUseful = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("useful", "1")
				sendMsg.Set("InAnimation", "1")
				sendMsg.Set("time", "2000")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-2-6"}
				m.srv.sendToOne(sendMsg, addr)
				m.library.FakeAnimationTime = opt.FakeAnimationTime / 1000
				m.library.FakeAnimationStep++
			case 4: //第四组 7,8,9,10//记忆水晶
				m.library.FakeAnimationTime = opt.FakeAnimationTime / 1000
				m.library.FakeAnimationStep++
			case 5: //第五组 11,12,13,14,15 蜡烛、灯箱换颜色
				m.library.CandleDelay = math.Max(m.library.CandleDelay-sec, 0)
				if m.library.CandleDelay == 0 {
					candlesN := make([]map[string]string, 1)
					candlesS := make([]map[string]string, 1)
					if m.library.CandleMode == 0 {
						candlesN[0] = map[string]string{"candle": "0", "color": "2"}
						m.srv.candlesControl(candlesN, "R-2-2")
						candlesS[0] = map[string]string{"candle": "0", "color": "3"}
						m.srv.candlesControl(candlesS, "R-2-4")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 1 {
						candlesN[0] = map[string]string{"candle": "1", "color": "2"}
						m.srv.candlesControl(candlesN, "R-2-2")
						candlesS[0] = map[string]string{"candle": "1", "color": "3"}
						m.srv.candlesControl(candlesS, "R-2-4")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 2 {
						candlesN[0] = map[string]string{"candle": "2", "color": "2"}
						m.srv.candlesControl(candlesN, "R-2-2")
						candlesS[0] = map[string]string{"candle": "2", "color": "3"}
						m.srv.candlesControl(candlesS, "R-2-4")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 3 {
						candlesN[0] = map[string]string{"candle": "0", "color": "2"}
						m.srv.candlesControl(candlesN, "R-2-3")
						candlesS[0] = map[string]string{"candle": "0", "color": "3"}
						m.srv.candlesControl(candlesS, "R-2-5")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 4 {
						candlesN[0] = map[string]string{"candle": "1", "color": "2"}
						m.srv.candlesControl(candlesN, "R-2-3")
						candlesS[0] = map[string]string{"candle": "1", "color": "3"}
						m.srv.candlesControl(candlesS, "R-2-5")
						m.library.CandleDelay = 0.5
						m.library.CandleMode++
					} else if m.library.CandleMode == 5 {
						candlesN[0] = map[string]string{"candle": "2", "color": "2"}
						m.srv.candlesControl(candlesN, "R-2-3")
						candlesS[0] = map[string]string{"candle": "2", "color": "3"}
						m.srv.candlesControl(candlesS, "R-2-5")
						m.library.CandleDelay = 0.5
						m.library.CandleMode = 0
						m.library.FakeAnimationTime = m.opt.FakeAnimationTime / 1000
						m.library.FakeAnimationStep++
					}
				}
			case 6:
				m.srv.fakeBooksControl("3", "0", "R-2-9")
				m.srv.fakeBooksControl("3", "0", "R-2-10")
				m.srv.fakeBooksControl("3", "0", "R-2-11")
				m.srv.fakeBooksControl("3", "0", "R-2-12")
				m.srv.fakeBooksControl("3", "0", "R-2-13")
				m.srv.fakeBooksControl("3", "0", "R-2-14")
				m.srv.fakeBooksControl("3", "0", "R-2-15")
				m.srv.fakeBooksControl("3", "0", "R-2-16")
				m.srv.fakeBooksControl("3", "0", "R-2-17")
				m.srv.fakeBooksControl("3", "0", "R-2-18")
				m.srv.fakeBooksControl("3", "0", "R-2-19")
				m.srv.fakeBooksControl("3", "0", "R-2-20")
				m.srv.fakeBooksControl("3", "0", "R-2-21")
				m.srv.fakeBooksControl("3", "0", "R-2-22")
				m.srv.fakeBooksControl("3", "0", "R-2-23")
				m.library.InAnimation = false
				m.library.CurrentFakeBookLight = 15
				m.library.Table.MarkAngle = m.library.Table.CurrentAngle
				m.library.Step = 2
				log.Println("fakebooks right!magic table launch!")
			}
		}
	}

}

func (m *Match) fakeBooksErrorAnimation(dt time.Duration) {
	//需要delay一个时间，为了方便第5本书亮起来
	sec := dt.Seconds()
	if !m.library.InAnimation {
		m.library.FakeAnimationTime = m.opt.FakeAnimationTime / 1000
		m.library.InAnimation = true
	} else {
		m.library.FakeAnimationTime = math.Max(m.library.FakeAnimationTime-sec, 0)
		m.library.CurrentFakeBookLight = 0
		for k, v := range m.library.FakeBooks {
			if v {
				switch k {
				case 1:
					m.srv.fakeBooksControl("1", "2", "R-2-9")
				case 2:
					m.srv.fakeBooksControl("1", "2", "R-2-10")
				case 3:
					m.srv.fakeBooksControl("1", "2", "R-2-11")
				case 4:
					m.srv.fakeBooksControl("1", "2", "R-2-12")
				case 5:
					m.srv.fakeBooksControl("1", "2", "R-2-13")
				case 6:
					m.srv.fakeBooksControl("1", "2", "R-2-14")
				case 7:
					m.srv.fakeBooksControl("1", "2", "R-2-15")
				case 8:
					m.srv.fakeBooksControl("1", "2", "R-2-16")
				case 9:
					m.srv.fakeBooksControl("1", "2", "R-2-17")
				case 10:
					m.srv.fakeBooksControl("1", "2", "R-2-18")
				case 11:
					m.srv.fakeBooksControl("1", "2", "R-2-19")
				case 12:
					m.srv.fakeBooksControl("1", "2", "R-2-20")
				case 13:
					m.srv.fakeBooksControl("1", "2", "R-2-21")
				case 14:
					m.srv.fakeBooksControl("1", "2", "R-2-22")
				case 15:
					m.srv.fakeBooksControl("1", "2", "R-2-23")
				}
			}
		}
		for i := 1; i < 16; i++ {
			m.library.FakeBooks[i] = false
		}
		m.library.InAnimation = false
		log.Println("fakebooks error!")
	}
}

func (m *Match) dealAngle(angle float64) {
	if !m.library.Table.IsUseful || m.library.Table.IsDestroyed {
		return
	}
	//day:2  moon:3
	candles1 := make([]map[string]string, 3)
	candles2 := make([]map[string]string, 3)
	candles3 := make([]map[string]string, 3)
	candles4 := make([]map[string]string, 3)
	if angle > 0 && angle < 30 {
		candles1[0] = map[string]string{"candle": "0", "color": "3"}
		candles1[1] = map[string]string{"candle": "1", "color": "2"}
		candles1[2] = map[string]string{"candle": "2", "color": "2"}

		candles2[0] = map[string]string{"candle": "0", "color": "2"}
		candles2[1] = map[string]string{"candle": "1", "color": "2"}
		candles2[2] = map[string]string{"candle": "2", "color": "2"}

		candles3[0] = map[string]string{"candle": "0", "color": "2"}
		candles3[1] = map[string]string{"candle": "1", "color": "3"}
		candles3[2] = map[string]string{"candle": "2", "color": "3"}

		candles4[0] = map[string]string{"candle": "0", "color": "3"}
		candles4[1] = map[string]string{"candle": "1", "color": "3"}
		candles4[2] = map[string]string{"candle": "2", "color": "3"}
	} else if angle > 30 && angle < 60 {
		candles1[0] = map[string]string{"candle": "0", "color": "3"}
		candles1[1] = map[string]string{"candle": "1", "color": "3"}
		candles1[2] = map[string]string{"candle": "2", "color": "2"}

		candles2[0] = map[string]string{"candle": "0", "color": "2"}
		candles2[1] = map[string]string{"candle": "1", "color": "2"}
		candles2[2] = map[string]string{"candle": "2", "color": "2"}

		candles3[0] = map[string]string{"candle": "0", "color": "2"}
		candles3[1] = map[string]string{"candle": "1", "color": "2"}
		candles3[2] = map[string]string{"candle": "2", "color": "3"}

		candles4[0] = map[string]string{"candle": "0", "color": "3"}
		candles4[1] = map[string]string{"candle": "1", "color": "3"}
		candles4[2] = map[string]string{"candle": "2", "color": "3"}
	} else if angle > 60 && angle < 90 {
		candles1[0] = map[string]string{"candle": "0", "color": "3"}
		candles1[1] = map[string]string{"candle": "1", "color": "3"}
		candles1[2] = map[string]string{"candle": "2", "color": "3"}

		candles2[0] = map[string]string{"candle": "0", "color": "2"}
		candles2[1] = map[string]string{"candle": "1", "color": "2"}
		candles2[2] = map[string]string{"candle": "2", "color": "2"}

		candles3[0] = map[string]string{"candle": "0", "color": "2"}
		candles3[1] = map[string]string{"candle": "1", "color": "2"}
		candles3[2] = map[string]string{"candle": "2", "color": "2"}

		candles4[0] = map[string]string{"candle": "0", "color": "3"}
		candles4[1] = map[string]string{"candle": "1", "color": "3"}
		candles4[2] = map[string]string{"candle": "2", "color": "3"}
	} else if angle > 90 && angle < 120 {
		candles1[0] = map[string]string{"candle": "0", "color": "3"}
		candles1[1] = map[string]string{"candle": "1", "color": "3"}
		candles1[2] = map[string]string{"candle": "2", "color": "3"}

		candles2[0] = map[string]string{"candle": "0", "color": "3"}
		candles2[1] = map[string]string{"candle": "1", "color": "2"}
		candles2[2] = map[string]string{"candle": "2", "color": "2"}

		candles3[0] = map[string]string{"candle": "0", "color": "2"}
		candles3[1] = map[string]string{"candle": "1", "color": "2"}
		candles3[2] = map[string]string{"candle": "2", "color": "2"}

		candles4[0] = map[string]string{"candle": "0", "color": "2"}
		candles4[1] = map[string]string{"candle": "1", "color": "3"}
		candles4[2] = map[string]string{"candle": "2", "color": "3"}
	} else if angle > 120 && angle < 150 {
		candles1[0] = map[string]string{"candle": "0", "color": "3"}
		candles1[1] = map[string]string{"candle": "1", "color": "3"}
		candles1[2] = map[string]string{"candle": "2", "color": "3"}

		candles2[0] = map[string]string{"candle": "0", "color": "3"}
		candles2[1] = map[string]string{"candle": "1", "color": "3"}
		candles2[2] = map[string]string{"candle": "2", "color": "2"}

		candles3[0] = map[string]string{"candle": "0", "color": "2"}
		candles3[1] = map[string]string{"candle": "1", "color": "2"}
		candles3[2] = map[string]string{"candle": "2", "color": "2"}

		candles4[0] = map[string]string{"candle": "0", "color": "2"}
		candles4[1] = map[string]string{"candle": "1", "color": "2"}
		candles4[2] = map[string]string{"candle": "2", "color": "3"}
	} else if angle > 150 && angle < 180 {
		candles1[0] = map[string]string{"candle": "0", "color": "3"}
		candles1[1] = map[string]string{"candle": "1", "color": "3"}
		candles1[2] = map[string]string{"candle": "2", "color": "3"}

		candles2[0] = map[string]string{"candle": "0", "color": "3"}
		candles2[1] = map[string]string{"candle": "1", "color": "3"}
		candles2[2] = map[string]string{"candle": "2", "color": "3"}

		candles3[0] = map[string]string{"candle": "0", "color": "2"}
		candles3[1] = map[string]string{"candle": "1", "color": "2"}
		candles3[2] = map[string]string{"candle": "2", "color": "2"}

		candles4[0] = map[string]string{"candle": "0", "color": "2"}
		candles4[1] = map[string]string{"candle": "1", "color": "2"}
		candles4[2] = map[string]string{"candle": "2", "color": "2"}
	} else if angle > 180 && angle < 210 {
		candles1[0] = map[string]string{"candle": "0", "color": "2"}
		candles1[1] = map[string]string{"candle": "1", "color": "3"}
		candles1[2] = map[string]string{"candle": "2", "color": "3"}

		candles2[0] = map[string]string{"candle": "0", "color": "3"}
		candles2[1] = map[string]string{"candle": "1", "color": "3"}
		candles2[2] = map[string]string{"candle": "2", "color": "3"}

		candles3[0] = map[string]string{"candle": "0", "color": "3"}
		candles3[1] = map[string]string{"candle": "1", "color": "2"}
		candles3[2] = map[string]string{"candle": "2", "color": "2"}

		candles4[0] = map[string]string{"candle": "0", "color": "2"}
		candles4[1] = map[string]string{"candle": "1", "color": "2"}
		candles4[2] = map[string]string{"candle": "2", "color": "2"}
	} else if angle > 210 && angle < 240 {
		candles1[0] = map[string]string{"candle": "0", "color": "2"}
		candles1[1] = map[string]string{"candle": "1", "color": "2"}
		candles1[2] = map[string]string{"candle": "2", "color": "3"}

		candles2[0] = map[string]string{"candle": "0", "color": "3"}
		candles2[1] = map[string]string{"candle": "1", "color": "3"}
		candles2[2] = map[string]string{"candle": "2", "color": "3"}

		candles3[0] = map[string]string{"candle": "0", "color": "3"}
		candles3[1] = map[string]string{"candle": "1", "color": "3"}
		candles3[2] = map[string]string{"candle": "2", "color": "2"}

		candles4[0] = map[string]string{"candle": "0", "color": "2"}
		candles4[1] = map[string]string{"candle": "1", "color": "2"}
		candles4[2] = map[string]string{"candle": "2", "color": "2"}
	} else if angle > 240 && angle < 270 {
		candles1[0] = map[string]string{"candle": "0", "color": "2"}
		candles1[1] = map[string]string{"candle": "1", "color": "2"}
		candles1[2] = map[string]string{"candle": "2", "color": "2"}

		candles2[0] = map[string]string{"candle": "0", "color": "3"}
		candles2[1] = map[string]string{"candle": "1", "color": "3"}
		candles2[2] = map[string]string{"candle": "2", "color": "3"}

		candles3[0] = map[string]string{"candle": "0", "color": "3"}
		candles3[1] = map[string]string{"candle": "1", "color": "3"}
		candles3[2] = map[string]string{"candle": "2", "color": "3"}

		candles4[0] = map[string]string{"candle": "0", "color": "2"}
		candles4[1] = map[string]string{"candle": "1", "color": "2"}
		candles4[2] = map[string]string{"candle": "2", "color": "2"}
	} else if angle > 270 && angle < 300 {
		candles1[0] = map[string]string{"candle": "0", "color": "2"}
		candles1[1] = map[string]string{"candle": "1", "color": "2"}
		candles1[2] = map[string]string{"candle": "2", "color": "2"}

		candles2[0] = map[string]string{"candle": "0", "color": "2"}
		candles2[1] = map[string]string{"candle": "1", "color": "3"}
		candles2[2] = map[string]string{"candle": "2", "color": "3"}

		candles3[0] = map[string]string{"candle": "0", "color": "3"}
		candles3[1] = map[string]string{"candle": "1", "color": "3"}
		candles3[2] = map[string]string{"candle": "2", "color": "3"}

		candles4[0] = map[string]string{"candle": "0", "color": "3"}
		candles4[1] = map[string]string{"candle": "1", "color": "2"}
		candles4[2] = map[string]string{"candle": "2", "color": "2"}
	} else if angle > 300 && angle < 330 {
		candles1[0] = map[string]string{"candle": "0", "color": "2"}
		candles1[1] = map[string]string{"candle": "1", "color": "2"}
		candles1[2] = map[string]string{"candle": "2", "color": "2"}

		candles2[0] = map[string]string{"candle": "0", "color": "2"}
		candles2[1] = map[string]string{"candle": "1", "color": "2"}
		candles2[2] = map[string]string{"candle": "2", "color": "3"}

		candles3[0] = map[string]string{"candle": "0", "color": "3"}
		candles3[1] = map[string]string{"candle": "1", "color": "3"}
		candles3[2] = map[string]string{"candle": "2", "color": "3"}

		candles4[0] = map[string]string{"candle": "0", "color": "3"}
		candles4[1] = map[string]string{"candle": "1", "color": "3"}
		candles4[2] = map[string]string{"candle": "2", "color": "2"}
	} else if angle > 330 && angle < 360 {
		candles1[0] = map[string]string{"candle": "0", "color": "2"}
		candles1[1] = map[string]string{"candle": "1", "color": "2"}
		candles1[2] = map[string]string{"candle": "2", "color": "2"}

		candles2[0] = map[string]string{"candle": "0", "color": "2"}
		candles2[1] = map[string]string{"candle": "1", "color": "2"}
		candles2[2] = map[string]string{"candle": "2", "color": "2"}

		candles3[0] = map[string]string{"candle": "0", "color": "3"}
		candles3[1] = map[string]string{"candle": "1", "color": "3"}
		candles3[2] = map[string]string{"candle": "2", "color": "3"}

		candles4[0] = map[string]string{"candle": "0", "color": "3"}
		candles4[1] = map[string]string{"candle": "1", "color": "3"}
		candles4[2] = map[string]string{"candle": "2", "color": "3"}
	}
	m.srv.candlesControl(candles1, "R-2-2")
	m.srv.candlesControl(candles2, "R-2-3")
	m.srv.candlesControl(candles3, "R-2-4")
	m.srv.candlesControl(candles4, "R-2-5")
	m.library.Table.CurrentAngle = angle
}

func (m *Match) ensureFakeBooks() bool {
	for _, v := range m.opt.FakeBooks {
		if !m.library.FakeBooks[v] {
			return false
		}
	}
	return true
}

//room3
func (m *Match) ensureCandlesPoweron() bool {
	for _, v := range m.stairRoom.Candles {
		if v == 0 {
			return false
		}
	}
	return true
}

//确定当前蜡烛是否都变色，共6盏蜡烛
func (m *Match) ensureCandlesColorNum() bool {
	num := 0
	for _, v := range m.stairRoom.Candles {
		if v != 0 && v != 1 {
			num++
		}
	}
	if num > 5 {
		return true
	} else {
		return false
	}
}

func (m *Match) ensureCandlesColor() bool {
	for k, v := range m.opt.CandlesColor {
		if m.stairRoom.Candles[k] != v {
			return false
		}
	}
	return true
}

func (m *Match) ensureMagicStandsPowerOn() bool {
	for _, v := range m.magicLab.Stands {
		if !v.IsPowerOn {
			return false
		}
	}
	return true
}

func (m *Match) ensureMagicStandsPowerFul() bool {
	for _, v := range m.magicLab.Stands {
		if !v.IsPowerful {
			return false
		}
	}
	return true
}

//法阵被充能和泄能的动画
func (m *Match) poweringAnimation() {

}

func (m *Match) powerDownAnimation() {

}

//room5
func (m *Match) ensureConstellationSymbol() bool {
	i := 0
	for _, v := range m.starTower.ConstellationSymbol {
		if v {
			i++
		}
	}
	if i != 5 {
		return false
	}
	for _, v := range m.opt.Constellations {
		if !m.starTower.ConstellationSymbol[v] {
			return false
		}
	}
	return true
}

func (m *Match) starControl(starNum int, isOpen bool) { //TODO
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("star_led")
	lights := make([]map[string]string, 0)
	leds := make([]map[string]string, 0)
	if isOpen {
		m.starTower.InAnimation = true
		switch starNum {
		case 1:
			m.starTower.ConstellationLed[2] = 2
			m.starTower.ConstellationLed[6] = 2
			m.starTower.ConstellationLed[8] = 2
			m.starTower.ConstellationLed[9] = 2
			m.starTower.ConstellationLight[9] = 1
			m.starTower.ConstellationLight[8] = 1
			m.starTower.ConstellationLight[10] = 1
			m.starTower.ConstellationLight[19] = 1
			leds = append(leds,
				map[string]string{"led_n": "2", "mode": "2"},
				map[string]string{"led_n": "6", "mode": "2"},
				map[string]string{"led_n": "8", "mode": "2"},
				map[string]string{"led_n": "9", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "9", "status": "1"},
				map[string]string{"light_n": "8", "status": "1"},
				map[string]string{"light_n": "10", "status": "1"},
				map[string]string{"light_n": "19", "status": "1"},
			)
		case 2:
			m.starTower.ConstellationLed[14] = 2
			m.starTower.ConstellationLed[15] = 2
			m.starTower.ConstellationLed[16] = 2
			m.starTower.ConstellationLight[21] = 1
			m.starTower.ConstellationLight[22] = 1
			m.starTower.ConstellationLight[23] = 1
			m.starTower.ConstellationLight[24] = 1
			m.starTower.ConstellationLight[25] = 1
			m.starTower.ConstellationLight[26] = 1
			m.starTower.ConstellationLight[31] = 1
			leds = append(leds,
				map[string]string{"led_n": "14", "mode": "2"},
				map[string]string{"led_n": "15", "mode": "2"},
				map[string]string{"led_n": "16", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "21", "status": "1"},
				map[string]string{"light_n": "22", "status": "1"},
				map[string]string{"light_n": "23", "status": "1"},
				map[string]string{"light_n": "24", "status": "1"},
				map[string]string{"light_n": "25", "status": "1"},
				map[string]string{"light_n": "26", "status": "1"},
				map[string]string{"light_n": "31", "status": "1"},
			)
		case 3:
			m.starTower.ConstellationLed[1] = 2
			m.starTower.ConstellationLed[10] = 2
			m.starTower.ConstellationLight[3] = 1
			m.starTower.ConstellationLight[4] = 1
			m.starTower.ConstellationLight[5] = 1
			m.starTower.ConstellationLight[6] = 1
			m.starTower.ConstellationLight[7] = 1
			m.starTower.ConstellationLight[8] = 1
			leds = append(leds,
				map[string]string{"led_n": "1", "mode": "2"},
				map[string]string{"led_n": "10", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "3", "status": "1"},
				map[string]string{"light_n": "4", "status": "1"},
				map[string]string{"light_n": "5", "status": "1"},
				map[string]string{"light_n": "6", "status": "1"},
				map[string]string{"light_n": "7", "status": "1"},
				map[string]string{"light_n": "8", "status": "1"},
			)
		case 4:
			m.starTower.ConstellationLed[26] = 2
			m.starTower.ConstellationLed[25] = 2
			m.starTower.ConstellationLed[24] = 2
			m.starTower.ConstellationLed[23] = 2
			m.starTower.ConstellationLed[22] = 2
			m.starTower.ConstellationLed[21] = 2
			m.starTower.ConstellationLed[20] = 2
			m.starTower.ConstellationLed[12] = 2
			m.starTower.ConstellationLight[18] = 1
			m.starTower.ConstellationLight[19] = 1
			m.starTower.ConstellationLight[20] = 1
			m.starTower.ConstellationLight[31] = 1
			m.starTower.ConstellationLight[32] = 1
			m.starTower.ConstellationLight[33] = 1
			m.starTower.ConstellationLight[34] = 1
			m.starTower.ConstellationLight[30] = 1
			leds = append(leds,
				map[string]string{"led_n": "26", "mode": "2"},
				map[string]string{"led_n": "25", "mode": "2"},
				map[string]string{"led_n": "24", "mode": "2"},
				map[string]string{"led_n": "23", "mode": "2"},
				map[string]string{"led_n": "22", "mode": "2"},
				map[string]string{"led_n": "21", "mode": "2"},
				map[string]string{"led_n": "20", "mode": "2"},
				map[string]string{"led_n": "12", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "18", "status": "1"},
				map[string]string{"light_n": "19", "status": "1"},
				map[string]string{"light_n": "20", "status": "1"},
				map[string]string{"light_n": "31", "status": "1"},
				map[string]string{"light_n": "32", "status": "1"},
				map[string]string{"light_n": "33", "status": "1"},
				map[string]string{"light_n": "34", "status": "1"},
				map[string]string{"light_n": "30", "status": "1"},
			)
		case 5:
			m.starTower.ConstellationLed[27] = 2
			m.starTower.ConstellationLed[28] = 2
			m.starTower.ConstellationLed[25] = 2
			m.starTower.ConstellationLed[30] = 2
			m.starTower.ConstellationLed[29] = 2
			m.starTower.ConstellationLight[10] = 1
			m.starTower.ConstellationLight[18] = 1
			m.starTower.ConstellationLight[16] = 1
			m.starTower.ConstellationLight[17] = 1
			m.starTower.ConstellationLight[34] = 1
			m.starTower.ConstellationLight[35] = 1
			m.starTower.ConstellationLight[36] = 1
			leds = append(leds,
				map[string]string{"led_n": "27", "mode": "2"},
				map[string]string{"led_n": "28", "mode": "2"},
				map[string]string{"led_n": "25", "mode": "2"},
				map[string]string{"led_n": "30", "mode": "2"},
				map[string]string{"led_n": "29", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "10", "status": "1"},
				map[string]string{"light_n": "18", "status": "1"},
				map[string]string{"light_n": "16", "status": "1"},
				map[string]string{"light_n": "17", "status": "1"},
				map[string]string{"light_n": "34", "status": "1"},
				map[string]string{"light_n": "35", "status": "1"},
				map[string]string{"light_n": "36", "status": "1"},
			)
		case 6:
			m.starTower.ConstellationLed[5] = 2
			m.starTower.ConstellationLed[8] = 2
			m.starTower.ConstellationLed[26] = 2
			m.starTower.ConstellationLed[27] = 2
			m.starTower.ConstellationLight[13] = 1
			m.starTower.ConstellationLight[10] = 1
			m.starTower.ConstellationLight[18] = 1
			m.starTower.ConstellationLight[19] = 1
			leds = append(leds,
				map[string]string{"led_n": "5", "mode": "2"},
				map[string]string{"led_n": "8", "mode": "2"},
				map[string]string{"led_n": "26", "mode": "2"},
				map[string]string{"led_n": "27", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "13", "status": "1"},
				map[string]string{"light_n": "10", "status": "1"},
				map[string]string{"light_n": "18", "status": "1"},
				map[string]string{"light_n": "19", "status": "1"},
			)
		case 7:
			m.starTower.ConstellationLed[11] = 2
			m.starTower.ConstellationLed[12] = 2
			m.starTower.ConstellationLed[13] = 2
			m.starTower.ConstellationLed[14] = 2
			m.starTower.ConstellationLed[9] = 2
			m.starTower.ConstellationLight[4] = 1
			m.starTower.ConstellationLight[7] = 1
			m.starTower.ConstellationLight[8] = 1
			m.starTower.ConstellationLight[19] = 1
			m.starTower.ConstellationLight[20] = 1
			m.starTower.ConstellationLight[21] = 1
			m.starTower.ConstellationLight[23] = 1
			m.starTower.ConstellationLight[24] = 1
			leds = append(leds,
				map[string]string{"led_n": "11", "mode": "2"},
				map[string]string{"led_n": "12", "mode": "2"},
				map[string]string{"led_n": "13", "mode": "2"},
				map[string]string{"led_n": "14", "mode": "2"},
				map[string]string{"led_n": "9", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "4", "status": "1"},
				map[string]string{"light_n": "7", "status": "1"},
				map[string]string{"light_n": "8", "status": "1"},
				map[string]string{"light_n": "19", "status": "1"},
				map[string]string{"light_n": "20", "status": "1"},
				map[string]string{"light_n": "21", "status": "1"},
				map[string]string{"light_n": "23", "status": "1"},
				map[string]string{"light_n": "24", "status": "1"},
			)
		case 8:
			m.starTower.ConstellationLed[23] = 2
			m.starTower.ConstellationLed[20] = 2
			m.starTower.ConstellationLed[19] = 2
			m.starTower.ConstellationLed[18] = 2
			m.starTower.ConstellationLight[33] = 1
			m.starTower.ConstellationLight[30] = 1
			m.starTower.ConstellationLight[29] = 1
			m.starTower.ConstellationLight[27] = 1
			m.starTower.ConstellationLight[31] = 1
			leds = append(leds,
				map[string]string{"led_n": "23", "mode": "2"},
				map[string]string{"led_n": "20", "mode": "2"},
				map[string]string{"led_n": "19", "mode": "2"},
				map[string]string{"led_n": "18", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "33", "status": "1"},
				map[string]string{"light_n": "30", "status": "1"},
				map[string]string{"light_n": "29", "status": "1"},
				map[string]string{"light_n": "27", "status": "1"},
				map[string]string{"light_n": "31", "status": "1"},
			)
		case 9:
			m.starTower.ConstellationLed[0] = 2
			m.starTower.ConstellationLed[1] = 2
			m.starTower.ConstellationLed[2] = 2
			m.starTower.ConstellationLed[3] = 2
			m.starTower.ConstellationLight[0] = 1
			m.starTower.ConstellationLight[1] = 1
			m.starTower.ConstellationLight[2] = 1
			m.starTower.ConstellationLight[3] = 1
			m.starTower.ConstellationLight[8] = 1
			m.starTower.ConstellationLight[9] = 1
			leds = append(leds,
				map[string]string{"led_n": "0", "mode": "2"},
				map[string]string{"led_n": "1", "mode": "2"},
				map[string]string{"led_n": "2", "mode": "2"},
				map[string]string{"led_n": "3", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "0", "status": "1"},
				map[string]string{"light_n": "1", "status": "1"},
				map[string]string{"light_n": "2", "status": "1"},
				map[string]string{"light_n": "3", "status": "1"},
				map[string]string{"light_n": "8", "status": "1"},
				map[string]string{"light_n": "9", "status": "1"},
			)
		case 10:
			m.starTower.ConstellationLed[3] = 2
			m.starTower.ConstellationLed[5] = 2
			m.starTower.ConstellationLed[6] = 2
			m.starTower.ConstellationLed[7] = 2
			m.starTower.ConstellationLed[27] = 2
			m.starTower.ConstellationLed[25] = 2
			m.starTower.ConstellationLight[0] = 1
			m.starTower.ConstellationLight[9] = 1
			m.starTower.ConstellationLight[10] = 1
			m.starTower.ConstellationLight[11] = 1
			m.starTower.ConstellationLight[13] = 1
			m.starTower.ConstellationLight[18] = 1
			m.starTower.ConstellationLight[34] = 1
			leds = append(leds,
				map[string]string{"led_n": "3", "mode": "2"},
				map[string]string{"led_n": "5", "mode": "2"},
				map[string]string{"led_n": "6", "mode": "2"},
				map[string]string{"led_n": "7", "mode": "2"},
				map[string]string{"led_n": "27", "mode": "2"},
				map[string]string{"led_n": "25", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "0", "status": "1"},
				map[string]string{"light_n": "9", "status": "1"},
				map[string]string{"light_n": "10", "status": "1"},
				map[string]string{"light_n": "11", "status": "1"},
				map[string]string{"light_n": "13", "status": "1"},
				map[string]string{"light_n": "18", "status": "1"},
				map[string]string{"light_n": "34", "status": "1"},
			)
		case 11:
			m.starTower.ConstellationLed[29] = 2
			m.starTower.ConstellationLed[31] = 2
			m.starTower.ConstellationLed[32] = 2
			m.starTower.ConstellationLed[19] = 2
			m.starTower.ConstellationLight[16] = 1
			m.starTower.ConstellationLight[17] = 1
			m.starTower.ConstellationLight[35] = 1
			m.starTower.ConstellationLight[36] = 1
			m.starTower.ConstellationLight[30] = 1
			m.starTower.ConstellationLight[29] = 1
			leds = append(leds,
				map[string]string{"led_n": "29", "mode": "2"},
				map[string]string{"led_n": "31", "mode": "2"},
				map[string]string{"led_n": "32", "mode": "2"},
				map[string]string{"led_n": "19", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "16", "status": "1"},
				map[string]string{"light_n": "17", "status": "1"},
				map[string]string{"light_n": "35", "status": "1"},
				map[string]string{"light_n": "36", "status": "1"},
				map[string]string{"light_n": "30", "status": "1"},
				map[string]string{"light_n": "29", "status": "1"},
			)
		case 12:
			m.starTower.ConstellationLed[24] = 2
			m.starTower.ConstellationLed[23] = 2
			m.starTower.ConstellationLed[30] = 2
			m.starTower.ConstellationLed[31] = 2
			m.starTower.ConstellationLight[33] = 1
			m.starTower.ConstellationLight[34] = 1
			m.starTower.ConstellationLight[35] = 1
			m.starTower.ConstellationLight[30] = 1
			leds = append(leds,
				map[string]string{"led_n": "24", "mode": "2"},
				map[string]string{"led_n": "23", "mode": "2"},
				map[string]string{"led_n": "30", "mode": "2"},
				map[string]string{"led_n": "31", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "33", "status": "1"},
				map[string]string{"light_n": "34", "status": "1"},
				map[string]string{"light_n": "35", "status": "1"},
				map[string]string{"light_n": "30", "status": "1"},
			)
		case 13:
			m.starTower.ConstellationLed[4] = 2
			m.starTower.ConstellationLight[0] = 1
			m.starTower.ConstellationLight[12] = 1
			m.starTower.ConstellationLight[13] = 1
			m.starTower.ConstellationLight[14] = 1
			m.starTower.ConstellationLight[15] = 1
			m.starTower.ConstellationLight[16] = 1
			leds = append(leds,
				map[string]string{"led_n": "4", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "0", "status": "1"},
				map[string]string{"light_n": "12", "status": "1"},
				map[string]string{"light_n": "13", "status": "1"},
				map[string]string{"light_n": "14", "status": "1"},
				map[string]string{"light_n": "15", "status": "1"},
				map[string]string{"light_n": "16", "status": "1"},
			)
		case 14:
			m.starTower.ConstellationLed[17] = 2
			m.starTower.ConstellationLed[18] = 2
			m.starTower.ConstellationLight[26] = 1
			m.starTower.ConstellationLight[27] = 1
			m.starTower.ConstellationLight[28] = 1
			m.starTower.ConstellationLight[29] = 1
			leds = append(leds,
				map[string]string{"led_n": "17", "mode": "2"},
				map[string]string{"led_n": "18", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "26", "status": "1"},
				map[string]string{"light_n": "27", "status": "1"},
				map[string]string{"light_n": "28", "status": "1"},
				map[string]string{"light_n": "29", "status": "1"},
			)
		case 15:
			m.starTower.ConstellationLed[13] = 2
			m.starTower.ConstellationLed[21] = 2
			m.starTower.ConstellationLed[16] = 2
			m.starTower.ConstellationLight[20] = 1
			m.starTower.ConstellationLight[21] = 1
			m.starTower.ConstellationLight[22] = 1
			m.starTower.ConstellationLight[25] = 1
			m.starTower.ConstellationLight[31] = 1
			m.starTower.ConstellationLight[32] = 1
			leds = append(leds,
				map[string]string{"led_n": "13", "mode": "2"},
				map[string]string{"led_n": "21", "mode": "2"},
				map[string]string{"led_n": "16", "mode": "2"},
			)
			lights = append(lights,
				map[string]string{"light_n": "20", "status": "1"},
				map[string]string{"light_n": "21", "status": "1"},
				map[string]string{"light_n": "22", "status": "1"},
				map[string]string{"light_n": "25", "status": "1"},
				map[string]string{"light_n": "31", "status": "1"},
				map[string]string{"light_n": "32", "status": "1"},
			)
		}
		sendMsg.Set("light", lights)
		sendMsg.Set("led", leds)
		addrs := []InboxAddress{
			{InboxAddressTypeRoomArduinoDevice, "R-5-1"},
			{InboxAddressTypeRoomArduinoDevice, "R-5-2"},
			{InboxAddressTypeRoomArduinoDevice, "R-5-3"},
			{InboxAddressTypeRoomArduinoDevice, "R-5-4"},
			{InboxAddressTypeRoomArduinoDevice, "R-5-5"}}
		m.srv.send(sendMsg, addrs)
		m.starTower.InAnimation = false
	} else {
		m.starTower.InAnimation = true
		switch starNum {
		case 1:
			m.starTower.ConstellationLed[2] = 1
			m.starTower.ConstellationLed[6] = 1
			m.starTower.ConstellationLed[8] = 1
			m.starTower.ConstellationLed[9] = 1
			m.starTower.ConstellationLight[9] = 0
			m.starTower.ConstellationLight[8] = 0
			m.starTower.ConstellationLight[10] = 0
			m.starTower.ConstellationLight[19] = 0
			leds = append(leds,
				map[string]string{"led_n": "2", "mode": "1"},
				map[string]string{"led_n": "6", "mode": "1"},
				map[string]string{"led_n": "8", "mode": "1"},
				map[string]string{"led_n": "9", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "9", "status": "0"},
				map[string]string{"light_n": "8", "status": "0"},
				map[string]string{"light_n": "10", "status": "0"},
				map[string]string{"light_n": "19", "status": "0"},
			)
		case 2:
			m.starTower.ConstellationLed[14] = 1
			m.starTower.ConstellationLed[15] = 1
			m.starTower.ConstellationLed[16] = 1
			m.starTower.ConstellationLight[21] = 0
			m.starTower.ConstellationLight[22] = 0
			m.starTower.ConstellationLight[23] = 0
			m.starTower.ConstellationLight[24] = 0
			m.starTower.ConstellationLight[25] = 0
			m.starTower.ConstellationLight[26] = 0
			m.starTower.ConstellationLight[31] = 0
			leds = append(leds,
				map[string]string{"led_n": "14", "mode": "1"},
				map[string]string{"led_n": "15", "mode": "1"},
				map[string]string{"led_n": "16", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "21", "status": "0"},
				map[string]string{"light_n": "22", "status": "0"},
				map[string]string{"light_n": "23", "status": "0"},
				map[string]string{"light_n": "24", "status": "0"},
				map[string]string{"light_n": "25", "status": "0"},
				map[string]string{"light_n": "26", "status": "0"},
				map[string]string{"light_n": "31", "status": "0"},
			)
		case 3:
			m.starTower.ConstellationLed[1] = 1
			m.starTower.ConstellationLed[10] = 1
			m.starTower.ConstellationLight[3] = 0
			m.starTower.ConstellationLight[4] = 0
			m.starTower.ConstellationLight[5] = 0
			m.starTower.ConstellationLight[6] = 0
			m.starTower.ConstellationLight[7] = 0
			m.starTower.ConstellationLight[8] = 0
			leds = append(leds,
				map[string]string{"led_n": "1", "mode": "1"},
				map[string]string{"led_n": "10", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "3", "status": "0"},
				map[string]string{"light_n": "4", "status": "0"},
				map[string]string{"light_n": "5", "status": "0"},
				map[string]string{"light_n": "6", "status": "0"},
				map[string]string{"light_n": "7", "status": "0"},
				map[string]string{"light_n": "8", "status": "0"},
			)
		case 4:
			m.starTower.ConstellationLed[26] = 1
			m.starTower.ConstellationLed[25] = 1
			m.starTower.ConstellationLed[24] = 1
			m.starTower.ConstellationLed[23] = 1
			m.starTower.ConstellationLed[22] = 1
			m.starTower.ConstellationLed[21] = 1
			m.starTower.ConstellationLed[20] = 1
			m.starTower.ConstellationLed[12] = 1
			m.starTower.ConstellationLight[18] = 0
			m.starTower.ConstellationLight[19] = 0
			m.starTower.ConstellationLight[20] = 0
			m.starTower.ConstellationLight[31] = 0
			m.starTower.ConstellationLight[32] = 0
			m.starTower.ConstellationLight[33] = 0
			m.starTower.ConstellationLight[34] = 0
			m.starTower.ConstellationLight[30] = 0
			leds = append(leds,
				map[string]string{"led_n": "26", "mode": "1"},
				map[string]string{"led_n": "25", "mode": "1"},
				map[string]string{"led_n": "24", "mode": "1"},
				map[string]string{"led_n": "23", "mode": "1"},
				map[string]string{"led_n": "22", "mode": "1"},
				map[string]string{"led_n": "21", "mode": "1"},
				map[string]string{"led_n": "20", "mode": "1"},
				map[string]string{"led_n": "12", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "18", "status": "0"},
				map[string]string{"light_n": "19", "status": "0"},
				map[string]string{"light_n": "20", "status": "0"},
				map[string]string{"light_n": "31", "status": "0"},
				map[string]string{"light_n": "32", "status": "0"},
				map[string]string{"light_n": "33", "status": "0"},
				map[string]string{"light_n": "34", "status": "0"},
				map[string]string{"light_n": "30", "status": "0"},
			)
		case 5:
			m.starTower.ConstellationLed[27] = 1
			m.starTower.ConstellationLed[28] = 1
			m.starTower.ConstellationLed[25] = 1
			m.starTower.ConstellationLed[30] = 1
			m.starTower.ConstellationLed[29] = 1
			m.starTower.ConstellationLight[10] = 0
			m.starTower.ConstellationLight[18] = 0
			m.starTower.ConstellationLight[16] = 0
			m.starTower.ConstellationLight[17] = 0
			m.starTower.ConstellationLight[34] = 0
			m.starTower.ConstellationLight[35] = 0
			m.starTower.ConstellationLight[36] = 0
			leds = append(leds,
				map[string]string{"led_n": "27", "mode": "1"},
				map[string]string{"led_n": "28", "mode": "1"},
				map[string]string{"led_n": "25", "mode": "1"},
				map[string]string{"led_n": "30", "mode": "1"},
				map[string]string{"led_n": "29", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "10", "status": "0"},
				map[string]string{"light_n": "18", "status": "0"},
				map[string]string{"light_n": "16", "status": "0"},
				map[string]string{"light_n": "17", "status": "0"},
				map[string]string{"light_n": "34", "status": "0"},
				map[string]string{"light_n": "35", "status": "0"},
				map[string]string{"light_n": "36", "status": "0"},
			)
		case 6:
			m.starTower.ConstellationLed[5] = 1
			m.starTower.ConstellationLed[8] = 1
			m.starTower.ConstellationLed[26] = 1
			m.starTower.ConstellationLed[27] = 1
			m.starTower.ConstellationLight[13] = 0
			m.starTower.ConstellationLight[10] = 0
			m.starTower.ConstellationLight[18] = 0
			m.starTower.ConstellationLight[19] = 0
			leds = append(leds,
				map[string]string{"led_n": "5", "mode": "1"},
				map[string]string{"led_n": "8", "mode": "1"},
				map[string]string{"led_n": "26", "mode": "1"},
				map[string]string{"led_n": "27", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "13", "status": "0"},
				map[string]string{"light_n": "10", "status": "0"},
				map[string]string{"light_n": "18", "status": "0"},
				map[string]string{"light_n": "19", "status": "0"},
			)
		case 7:
			m.starTower.ConstellationLed[11] = 1
			m.starTower.ConstellationLed[12] = 1
			m.starTower.ConstellationLed[13] = 1
			m.starTower.ConstellationLed[14] = 1
			m.starTower.ConstellationLed[9] = 1
			m.starTower.ConstellationLight[4] = 0
			m.starTower.ConstellationLight[7] = 0
			m.starTower.ConstellationLight[8] = 0
			m.starTower.ConstellationLight[19] = 0
			m.starTower.ConstellationLight[20] = 0
			m.starTower.ConstellationLight[21] = 0
			m.starTower.ConstellationLight[23] = 0
			m.starTower.ConstellationLight[24] = 0
			leds = append(leds,
				map[string]string{"led_n": "11", "mode": "1"},
				map[string]string{"led_n": "12", "mode": "1"},
				map[string]string{"led_n": "13", "mode": "1"},
				map[string]string{"led_n": "14", "mode": "1"},
				map[string]string{"led_n": "9", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "4", "status": "0"},
				map[string]string{"light_n": "7", "status": "0"},
				map[string]string{"light_n": "8", "status": "0"},
				map[string]string{"light_n": "19", "status": "0"},
				map[string]string{"light_n": "20", "status": "0"},
				map[string]string{"light_n": "21", "status": "0"},
				map[string]string{"light_n": "23", "status": "0"},
				map[string]string{"light_n": "24", "status": "0"},
			)
		case 8:
			m.starTower.ConstellationLed[23] = 1
			m.starTower.ConstellationLed[20] = 1
			m.starTower.ConstellationLed[19] = 1
			m.starTower.ConstellationLed[18] = 1
			m.starTower.ConstellationLight[33] = 0
			m.starTower.ConstellationLight[30] = 0
			m.starTower.ConstellationLight[29] = 0
			m.starTower.ConstellationLight[27] = 0
			m.starTower.ConstellationLight[31] = 0
			leds = append(leds,
				map[string]string{"led_n": "23", "mode": "1"},
				map[string]string{"led_n": "20", "mode": "1"},
				map[string]string{"led_n": "19", "mode": "1"},
				map[string]string{"led_n": "18", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "33", "status": "0"},
				map[string]string{"light_n": "30", "status": "0"},
				map[string]string{"light_n": "29", "status": "0"},
				map[string]string{"light_n": "27", "status": "0"},
				map[string]string{"light_n": "31", "status": "0"},
			)
		case 9:
			m.starTower.ConstellationLed[0] = 1
			m.starTower.ConstellationLed[1] = 1
			m.starTower.ConstellationLed[2] = 1
			m.starTower.ConstellationLed[3] = 1
			m.starTower.ConstellationLight[0] = 0
			m.starTower.ConstellationLight[1] = 0
			m.starTower.ConstellationLight[2] = 0
			m.starTower.ConstellationLight[3] = 0
			m.starTower.ConstellationLight[8] = 0
			m.starTower.ConstellationLight[9] = 0
			leds = append(leds,
				map[string]string{"led_n": "0", "mode": "1"},
				map[string]string{"led_n": "1", "mode": "1"},
				map[string]string{"led_n": "2", "mode": "1"},
				map[string]string{"led_n": "3", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "0", "status": "0"},
				map[string]string{"light_n": "1", "status": "0"},
				map[string]string{"light_n": "2", "status": "0"},
				map[string]string{"light_n": "3", "status": "0"},
				map[string]string{"light_n": "8", "status": "0"},
				map[string]string{"light_n": "9", "status": "0"},
			)
		case 10:
			m.starTower.ConstellationLed[3] = 1
			m.starTower.ConstellationLed[5] = 1
			m.starTower.ConstellationLed[6] = 1
			m.starTower.ConstellationLed[7] = 1
			m.starTower.ConstellationLed[27] = 1
			m.starTower.ConstellationLed[25] = 1
			m.starTower.ConstellationLight[0] = 0
			m.starTower.ConstellationLight[9] = 0
			m.starTower.ConstellationLight[10] = 0
			m.starTower.ConstellationLight[11] = 0
			m.starTower.ConstellationLight[13] = 0
			m.starTower.ConstellationLight[18] = 0
			m.starTower.ConstellationLight[34] = 0
			leds = append(leds,
				map[string]string{"led_n": "3", "mode": "1"},
				map[string]string{"led_n": "5", "mode": "1"},
				map[string]string{"led_n": "6", "mode": "1"},
				map[string]string{"led_n": "7", "mode": "1"},
				map[string]string{"led_n": "27", "mode": "1"},
				map[string]string{"led_n": "25", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "0", "status": "0"},
				map[string]string{"light_n": "9", "status": "0"},
				map[string]string{"light_n": "10", "status": "0"},
				map[string]string{"light_n": "11", "status": "0"},
				map[string]string{"light_n": "13", "status": "0"},
				map[string]string{"light_n": "18", "status": "0"},
				map[string]string{"light_n": "34", "status": "0"},
			)
		case 11:
			m.starTower.ConstellationLed[29] = 1
			m.starTower.ConstellationLed[31] = 1
			m.starTower.ConstellationLed[32] = 1
			m.starTower.ConstellationLed[19] = 1
			m.starTower.ConstellationLight[16] = 0
			m.starTower.ConstellationLight[17] = 0
			m.starTower.ConstellationLight[35] = 0
			m.starTower.ConstellationLight[36] = 0
			m.starTower.ConstellationLight[30] = 0
			m.starTower.ConstellationLight[29] = 0
			leds = append(leds,
				map[string]string{"led_n": "29", "mode": "1"},
				map[string]string{"led_n": "31", "mode": "1"},
				map[string]string{"led_n": "32", "mode": "1"},
				map[string]string{"led_n": "19", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "16", "status": "0"},
				map[string]string{"light_n": "17", "status": "0"},
				map[string]string{"light_n": "35", "status": "0"},
				map[string]string{"light_n": "36", "status": "0"},
				map[string]string{"light_n": "30", "status": "0"},
				map[string]string{"light_n": "29", "status": "0"},
			)
		case 12:
			m.starTower.ConstellationLed[24] = 1
			m.starTower.ConstellationLed[23] = 1
			m.starTower.ConstellationLed[30] = 1
			m.starTower.ConstellationLed[31] = 1
			m.starTower.ConstellationLight[33] = 0
			m.starTower.ConstellationLight[34] = 0
			m.starTower.ConstellationLight[35] = 0
			m.starTower.ConstellationLight[30] = 0
			leds = append(leds,
				map[string]string{"led_n": "24", "mode": "1"},
				map[string]string{"led_n": "23", "mode": "1"},
				map[string]string{"led_n": "30", "mode": "1"},
				map[string]string{"led_n": "31", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "33", "status": "0"},
				map[string]string{"light_n": "34", "status": "0"},
				map[string]string{"light_n": "35", "status": "0"},
				map[string]string{"light_n": "30", "status": "0"},
			)
		case 13:
			m.starTower.ConstellationLed[4] = 1
			m.starTower.ConstellationLight[0] = 0
			m.starTower.ConstellationLight[12] = 0
			m.starTower.ConstellationLight[13] = 0
			m.starTower.ConstellationLight[14] = 0
			m.starTower.ConstellationLight[15] = 0
			m.starTower.ConstellationLight[16] = 0
			leds = append(leds,
				map[string]string{"led_n": "4", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "0", "status": "0"},
				map[string]string{"light_n": "12", "status": "0"},
				map[string]string{"light_n": "13", "status": "0"},
				map[string]string{"light_n": "14", "status": "0"},
				map[string]string{"light_n": "15", "status": "0"},
				map[string]string{"light_n": "16", "status": "0"},
			)
		case 14:
			m.starTower.ConstellationLed[17] = 1
			m.starTower.ConstellationLed[18] = 1
			m.starTower.ConstellationLight[26] = 0
			m.starTower.ConstellationLight[27] = 0
			m.starTower.ConstellationLight[28] = 0
			m.starTower.ConstellationLight[29] = 0
			leds = append(leds,
				map[string]string{"led_n": "17", "mode": "1"},
				map[string]string{"led_n": "18", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "26", "status": "0"},
				map[string]string{"light_n": "27", "status": "0"},
				map[string]string{"light_n": "28", "status": "0"},
				map[string]string{"light_n": "29", "status": "0"},
			)
		case 15:
			m.starTower.ConstellationLed[13] = 1
			m.starTower.ConstellationLed[21] = 1
			m.starTower.ConstellationLed[16] = 1
			m.starTower.ConstellationLight[20] = 0
			m.starTower.ConstellationLight[21] = 0
			m.starTower.ConstellationLight[22] = 0
			m.starTower.ConstellationLight[25] = 0
			m.starTower.ConstellationLight[31] = 0
			m.starTower.ConstellationLight[32] = 0
			leds = append(leds,
				map[string]string{"led_n": "13", "mode": "1"},
				map[string]string{"led_n": "21", "mode": "1"},
				map[string]string{"led_n": "16", "mode": "1"},
			)
			lights = append(lights,
				map[string]string{"light_n": "20", "status": "0"},
				map[string]string{"light_n": "21", "status": "0"},
				map[string]string{"light_n": "22", "status": "0"},
				map[string]string{"light_n": "25", "status": "0"},
				map[string]string{"light_n": "31", "status": "0"},
				map[string]string{"light_n": "32", "status": "0"},
			)
		}
		sendMsg.Set("light", lights)
		sendMsg.Set("led", leds)
		addrs := []InboxAddress{
			{InboxAddressTypeRoomArduinoDevice, "R-5-1"},
			{InboxAddressTypeRoomArduinoDevice, "R-5-2"},
			{InboxAddressTypeRoomArduinoDevice, "R-5-3"},
			{InboxAddressTypeRoomArduinoDevice, "R-5-4"},
			{InboxAddressTypeRoomArduinoDevice, "R-5-5"}}
		m.srv.send(sendMsg, addrs)
		time.Sleep(2 * time.Second)
	}
}

func (m *Match) dealStar(starNum int) {
	switch starNum {
	case 0:
	case 1:
		if m.starTower.ConstellationSymbol["sct"] {
			m.starTower.ConstellationSymbol["sct"] = false
		} else {
			m.starTower.ConstellationSymbol["sct"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["sct"])
	case 2:
		if m.starTower.ConstellationSymbol["vol"] {
			m.starTower.ConstellationSymbol["vol"] = false
		} else {
			m.starTower.ConstellationSymbol["vol"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["vol"])
	case 3:
		if m.starTower.ConstellationSymbol["phe"] {
			m.starTower.ConstellationSymbol["phe"] = false
		} else {
			m.starTower.ConstellationSymbol["phe"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["phe"])
	case 4:
		if m.starTower.ConstellationSymbol["crt"] {
			m.starTower.ConstellationSymbol["crt"] = false
		} else {
			m.starTower.ConstellationSymbol["crt"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["crt"])
	case 5:
		if m.starTower.ConstellationSymbol["can"] {
			m.starTower.ConstellationSymbol["can"] = false
		} else {
			m.starTower.ConstellationSymbol["can"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["can"])
	case 6:
		if m.starTower.ConstellationSymbol["cam"] {
			m.starTower.ConstellationSymbol["cam"] = false
		} else {
			m.starTower.ConstellationSymbol["cam"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["cam"])
	case 7:
		if m.starTower.ConstellationSymbol["boo"] {
			m.starTower.ConstellationSymbol["boo"] = false
		} else {
			m.starTower.ConstellationSymbol["boo"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["boo"])
	case 8:
		if m.starTower.ConstellationSymbol["mon"] {
			m.starTower.ConstellationSymbol["mon"] = false
		} else {
			m.starTower.ConstellationSymbol["mon"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["mon"])
	case 9:
		if m.starTower.ConstellationSymbol["cap"] {
			m.starTower.ConstellationSymbol["cap"] = false
		} else {
			m.starTower.ConstellationSymbol["cap"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["cap"])
	case 10:
		if m.starTower.ConstellationSymbol["gru"] {
			m.starTower.ConstellationSymbol["gru"] = false
		} else {
			m.starTower.ConstellationSymbol["gru"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["gru"])
	case 11:
		if m.starTower.ConstellationSymbol["lyr"] {
			m.starTower.ConstellationSymbol["lyr"] = false
		} else {
			m.starTower.ConstellationSymbol["lyr"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["lyr"])
	case 12:
		if m.starTower.ConstellationSymbol["crv"] {
			m.starTower.ConstellationSymbol["crv"] = false
		} else {
			m.starTower.ConstellationSymbol["crv"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["crv"])
	case 13:
		if m.starTower.ConstellationSymbol["lac"] {
			m.starTower.ConstellationSymbol["lac"] = false
		} else {
			m.starTower.ConstellationSymbol["lac"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["lac"])
	case 14:
		if m.starTower.ConstellationSymbol["leo"] {
			m.starTower.ConstellationSymbol["leo"] = false
		} else {
			m.starTower.ConstellationSymbol["leo"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["leo"])
	case 15:
		if m.starTower.ConstellationSymbol["aur"] {
			m.starTower.ConstellationSymbol["aur"] = false
		} else {
			m.starTower.ConstellationSymbol["aur"] = true
		}
		m.starControl(starNum, m.starTower.ConstellationSymbol["aur"])
	}
	var starName string
	switch starNum {
	case 1:
		starName = "sct"
	case 2:
		starName = "vol"
	case 3:
		starName = "phe"
	case 4:
		starName = "crt"
	case 5:
		starName = "can"
	case 6:
		starName = "cam"
	case 7:
		starName = "boo"
	case 8:
		starName = "mon"
	case 9:
		starName = "cap"
	case 10:
		starName = "gru"
	case 11:
		starName = "lyr"
	case 12:
		starName = "crv"
	case 13:
		starName = "lac"
	case 14:
		starName = "leo"
	case 15:
		starName = "aur"
	}
	if !m.starTower.ConstellationSymbol[starName] {
		for k, v := range m.starTower.ConstellationSymbol {
			if v {
				m.updateStarStatus(k)
			}
		}
	}
	m.starTower.InAnimation = false
}

func (m *Match) updateStarStatus(starName string) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("star_led")
	lights := make([]map[string]string, 0)
	leds := make([]map[string]string, 0)
	switch starName {
	case "sct":
		m.starTower.ConstellationLed[2] = 5
		m.starTower.ConstellationLed[6] = 5
		m.starTower.ConstellationLed[8] = 5
		m.starTower.ConstellationLed[9] = 5
		m.starTower.ConstellationLight[9] = 1
		m.starTower.ConstellationLight[8] = 1
		m.starTower.ConstellationLight[10] = 1
		m.starTower.ConstellationLight[19] = 1
		leds = append(leds,
			map[string]string{"led_n": "2", "mode": "5"},
			map[string]string{"led_n": "6", "mode": "5"},
			map[string]string{"led_n": "8", "mode": "5"},
			map[string]string{"led_n": "9", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "9", "status": "1"},
			map[string]string{"light_n": "8", "status": "1"},
			map[string]string{"light_n": "10", "status": "1"},
			map[string]string{"light_n": "19", "status": "1"},
		)
	case "vol":
		m.starTower.ConstellationLed[14] = 5
		m.starTower.ConstellationLed[15] = 5
		m.starTower.ConstellationLed[16] = 5
		m.starTower.ConstellationLight[21] = 1
		m.starTower.ConstellationLight[22] = 1
		m.starTower.ConstellationLight[23] = 1
		m.starTower.ConstellationLight[24] = 1
		m.starTower.ConstellationLight[25] = 1
		m.starTower.ConstellationLight[26] = 1
		m.starTower.ConstellationLight[31] = 1
		leds = append(leds,
			map[string]string{"led_n": "14", "mode": "5"},
			map[string]string{"led_n": "15", "mode": "5"},
			map[string]string{"led_n": "16", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "21", "status": "1"},
			map[string]string{"light_n": "22", "status": "1"},
			map[string]string{"light_n": "23", "status": "1"},
			map[string]string{"light_n": "24", "status": "1"},
			map[string]string{"light_n": "25", "status": "1"},
			map[string]string{"light_n": "26", "status": "1"},
			map[string]string{"light_n": "31", "status": "1"},
		)
	case "phe":
		m.starTower.ConstellationLed[1] = 5
		m.starTower.ConstellationLed[10] = 5
		m.starTower.ConstellationLight[3] = 1
		m.starTower.ConstellationLight[4] = 1
		m.starTower.ConstellationLight[5] = 1
		m.starTower.ConstellationLight[6] = 1
		m.starTower.ConstellationLight[7] = 1
		m.starTower.ConstellationLight[8] = 1
		leds = append(leds,
			map[string]string{"led_n": "1", "mode": "5"},
			map[string]string{"led_n": "10", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "3", "status": "1"},
			map[string]string{"light_n": "4", "status": "1"},
			map[string]string{"light_n": "5", "status": "1"},
			map[string]string{"light_n": "6", "status": "1"},
			map[string]string{"light_n": "7", "status": "1"},
			map[string]string{"light_n": "8", "status": "1"},
		)
	case "crt":
		m.starTower.ConstellationLed[26] = 5
		m.starTower.ConstellationLed[25] = 5
		m.starTower.ConstellationLed[24] = 5
		m.starTower.ConstellationLed[23] = 5
		m.starTower.ConstellationLed[22] = 5
		m.starTower.ConstellationLed[21] = 5
		m.starTower.ConstellationLed[20] = 5
		m.starTower.ConstellationLed[12] = 5
		m.starTower.ConstellationLight[18] = 1
		m.starTower.ConstellationLight[19] = 1
		m.starTower.ConstellationLight[20] = 1
		m.starTower.ConstellationLight[31] = 1
		m.starTower.ConstellationLight[32] = 1
		m.starTower.ConstellationLight[33] = 1
		m.starTower.ConstellationLight[34] = 1
		m.starTower.ConstellationLight[30] = 1
		leds = append(leds,
			map[string]string{"led_n": "26", "mode": "5"},
			map[string]string{"led_n": "25", "mode": "5"},
			map[string]string{"led_n": "24", "mode": "5"},
			map[string]string{"led_n": "23", "mode": "5"},
			map[string]string{"led_n": "22", "mode": "5"},
			map[string]string{"led_n": "21", "mode": "5"},
			map[string]string{"led_n": "20", "mode": "5"},
			map[string]string{"led_n": "12", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "18", "status": "1"},
			map[string]string{"light_n": "19", "status": "1"},
			map[string]string{"light_n": "20", "status": "1"},
			map[string]string{"light_n": "31", "status": "1"},
			map[string]string{"light_n": "32", "status": "1"},
			map[string]string{"light_n": "33", "status": "1"},
			map[string]string{"light_n": "34", "status": "1"},
			map[string]string{"light_n": "30", "status": "1"},
		)
	case "can":
		m.starTower.ConstellationLed[27] = 5
		m.starTower.ConstellationLed[28] = 5
		m.starTower.ConstellationLed[25] = 5
		m.starTower.ConstellationLed[30] = 5
		m.starTower.ConstellationLed[29] = 5
		m.starTower.ConstellationLight[10] = 1
		m.starTower.ConstellationLight[18] = 1
		m.starTower.ConstellationLight[16] = 1
		m.starTower.ConstellationLight[17] = 1
		m.starTower.ConstellationLight[34] = 1
		m.starTower.ConstellationLight[35] = 1
		m.starTower.ConstellationLight[36] = 1
		leds = append(leds,
			map[string]string{"led_n": "27", "mode": "5"},
			map[string]string{"led_n": "28", "mode": "5"},
			map[string]string{"led_n": "25", "mode": "5"},
			map[string]string{"led_n": "30", "mode": "5"},
			map[string]string{"led_n": "29", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "10", "status": "1"},
			map[string]string{"light_n": "18", "status": "1"},
			map[string]string{"light_n": "16", "status": "1"},
			map[string]string{"light_n": "17", "status": "1"},
			map[string]string{"light_n": "34", "status": "1"},
			map[string]string{"light_n": "35", "status": "1"},
			map[string]string{"light_n": "36", "status": "1"},
		)
	case "cam":
		m.starTower.ConstellationLed[5] = 5
		m.starTower.ConstellationLed[8] = 5
		m.starTower.ConstellationLed[26] = 5
		m.starTower.ConstellationLed[27] = 5
		m.starTower.ConstellationLight[13] = 1
		m.starTower.ConstellationLight[10] = 1
		m.starTower.ConstellationLight[18] = 1
		m.starTower.ConstellationLight[19] = 1
		leds = append(leds,
			map[string]string{"led_n": "5", "mode": "5"},
			map[string]string{"led_n": "8", "mode": "5"},
			map[string]string{"led_n": "26", "mode": "5"},
			map[string]string{"led_n": "27", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "13", "status": "1"},
			map[string]string{"light_n": "10", "status": "1"},
			map[string]string{"light_n": "18", "status": "1"},
			map[string]string{"light_n": "19", "status": "1"},
		)
	case "boo":
		m.starTower.ConstellationLed[11] = 5
		m.starTower.ConstellationLed[12] = 5
		m.starTower.ConstellationLed[13] = 5
		m.starTower.ConstellationLed[14] = 5
		m.starTower.ConstellationLed[9] = 5
		m.starTower.ConstellationLight[4] = 1
		m.starTower.ConstellationLight[7] = 1
		m.starTower.ConstellationLight[8] = 1
		m.starTower.ConstellationLight[19] = 1
		m.starTower.ConstellationLight[20] = 1
		m.starTower.ConstellationLight[21] = 1
		m.starTower.ConstellationLight[23] = 1
		m.starTower.ConstellationLight[24] = 1
		leds = append(leds,
			map[string]string{"led_n": "11", "mode": "5"},
			map[string]string{"led_n": "12", "mode": "5"},
			map[string]string{"led_n": "13", "mode": "5"},
			map[string]string{"led_n": "14", "mode": "5"},
			map[string]string{"led_n": "9", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "4", "status": "1"},
			map[string]string{"light_n": "7", "status": "1"},
			map[string]string{"light_n": "8", "status": "1"},
			map[string]string{"light_n": "19", "status": "1"},
			map[string]string{"light_n": "20", "status": "1"},
			map[string]string{"light_n": "21", "status": "1"},
			map[string]string{"light_n": "23", "status": "1"},
			map[string]string{"light_n": "24", "status": "1"},
		)
	case "mon":
		m.starTower.ConstellationLed[23] = 5
		m.starTower.ConstellationLed[20] = 5
		m.starTower.ConstellationLed[19] = 5
		m.starTower.ConstellationLed[18] = 5
		m.starTower.ConstellationLight[33] = 1
		m.starTower.ConstellationLight[30] = 1
		m.starTower.ConstellationLight[29] = 1
		m.starTower.ConstellationLight[27] = 1
		m.starTower.ConstellationLight[31] = 1
		leds = append(leds,
			map[string]string{"led_n": "23", "mode": "5"},
			map[string]string{"led_n": "20", "mode": "5"},
			map[string]string{"led_n": "19", "mode": "5"},
			map[string]string{"led_n": "18", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "33", "status": "1"},
			map[string]string{"light_n": "30", "status": "1"},
			map[string]string{"light_n": "29", "status": "1"},
			map[string]string{"light_n": "27", "status": "1"},
			map[string]string{"light_n": "31", "status": "1"},
		)
	case "cap":
		m.starTower.ConstellationLed[0] = 5
		m.starTower.ConstellationLed[1] = 5
		m.starTower.ConstellationLed[2] = 5
		m.starTower.ConstellationLed[3] = 5
		m.starTower.ConstellationLight[0] = 1
		m.starTower.ConstellationLight[1] = 1
		m.starTower.ConstellationLight[2] = 1
		m.starTower.ConstellationLight[3] = 1
		m.starTower.ConstellationLight[8] = 1
		m.starTower.ConstellationLight[9] = 1
		leds = append(leds,
			map[string]string{"led_n": "0", "mode": "5"},
			map[string]string{"led_n": "1", "mode": "5"},
			map[string]string{"led_n": "2", "mode": "5"},
			map[string]string{"led_n": "3", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "0", "status": "1"},
			map[string]string{"light_n": "1", "status": "1"},
			map[string]string{"light_n": "2", "status": "1"},
			map[string]string{"light_n": "3", "status": "1"},
			map[string]string{"light_n": "8", "status": "1"},
			map[string]string{"light_n": "9", "status": "1"},
		)
	case "gru":
		m.starTower.ConstellationLed[3] = 5
		m.starTower.ConstellationLed[5] = 5
		m.starTower.ConstellationLed[6] = 5
		m.starTower.ConstellationLed[7] = 5
		m.starTower.ConstellationLed[27] = 5
		m.starTower.ConstellationLed[25] = 5
		m.starTower.ConstellationLight[0] = 1
		m.starTower.ConstellationLight[9] = 1
		m.starTower.ConstellationLight[10] = 1
		m.starTower.ConstellationLight[11] = 1
		m.starTower.ConstellationLight[13] = 1
		m.starTower.ConstellationLight[18] = 1
		m.starTower.ConstellationLight[34] = 1
		leds = append(leds,
			map[string]string{"led_n": "3", "mode": "5"},
			map[string]string{"led_n": "5", "mode": "5"},
			map[string]string{"led_n": "6", "mode": "5"},
			map[string]string{"led_n": "7", "mode": "5"},
			map[string]string{"led_n": "27", "mode": "5"},
			map[string]string{"led_n": "25", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "0", "status": "1"},
			map[string]string{"light_n": "9", "status": "1"},
			map[string]string{"light_n": "10", "status": "1"},
			map[string]string{"light_n": "11", "status": "1"},
			map[string]string{"light_n": "13", "status": "1"},
			map[string]string{"light_n": "18", "status": "1"},
			map[string]string{"light_n": "34", "status": "1"},
		)
	case "lyr":
		m.starTower.ConstellationLed[29] = 5
		m.starTower.ConstellationLed[31] = 5
		m.starTower.ConstellationLed[32] = 5
		m.starTower.ConstellationLed[19] = 5
		m.starTower.ConstellationLight[16] = 1
		m.starTower.ConstellationLight[17] = 1
		m.starTower.ConstellationLight[35] = 1
		m.starTower.ConstellationLight[36] = 1
		m.starTower.ConstellationLight[30] = 1
		m.starTower.ConstellationLight[29] = 1
		leds = append(leds,
			map[string]string{"led_n": "29", "mode": "5"},
			map[string]string{"led_n": "31", "mode": "5"},
			map[string]string{"led_n": "32", "mode": "5"},
			map[string]string{"led_n": "19", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "16", "status": "1"},
			map[string]string{"light_n": "17", "status": "1"},
			map[string]string{"light_n": "35", "status": "1"},
			map[string]string{"light_n": "36", "status": "1"},
			map[string]string{"light_n": "30", "status": "1"},
			map[string]string{"light_n": "29", "status": "1"},
		)
	case "crv":
		m.starTower.ConstellationLed[24] = 5
		m.starTower.ConstellationLed[23] = 5
		m.starTower.ConstellationLed[30] = 5
		m.starTower.ConstellationLed[31] = 5
		m.starTower.ConstellationLight[33] = 1
		m.starTower.ConstellationLight[34] = 1
		m.starTower.ConstellationLight[35] = 1
		m.starTower.ConstellationLight[30] = 1
		leds = append(leds,
			map[string]string{"led_n": "24", "mode": "5"},
			map[string]string{"led_n": "23", "mode": "5"},
			map[string]string{"led_n": "30", "mode": "5"},
			map[string]string{"led_n": "31", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "33", "status": "1"},
			map[string]string{"light_n": "34", "status": "1"},
			map[string]string{"light_n": "35", "status": "1"},
			map[string]string{"light_n": "30", "status": "1"},
		)
	case "lac":
		m.starTower.ConstellationLed[4] = 5
		m.starTower.ConstellationLight[0] = 1
		m.starTower.ConstellationLight[12] = 1
		m.starTower.ConstellationLight[13] = 1
		m.starTower.ConstellationLight[14] = 1
		m.starTower.ConstellationLight[15] = 1
		m.starTower.ConstellationLight[16] = 1
		leds = append(leds,
			map[string]string{"led_n": "4", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "0", "status": "1"},
			map[string]string{"light_n": "12", "status": "1"},
			map[string]string{"light_n": "13", "status": "1"},
			map[string]string{"light_n": "14", "status": "1"},
			map[string]string{"light_n": "15", "status": "1"},
			map[string]string{"light_n": "16", "status": "1"},
		)
	case "leo":
		m.starTower.ConstellationLed[17] = 5
		m.starTower.ConstellationLed[18] = 5
		m.starTower.ConstellationLight[26] = 1
		m.starTower.ConstellationLight[27] = 1
		m.starTower.ConstellationLight[28] = 1
		m.starTower.ConstellationLight[29] = 1
		leds = append(leds,
			map[string]string{"led_n": "17", "mode": "5"},
			map[string]string{"led_n": "18", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "26", "status": "1"},
			map[string]string{"light_n": "27", "status": "1"},
			map[string]string{"light_n": "28", "status": "1"},
			map[string]string{"light_n": "29", "status": "1"},
		)
	case "aur":
		m.starTower.ConstellationLed[13] = 5
		m.starTower.ConstellationLed[21] = 5
		m.starTower.ConstellationLed[16] = 5
		m.starTower.ConstellationLight[20] = 1
		m.starTower.ConstellationLight[21] = 1
		m.starTower.ConstellationLight[22] = 1
		m.starTower.ConstellationLight[25] = 1
		m.starTower.ConstellationLight[31] = 1
		m.starTower.ConstellationLight[32] = 1
		leds = append(leds,
			map[string]string{"led_n": "13", "mode": "5"},
			map[string]string{"led_n": "21", "mode": "5"},
			map[string]string{"led_n": "16", "mode": "5"},
		)
		lights = append(lights,
			map[string]string{"light_n": "20", "status": "1"},
			map[string]string{"light_n": "21", "status": "1"},
			map[string]string{"light_n": "22", "status": "1"},
			map[string]string{"light_n": "25", "status": "1"},
			map[string]string{"light_n": "31", "status": "1"},
			map[string]string{"light_n": "32", "status": "1"},
		)
	}
	sendMsg.Set("light", lights)
	sendMsg.Set("led", leds)
	addrs := []InboxAddress{
		{InboxAddressTypeRoomArduinoDevice, "R-5-1"},
		{InboxAddressTypeRoomArduinoDevice, "R-5-2"},
		{InboxAddressTypeRoomArduinoDevice, "R-5-3"},
		{InboxAddressTypeRoomArduinoDevice, "R-5-4"},
		{InboxAddressTypeRoomArduinoDevice, "R-5-5"}}
	m.srv.send(sendMsg, addrs)
}

//room6 animation
func (m *Match) amMagicAnimation() {

}

func (m *Match) bgmPlay(bgm int) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("mp3_ctrl")
	sendMsg.Set("music", strconv.Itoa(bgm))
	addr := InboxAddress{InboxAddressTypeMusicArduino, "B-1"}
	m.srv.sendToOne(sendMsg, addr)

}

func (m *Match) broadSymbolToArduino(symbol int) {
	if !m.endRoom.Table.IsUseful || m.endRoom.Table.IsDestroyed {
		return
	}
	m.endRoom.CurrentSymbol = symbol
	addrs := []InboxAddress{
		{InboxAddressTypeRoomArduinoDevice, "R-6-1"},
		{InboxAddressTypeRoomArduinoDevice, "R-6-2"},
		{InboxAddressTypeRoomArduinoDevice, "R-6-3"},
		{InboxAddressTypeRoomArduinoDevice, "R-6-4"},
		{InboxAddressTypeRoomArduinoDevice, "R-6-5"},
		{InboxAddressTypeRoomArduinoDevice, "R-6-6"},
	}
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("power_point")
	sendMsg.Set("mode", "0")
	sendMsg.Set("useful", "1")
	sendMsg.Set("type_power", strconv.Itoa(symbol))
	m.srv.send(sendMsg, addrs)
	log.Println("board symbol:", symbol)
}

func (m *Match) ensureElementSymbol() bool {
	for k, v := range m.opt.ElementSymbol {
		if m.endRoom.PowerPoint[k] != v {
			return false
		}
	}
	return true
}

func (m *Match) magicTableAnimation(s string, dt time.Duration) {
	sec := dt.Seconds()
	switch s {
	case StageRoom2:
	case StageRoom3:
		m.stairRoom.Table.IsUseful = true
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("magic_table")
		sendMsg.Set("useful", "1")
		sendMsg.Set("time", strconv.FormatFloat(GetOptions().FakeAnimationTime, 'f', 0, 64))
		sendMsg.Set("InAnimation", "1")
		addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-3-7"}
		m.srv.sendToOne(sendMsg, addr)
	case StageRoom4:
	case StageRoom5:
	case StageRoom6:
		if !m.endRoom.InAnimation {
			m.endRoom.LaunchDelayTime = m.opt.Room6LaunchDelayTime / 1000
			m.endRoom.LaunchStep = 1
			m.endRoom.InAnimation = true
			log.Println("magic table will launch!")
		}
		m.endRoom.LaunchDelayTime = math.Max(m.endRoom.LaunchDelayTime-sec, 0)
		if m.endRoom.LaunchDelayTime == 0 {
			switch m.endRoom.LaunchStep {
			case 1:
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("useful", "1")
				sendMsg.Set("InAnimation", "1")
				sendMsg.Set("time", strconv.FormatFloat(GetOptions().FakeAnimationTime, 'f', 0, 64))
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg, addr)

				m.endRoom.LightStatus = true
				sendMsg2 := NewInboxMessage()
				sendMsg2.SetCmd("water_light")
				sendMsg2.Set("status", "1")
				addr2 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-9"}
				m.srv.sendToOne(sendMsg2, addr2)

				m.endRoom.PowerPointUseful[0] = 1
				m.endRoom.PowerPointUseful[1] = 1
				m.endRoom.PowerPointUseful[2] = 1
				m.endRoom.PowerPointUseful[3] = 1
				m.endRoom.PowerPointUseful[4] = 1
				m.endRoom.PowerPointUseful[5] = 1
				sendMsg3 := NewInboxMessage()
				sendMsg3.SetCmd("power_point")
				sendMsg3.Set("useful", "1")
				sendMsg3.Set("powerMode", "0")
				sendMsg3.Set("type_power", "0")
				addrs := []InboxAddress{
					{InboxAddressTypeRoomArduinoDevice, "R-6-1"},
					{InboxAddressTypeRoomArduinoDevice, "R-6-2"},
					{InboxAddressTypeRoomArduinoDevice, "R-6-3"},
					{InboxAddressTypeRoomArduinoDevice, "R-6-4"},
					{InboxAddressTypeRoomArduinoDevice, "R-6-5"},
					{InboxAddressTypeRoomArduinoDevice, "R-6-6"},
				}
				m.srv.send(sendMsg3, addrs)
				m.endRoom.LaunchStep = 2
				m.endRoom.LaunchDelayTime = 3
				log.Println("launch table")
			case 2:
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("led_candle")
				sendMsg1.Set("mode", "1")
				candles := make([]map[string]string, 0)
				for i := 0; i < 7; i++ {
					candle := map[string]string{"candle_n": strconv.Itoa(i), "color": "1"}
					candles = append(candles, candle)
				}
				sendMsg1.Set("candles", candles)
				addr1 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-8"}
				m.srv.sendToOne(sendMsg1, addr1)
				m.endRoom.Table.IsUseful = true
				m.endRoom.InAnimation = false
				log.Println("launch candles")
			}

		}
	}
}

func (m *Match) endingAnimation(s string, dt time.Duration) {
	//sec := dt.Seconds()
	switch s {
	case StageRoom2:
		addrs := []InboxAddress{
			{InboxAddressTypeRoomArduinoDevice, "R-2-7"},
			{InboxAddressTypeRoomArduinoDevice, "R-2-8"},
		}
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("magic_book")
		sendMsg.Set("status", "0")
		m.srv.send(sendMsg, addrs)
		m.library.MagicBooksLightStatus[0] = false
		m.library.MagicBooksLightStatus[1] = false

		m.srv.fakeBooksControl("0", "0", "R-2-9")
		m.srv.fakeBooksControl("0", "0", "R-2-10")
		m.srv.fakeBooksControl("0", "0", "R-2-11")
		m.srv.fakeBooksControl("0", "0", "R-2-12")
		m.srv.fakeBooksControl("0", "0", "R-2-13")
		m.srv.fakeBooksControl("0", "0", "R-2-14")
		m.srv.fakeBooksControl("0", "0", "R-2-15")
		m.srv.fakeBooksControl("0", "0", "R-2-16")
		m.srv.fakeBooksControl("0", "0", "R-2-17")
		m.srv.fakeBooksControl("0", "0", "R-2-18")
		m.srv.fakeBooksControl("0", "0", "R-2-19")
		m.srv.fakeBooksControl("0", "0", "R-2-20")
		m.srv.fakeBooksControl("0", "0", "R-2-21")
		m.srv.fakeBooksControl("0", "0", "R-2-22")
		m.srv.fakeBooksControl("0", "0", "R-2-23")

		//TODO 蜡烛
		candles := make([]map[string]string, 3)
		candles[0] = map[string]string{"candle": "0", "color": "0"}
		candles[1] = map[string]string{"candle": "1", "color": "0"}
		candles[2] = map[string]string{"candle": "2", "color": "0"}
		m.srv.candlesControl(candles, "R-2-2")
		m.srv.candlesControl(candles, "R-2-3")
		m.srv.candlesControl(candles, "R-2-4")
		m.srv.candlesControl(candles, "R-2-5")
		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("door_ctrl")
		sendMsg1.Set("status", "1")
		sendMsg1.Set("time", strconv.FormatFloat(m.opt.Room2OpenDoorDelayTime, 'f', 0, 64))
		addr1 := InboxAddress{InboxAddressTypeDoorArduino, "D-2"}
		m.srv.sendToOne(sendMsg1, addr1)
		log.Println("room2 finish!")
	case StageRoom3:
		m.srv.stairRoomCandlesCtrl("0", "R-3-1")
		m.srv.stairRoomCandlesCtrl("0", "R-3-2")
		m.srv.stairRoomCandlesCtrl("0", "R-3-3")
		m.srv.stairRoomCandlesCtrl("0", "R-3-4")
		m.srv.stairRoomCandlesCtrl("0", "R-3-5")
		m.srv.stairRoomCandlesCtrl("0", "R-3-6")
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("door_ctrl")
		sendMsg.Set("status", "1")
		sendMsg.Set("time", strconv.FormatFloat(m.opt.Room3OpenDoorDelayTime, 'f', 0, 64))
		addr := InboxAddress{InboxAddressTypeDoorArduino, "D-3"}
		m.srv.sendToOne(sendMsg, addr)
		log.Println("room3 finish!")
	case StageRoom4:
		m.magicLab.DeskLight = false
		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("book_desk")
		sendMsg1.Set("status", "0")
		addr1 := InboxAddress{InboxAddressTypeDoorArduino, "R-4-6"}
		m.srv.sendToOne(sendMsg1, addr1)

		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("door_ctrl")
		sendMsg.Set("status", "1")
		sendMsg.Set("time", strconv.FormatFloat(m.opt.Room4OpenDoorDelayTime, 'f', 0, 64))
		addr := InboxAddress{InboxAddressTypeDoorArduino, "D-4"}
		m.srv.sendToOne(sendMsg, addr)
		log.Println("room4 finish!")
	case StageRoom5:
		sendMsg51 := NewInboxMessage()
		sendMsg51.SetCmd("star_led")
		lights51 := make([]map[string]string, 0)
		leds51 := make([]map[string]string, 0)
		leds51 = append(leds51,
			map[string]string{"led_n": "0", "mode": "4"},
			map[string]string{"led_n": "1", "mode": "4"},
			map[string]string{"led_n": "2", "mode": "4"},
			map[string]string{"led_n": "3", "mode": "4"},
			map[string]string{"led_n": "6", "mode": "4"},
			map[string]string{"led_n": "7", "mode": "4"},
			map[string]string{"led_n": "8", "mode": "3"},
			map[string]string{"led_n": "9", "mode": "3"},
		)
		lights51 = append(lights51,
			map[string]string{"light_n": "0", "status": "1"},
			map[string]string{"light_n": "1", "status": "0"},
			map[string]string{"light_n": "2", "status": "0"},
			map[string]string{"light_n": "3", "status": "0"},
			map[string]string{"light_n": "8", "status": "1"},
			map[string]string{"light_n": "9", "status": "0"},
			map[string]string{"light_n": "10", "status": "1"},
			map[string]string{"light_n": "11", "status": "0"},
			map[string]string{"light_n": "19", "status": "1"},
		)
		sendMsg51.Set("light", lights51)
		sendMsg51.Set("led", leds51)
		addr51 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-1"}
		m.srv.sendToOne(sendMsg51, addr51)

		sendMsg52 := NewInboxMessage()
		sendMsg52.SetCmd("star_led")
		lights52 := make([]map[string]string, 0)
		leds52 := make([]map[string]string, 0)
		leds52 = append(leds52,
			map[string]string{"led_n": "4", "mode": "3"},
			map[string]string{"led_n": "5", "mode": "3"},
			map[string]string{"led_n": "26", "mode": "3"},
			map[string]string{"led_n": "27", "mode": "3"},
			map[string]string{"led_n": "28", "mode": "4"},
			map[string]string{"led_n": "29", "mode": "4"},
		)
		lights52 = append(lights52,
			map[string]string{"light_n": "12", "status": "1"},
			map[string]string{"light_n": "13", "status": "1"},
			map[string]string{"light_n": "14", "status": "1"},
			map[string]string{"light_n": "15", "status": "1"},
			map[string]string{"light_n": "16", "status": "1"},
			map[string]string{"light_n": "17", "status": "0"},
			map[string]string{"light_n": "18", "status": "1"},
			map[string]string{"light_n": "35", "status": "0"},
			map[string]string{"light_n": "36", "status": "0"},
		)
		sendMsg52.Set("light", lights52)
		sendMsg52.Set("led", leds52)
		addr52 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-2"}
		m.srv.sendToOne(sendMsg52, addr52)

		sendMsg53 := NewInboxMessage()
		sendMsg53.SetCmd("star_led")
		lights53 := make([]map[string]string, 0)
		leds53 := make([]map[string]string, 0)
		leds53 = append(leds53,
			map[string]string{"led_n": "19", "mode": "3"},
			map[string]string{"led_n": "21", "mode": "3"},
			map[string]string{"led_n": "22", "mode": "3"},
			map[string]string{"led_n": "23", "mode": "3"},
			map[string]string{"led_n": "24", "mode": "3"},
			map[string]string{"led_n": "25", "mode": "3"},
			map[string]string{"led_n": "30", "mode": "4"},
			map[string]string{"led_n": "31", "mode": "4"},
			map[string]string{"led_n": "32", "mode": "4"},
		)
		lights53 = append(lights53,
			map[string]string{"light_n": "29", "status": "1"},
			map[string]string{"light_n": "30", "status": "1"},
			map[string]string{"light_n": "32", "status": "1"},
			map[string]string{"light_n": "33", "status": "1"},
			map[string]string{"light_n": "34", "status": "1"},
		)
		sendMsg53.Set("light", lights53)
		sendMsg53.Set("led", leds53)
		addr53 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-3"}
		m.srv.sendToOne(sendMsg53, addr53)

		sendMsg54 := NewInboxMessage()
		sendMsg54.SetCmd("star_led")
		lights54 := make([]map[string]string, 0)
		leds54 := make([]map[string]string, 0)
		leds54 = append(leds54,
			map[string]string{"led_n": "15", "mode": "4"},
			map[string]string{"led_n": "16", "mode": "4"},
			map[string]string{"led_n": "17", "mode": "4"},
			map[string]string{"led_n": "18", "mode": "3"},
			map[string]string{"led_n": "20", "mode": "3"},
		)
		lights54 = append(lights54,
			map[string]string{"light_n": "22", "status": "0"},
			map[string]string{"light_n": "25", "status": "0"},
			map[string]string{"light_n": "26", "status": "0"},
			map[string]string{"light_n": "27", "status": "1"},
			map[string]string{"light_n": "28", "status": "0"},
			map[string]string{"light_n": "31", "status": "1"},
		)
		sendMsg54.Set("light", lights54)
		sendMsg54.Set("led", leds54)
		addr54 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-4"}
		m.srv.sendToOne(sendMsg54, addr54)

		sendMsg55 := NewInboxMessage()
		sendMsg55.SetCmd("star_led")
		lights55 := make([]map[string]string, 0)
		leds55 := make([]map[string]string, 0)
		leds55 = append(leds55,
			map[string]string{"led_n": "10", "mode": "4"},
			map[string]string{"led_n": "11", "mode": "3"},
			map[string]string{"led_n": "12", "mode": "3"},
			map[string]string{"led_n": "13", "mode": "3"},
			map[string]string{"led_n": "14", "mode": "3"},
		)
		lights55 = append(lights55,
			map[string]string{"light_n": "4", "status": "1"},
			map[string]string{"light_n": "5", "status": "0"},
			map[string]string{"light_n": "6", "status": "0"},
			map[string]string{"light_n": "7", "status": "1"},
			map[string]string{"light_n": "20", "status": "1"},
			map[string]string{"light_n": "21", "status": "1"},
			map[string]string{"light_n": "23", "status": "1"},
			map[string]string{"light_n": "24", "status": "1"},
		)
		sendMsg55.Set("light", lights55)
		sendMsg55.Set("led", leds55)
		addr55 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-5"}
		m.srv.sendToOne(sendMsg55, addr55)

		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("magic_rob")
		sendMsg1.Set("status", "1")
		addr1 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-7"}
		m.srv.sendToOne(sendMsg1, addr1)

		sendMsg2 := NewInboxMessage()
		sendMsg2.SetCmd("door_ctrl")
		sendMsg2.Set("status", "1")
		sendMsg2.Set("time", strconv.FormatFloat(m.opt.Room5OpenDoorDelayTime, 'f', 0, 64))
		addr2 := InboxAddress{InboxAddressTypeDoorArduino, "D-5"}
		m.srv.sendToOne(sendMsg2, addr2)
		log.Println("room5 finish!")
	case StageRoom6:
		if m.endRoom.Ending == 1 { //good
			m.endRoom.CandleMode = 0
			sendMsg1 := NewInboxMessage()
			sendMsg1.SetCmd("led_candle")
			sendMsg1.Set("mode", "0")
			addr1 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-8"}
			m.srv.sendToOne(sendMsg1, addr1)
		}
		m.endRoom.Table.IsDestroyed = true
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("magic_table")
		sendMsg.Set("destroyed", "1")
		addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
		m.srv.sendToOne(sendMsg, addr)

	}
}

func (m *Match) dealMagicWords(room interface{}, magicWords int) {
	if magicWords == 0 {
		return
	}
	sendMsg := NewInboxMessage()
	switch room.(type) {
	case *Room2:
		if magicWords == 1 {
			if !m.library.MagicBooksLightStatus[0] {
				sendMsg.SetCmd("magic_book")
				sendMsg.Set("status", "1")
				addrs := []InboxAddress{{InboxAddressTypeRoomArduinoDevice, "R-2-7"}, {InboxAddressTypeRoomArduinoDevice, "R-2-8"}}
				m.srv.send(sendMsg, addrs)
				m.library.MagicBooksLightStatus[0] = true
				m.library.MagicBooksLightStatus[1] = true
			}
		} else if magicWords == 2 {
			if m.library.MagicBooksLightStatus[0] {
				sendMsg.SetCmd("magic_book")
				sendMsg.Set("status", "0")
				addrs := []InboxAddress{{InboxAddressTypeRoomArduinoDevice, "R-2-7"}, {InboxAddressTypeRoomArduinoDevice, "R-2-8"}}
				m.srv.send(sendMsg, addrs)
				m.library.MagicBooksLightStatus[0] = false
				m.library.MagicBooksLightStatus[1] = false
			}
		} else if !m.library.Table.IsDestroyed {
			switch magicWords {
			case 3:
				if m.library.Table.IsFinish {
					m.library.Table.IsDestroyed = true
				}
			case 4:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-9")
					m.library.FakeBooks[1] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-9")
					m.library.FakeBooks[1] = false
					m.library.CurrentFakeBookLight--
				}
			case 5:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-10")
					m.library.FakeBooks[2] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-10")
					m.library.FakeBooks[2] = false
					m.library.CurrentFakeBookLight--
				}
			case 6:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-11")
					m.library.FakeBooks[3] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-11")
					m.library.FakeBooks[3] = false
					m.library.CurrentFakeBookLight--
				}
			case 7:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-12")
					m.library.FakeBooks[4] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-12")
					m.library.FakeBooks[4] = false
					m.library.CurrentFakeBookLight--
				}
			case 8:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-13")
					m.library.FakeBooks[5] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-13")
					m.library.FakeBooks[5] = false
					m.library.CurrentFakeBookLight--
				}
			case 9:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-14")
					m.library.FakeBooks[6] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-14")
					m.library.FakeBooks[6] = false
					m.library.CurrentFakeBookLight--
				}
			case 10:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-15")
					m.library.FakeBooks[7] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-15")
					m.library.FakeBooks[7] = false
					m.library.CurrentFakeBookLight--
				}
			case 11:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-16")
					m.library.FakeBooks[8] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-16")
					m.library.FakeBooks[8] = false
					m.library.CurrentFakeBookLight--
				}
			case 12:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-17")
					m.library.FakeBooks[9] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-17")
					m.library.FakeBooks[9] = false
					m.library.CurrentFakeBookLight--
				}
			case 13:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-18")
					m.library.FakeBooks[10] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-18")
					m.library.FakeBooks[10] = false
					m.library.CurrentFakeBookLight--
				}
			case 14:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-19")
					m.library.FakeBooks[11] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-19")
					m.library.FakeBooks[11] = false
					m.library.CurrentFakeBookLight--
				}
			case 15:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-20")
					m.library.FakeBooks[12] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-20")
					m.library.FakeBooks[12] = false
					m.library.CurrentFakeBookLight--
				}
			case 16:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-21")
					m.library.FakeBooks[13] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-21")
					m.library.FakeBooks[13] = false
					m.library.CurrentFakeBookLight--
				}
			case 17:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-22")
					m.library.FakeBooks[14] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-22")
					m.library.FakeBooks[14] = false
					m.library.CurrentFakeBookLight--
				}
			case 18:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-23")
					m.library.FakeBooks[15] = true
					m.library.CurrentFakeBookLight++
				} else if !m.library.Table.IsUseful && m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("1", "2", "R-2-23")
					m.library.FakeBooks[15] = false
					m.library.CurrentFakeBookLight--
				}
			}
		}
		m.library.MagicWords = 0
	case *Room3:
		if m.stairRoom.Table.IsUseful && !m.stairRoom.Table.IsDestroyed {
			switch magicWords {
			case 3:
				if m.stairRoom.Table.IsFinish {
					m.stairRoom.Table.IsDestroyed = true
				} else {
					m.stairRoom.Step = 2
					m.destoryFailed(m.stairRoom)
					for k,_ := range m.stairRoom.Candles {
						m.stairRoom.Candles[k] = 1
					}
					log.Print("destory failed!turn to step 2")
				}
			}
		}
		m.stairRoom.MagicWords = 0
	case *Room4:
		if magicWords == 1 {
			m.magicLab.DeskLight = true
			sendMsg.SetCmd("book_desk")
			sendMsg.Set("status", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-6"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 2 {
			m.magicLab.DeskLight = false
			sendMsg.SetCmd("book_desk")
			sendMsg.Set("status", "0")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-6"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 4 && !m.magicLab.Table.IsUseful {
			if m.magicLab.Stands[0].IsPowerOn {
				return
			}
			m.magicLab.Stands[0].IsPowerOn = true
			sendMsg.SetCmd("magic_desk")
			sendMsg.Set("useful", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-1"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 5 && !m.magicLab.Table.IsUseful {
			if m.magicLab.Stands[1].IsPowerOn {
				return
			}
			m.magicLab.Stands[1].IsPowerOn = true
			sendMsg.SetCmd("magic_desk")
			sendMsg.Set("useful", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-2"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 6 && !m.magicLab.Table.IsUseful {
			if m.magicLab.Stands[2].IsPowerOn {
				return
			}
			m.magicLab.Stands[2].IsPowerOn = true
			sendMsg.SetCmd("magic_desk")
			sendMsg.Set("useful", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-3"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 7 && !m.magicLab.Table.IsUseful {
			if m.magicLab.Stands[3].IsPowerOn {
				return
			}
			m.magicLab.Stands[3].IsPowerOn = true
			sendMsg.SetCmd("magic_desk")
			sendMsg.Set("useful", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-4"}
			m.srv.sendToOne(sendMsg, addr)
		} else if m.magicLab.Table.IsUseful {
			switch magicWords {
			case 3:
				if m.magicLab.Table.IsFinish {
					m.magicLab.Table.IsDestroyed = true
				}
			}
		}
		m.magicLab.MagicWords = 0
	case *Room5:
		if magicWords == 1 {
			m.starTower.LightWall = true
			sendMsg.SetCmd("light_ctrl")
			sendMsg.Set("status", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-8"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 2 {
			m.starTower.LightWall = false
			sendMsg.SetCmd("light_ctrl")
			sendMsg.Set("status", "0")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-8"}
			m.srv.sendToOne(sendMsg, addr)
		}
		if m.starTower.Table.IsUseful && !m.starTower.Table.IsDestroyed {
			switch magicWords {
			case 3:
				if m.starTower.Table.IsFinish {
					m.starTower.Table.IsDestroyed = true
				}
			case 21:
				log.Println("reset star!")
				for k, _ := range m.starTower.ConstellationSymbol {
					m.starTower.ConstellationSymbol[k] = false
				}
				sendMsg51 := NewInboxMessage()
				sendMsg51.SetCmd("star_led")
				lights51 := make([]map[string]string, 0)
				leds51 := make([]map[string]string, 0)
				leds51 = append(leds51,
					map[string]string{"led_n": "0", "mode": "0"},
					map[string]string{"led_n": "1", "mode": "0"},
					map[string]string{"led_n": "2", "mode": "0"},
					map[string]string{"led_n": "3", "mode": "0"},
					map[string]string{"led_n": "6", "mode": "0"},
					map[string]string{"led_n": "7", "mode": "0"},
					map[string]string{"led_n": "8", "mode": "0"},
					map[string]string{"led_n": "9", "mode": "0"},
				)
				lights51 = append(lights51,
					map[string]string{"light_n": "0", "status": "0"},
					map[string]string{"light_n": "1", "status": "0"},
					map[string]string{"light_n": "2", "status": "0"},
					map[string]string{"light_n": "3", "status": "0"},
					map[string]string{"light_n": "8", "status": "0"},
					map[string]string{"light_n": "9", "status": "0"},
					map[string]string{"light_n": "10", "status": "0"},
					map[string]string{"light_n": "11", "status": "0"},
					map[string]string{"light_n": "19", "status": "0"},
				)
				sendMsg51.Set("light", lights51)
				sendMsg51.Set("led", leds51)
				addr51 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-1"}
				m.srv.sendToOne(sendMsg51, addr51)

				sendMsg52 := NewInboxMessage()
				sendMsg52.SetCmd("star_led")
				lights52 := make([]map[string]string, 0)
				leds52 := make([]map[string]string, 0)
				leds52 = append(leds52,
					map[string]string{"led_n": "4", "mode": "0"},
					map[string]string{"led_n": "5", "mode": "0"},
					map[string]string{"led_n": "26", "mode": "0"},
					map[string]string{"led_n": "27", "mode": "0"},
					map[string]string{"led_n": "28", "mode": "0"},
					map[string]string{"led_n": "29", "mode": "0"},
				)
				lights52 = append(lights52,
					map[string]string{"light_n": "12", "status": "0"},
					map[string]string{"light_n": "13", "status": "0"},
					map[string]string{"light_n": "14", "status": "0"},
					map[string]string{"light_n": "15", "status": "0"},
					map[string]string{"light_n": "16", "status": "0"},
					map[string]string{"light_n": "17", "status": "0"},
					map[string]string{"light_n": "18", "status": "0"},
					map[string]string{"light_n": "35", "status": "0"},
					map[string]string{"light_n": "36", "status": "0"},
				)
				sendMsg52.Set("light", lights52)
				sendMsg52.Set("led", leds52)
				addr52 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-2"}
				m.srv.sendToOne(sendMsg52, addr52)

				sendMsg53 := NewInboxMessage()
				sendMsg53.SetCmd("star_led")
				lights53 := make([]map[string]string, 0)
				leds53 := make([]map[string]string, 0)
				leds53 = append(leds53,
					map[string]string{"led_n": "19", "mode": "0"},
					map[string]string{"led_n": "21", "mode": "0"},
					map[string]string{"led_n": "22", "mode": "0"},
					map[string]string{"led_n": "23", "mode": "0"},
					map[string]string{"led_n": "24", "mode": "0"},
					map[string]string{"led_n": "25", "mode": "0"},
					map[string]string{"led_n": "30", "mode": "0"},
					map[string]string{"led_n": "31", "mode": "0"},
					map[string]string{"led_n": "32", "mode": "0"},
				)
				lights53 = append(lights53,
					map[string]string{"light_n": "29", "status": "0"},
					map[string]string{"light_n": "30", "status": "0"},
					map[string]string{"light_n": "32", "status": "0"},
					map[string]string{"light_n": "33", "status": "0"},
					map[string]string{"light_n": "34", "status": "0"},
				)
				sendMsg53.Set("light", lights53)
				sendMsg53.Set("led", leds53)
				addr53 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-3"}
				m.srv.sendToOne(sendMsg53, addr53)

				sendMsg54 := NewInboxMessage()
				sendMsg54.SetCmd("star_led")
				lights54 := make([]map[string]string, 0)
				leds54 := make([]map[string]string, 0)
				leds54 = append(leds54,
					map[string]string{"led_n": "15", "mode": "0"},
					map[string]string{"led_n": "16", "mode": "0"},
					map[string]string{"led_n": "17", "mode": "0"},
					map[string]string{"led_n": "18", "mode": "0"},
					map[string]string{"led_n": "20", "mode": "0"},
				)
				lights54 = append(lights54,
					map[string]string{"light_n": "22", "status": "0"},
					map[string]string{"light_n": "25", "status": "0"},
					map[string]string{"light_n": "26", "status": "0"},
					map[string]string{"light_n": "27", "status": "0"},
					map[string]string{"light_n": "28", "status": "0"},
					map[string]string{"light_n": "31", "status": "0"},
				)
				sendMsg54.Set("light", lights54)
				sendMsg54.Set("led", leds54)
				addr54 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-4"}
				m.srv.sendToOne(sendMsg54, addr54)

				sendMsg55 := NewInboxMessage()
				sendMsg55.SetCmd("star_led")
				lights55 := make([]map[string]string, 0)
				leds55 := make([]map[string]string, 0)
				leds55 = append(leds55,
					map[string]string{"led_n": "10", "mode": "0"},
					map[string]string{"led_n": "11", "mode": "0"},
					map[string]string{"led_n": "12", "mode": "0"},
					map[string]string{"led_n": "13", "mode": "0"},
					map[string]string{"led_n": "14", "mode": "0"},
				)
				lights55 = append(lights55,
					map[string]string{"light_n": "4", "status": "0"},
					map[string]string{"light_n": "5", "status": "0"},
					map[string]string{"light_n": "6", "status": "0"},
					map[string]string{"light_n": "7", "status": "0"},
					map[string]string{"light_n": "20", "status": "0"},
					map[string]string{"light_n": "21", "status": "0"},
					map[string]string{"light_n": "23", "status": "0"},
					map[string]string{"light_n": "24", "status": "0"},
				)
				sendMsg55.Set("light", lights55)
			}
		}
		m.starTower.MagicWords = 0
	case *Room6:
		//if m.endRoom.Table.IsUseful && !m.endRoom.Table.IsDestroyed {
		//switch magicWords {
		//case 3:
		//case 20:
		//}
		//}
	}

}

func att(id string) InboxAddressType {
	if id == "" {
		return InboxAddressTypeUnknown
	} else if strings.HasPrefix(id, "R") {
		return InboxAddressTypeRoomArduinoDevice
	} else if strings.HasPrefix(id, "L") {
		return InboxAddressTypeLightArduinoDevice
	} else if strings.HasPrefix(id, "B") {
		return InboxAddressTypeMusicArduino
	} else if strings.HasPrefix(id, "D") {
		return InboxAddressTypeDoorArduino
	}
	return InboxAddressTypeUnknown
}
