package core

import (
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
	StageRoom1 = "going-r1"
	StageRoom2 = "going-r2"
	StageRoom3 = "going-r3"
	StageRoom4 = "going-r4"
	StageRoom5 = "going-r5"
	StageRoom6 = "going-r6"
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
		m.TotalTime += dt.Seconds()
		<-tickChan
		m.handleInputs()
		if m.Stage == StageEnd {
			break
		}
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
		case "R-1-2":
			if msg.GetStr("ST") == "1" {
				m.livingRoom.CandleStatus = 1
			} else {
				m.livingRoom.CandleStatus = 0
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
		case "R-2-1":
			// mode := msg.GetStr("MD")
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("fake_book")
			//sendMsg.Set("time", strconv.FormatFloat(m.opt.FakeAnimationTime, 'f', 0, 64))
			//books := make([]map[string]string, 0)
			//if mode == "0" {
			//sendMsg.Set("mode", "0")
			//c := []rune(msg.GetStr("BK"))
			//var open bool
			//for k, v := range c {
			//if v == '1' {
			//open = true
			//} else {
			//open = false
			//}
			//if open != m.library.FakeBooks[k] {
			//if m.library.FakeBooks[k] {
			//books = append(books, map[string]string{
			//"book_n": strconv.Itoa(k),
			//"book_m": "1",
			//})
			//} else {
			//books = append(books, map[string]string{
			//"book_n": strconv.Itoa(k),
			//"book_m": "0",
			//})
			//}
			//}
			//}
			//if len(books) > 0 {
			//sendMsg.Set("book", books)
			//m.srv.sendToOne(sendMsg, addr)
			//}
			//}
			//if mode == "1" {
			//m.library.InAnimation = true
			//} else {
			//m.library.InAnimation = false
			//}
		case "R-2-2":
			//c := []rune(msg.GetStr("C"))
			//candles := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("led_candle")
			//for k, v := range c {
			//if m.library.Candles[k] != int(v-'0') {
			//candles = append(candles, map[string]string{
			//"candle": strconv.Itoa(k),
			//"color":  strconv.Itoa(m.library.Candles[k]),
			//})
			//}
			//}
			//if len(candles) > 0 {
			//sendMsg.Set("candles", candles)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-2-3":
			//c := []rune(msg.GetStr("C"))
			//candles := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("led_candle")
			//for k, v := range c {
			//if m.library.Candles[k] != int(v-'0') {
			//candles = append(candles, map[string]string{
			//"candle": strconv.Itoa(k + 3),
			//"color":  strconv.Itoa(m.library.Candles[k+3]),
			//})
			//}
			//}
			//if len(candles) > 0 {
			//sendMsg.Set("candles", candles)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-2-4":
			//c := []rune(msg.GetStr("C"))
			//candles := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("led_candle")
			//for k, v := range c {
			//if m.library.Candles[k] != int(v-'0') {
			//candles = append(candles, map[string]string{
			//"candle": strconv.Itoa(k + 6),
			//"color":  strconv.Itoa(m.library.Candles[k+6]),
			//})
			//}
			//}
			//if len(candles) > 0 {
			//sendMsg.Set("candles", candles)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-2-5":
			// c := []rune(msg.GetStr("C"))
			//candles := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("led_candle")
			//for k, v := range c {
			//if m.library.Candles[k] != int(v-'0') {
			//candles = append(candles, map[string]string{
			//"candle": strconv.Itoa(k + 9),
			//"color":  strconv.Itoa(m.library.Candles[k+9]),
			//})
			//}
			//}
			//if len(candles) > 0 {
			//sendMsg.Set("candles", candles)
			//m.srv.sendToOne(sendMsg, addr)
			//}
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
				m.stairRoom.Candles[0], _ = strconv.Atoi(color)
			}
		case "R-3-2":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[1] = 0
			} else {
				color := msg.GetStr("C")
				m.stairRoom.Candles[1], _ = strconv.Atoi(color)
			}
		case "R-3-3":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[2] = 0
			} else {
				color := msg.GetStr("C")
				m.stairRoom.Candles[2], _ = strconv.Atoi(color)
			}
		case "R-3-4":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[3] = 0
			} else {
				color := msg.GetStr("C")
				m.stairRoom.Candles[3], _ = strconv.Atoi(color)
			}
		case "R-3-5":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[4] = 0
			} else {
				color := msg.GetStr("C")
				m.stairRoom.Candles[4], _ = strconv.Atoi(color)
			}
		case "R-3-6":
			st := msg.GetStr("ST")
			if st == "0" {
				m.stairRoom.Candles[5] = 0
			} else {
				color := msg.GetStr("C")
				m.stairRoom.Candles[5], _ = strconv.Atoi(color)
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
			m.library.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.stairRoom, m.stairRoom.MagicWords)
		case "D-3":
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
		case "R-4-1":
			status := msg.GetStr("P")
			power := make([]map[string]string, 1)
			power = append(power, map[string]string{"power_type": "1", "status": status})
			//m.srv.powerStatus(power)
			if status == "2" {
				m.magicLab.Stands[0].IsPowerful = true
			}
		case "R-4-2":
			status := msg.GetStr("P")
			power := make([]map[string]string, 1)
			power = append(power, map[string]string{"power_type": "2", "status": status})
			//m.srv.powerStatus(power)
			if status == "2" {
				m.magicLab.Stands[1].IsPowerful = true
			}
		case "R-4-3":
			status := msg.GetStr("P")
			power := make([]map[string]string, 1)
			power = append(power, map[string]string{"power_type": "3", "status": status})
			//m.srv.powerStatus(power)
			if status == "2" {
				m.magicLab.Stands[2].IsPowerful = true
			}
		case "R-4-4":
			status := msg.GetStr("P")
			power := make([]map[string]string, 1)
			power = append(power, map[string]string{"power_type": "4", "status": status})
			//m.srv.powerStatus(power)
			if status == "2" {
				m.magicLab.Stands[3].IsPowerful = true
			}
		case "R-4-5":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_table")
			if msg.GetStr("U") == "1" {
				if m.magicLab.Table.IsUseful != true {
					sendMsg.Set("useful", "0")
				}
			} else {
				if m.magicLab.Table.IsUseful != false {
					sendMsg.Set("useful", "1")
				}
			}
			if msg.GetStr("F") == "1" {
				if m.magicLab.Table.IsFinish != true {
					sendMsg.Set("finish", "0")
				}
			} else {
				if m.magicLab.Table.IsFinish != false {
					sendMsg.Set("finish", "1")
				}
			}
			if len(sendMsg.Data) > 2 {
				m.srv.sendToOne(sendMsg, addr)
			}
			m.magicLab.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.magicLab, m.magicLab.MagicWords)

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
		case "R-5-1":
			//c := []rune(msg.GetStr("L"))
			//lights := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("star_led")
			//for k, v := range c {
			//if v == '1' {
			//if !m.starTower.ConstellationLight[k] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k),
			//"status":  "0",
			//})
			//}
			//} else {
			//if m.starTower.ConstellationLight[k] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k),
			//"status":  "1",
			//})

			//}
			//}
			//}
			//if len(lights) > 0 {
			//sendMsg.Set("light", lights)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-5-2":
			//c := []rune(msg.GetStr("L"))
			//lights := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("star_led")
			//for k, v := range c {
			//if v == '1' {
			//if !m.starTower.ConstellationLight[k+9] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k + 9),
			//"status":  "0",
			//})
			//}
			//} else {
			//if m.starTower.ConstellationLight[k+9] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k + 9),
			//"status":  "1",
			//})
			//}
			//}
			//}
			//if len(lights) > 0 {
			//sendMsg.Set("light", lights)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-5-3":
			//c := []rune(msg.GetStr("L"))
			//lights := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("star_led")
			//for k, v := range c {
			//if v == '1' {
			//if !m.starTower.ConstellationLight[k+18] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k + 18),
			//"status":  "0",
			//})
			//}
			//} else {
			//if m.starTower.ConstellationLight[k+18] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k + 18),
			//"status":  "1",
			//})
			//}
			//}
			//}
			//if len(lights) > 0 {
			//sendMsg.Set("light", lights)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-5-4":
			//c := []rune(msg.GetStr("L"))
			//lights := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("star_led")
			//for k, v := range c {
			//if v == '1' {
			//if !m.starTower.ConstellationLight[k+23] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k + 23),
			//"status":  "0",
			//})
			//}
			//} else {
			//if m.starTower.ConstellationLight[k+23] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k + 23),
			//"status":  "1",
			//})
			//}
			//}
			//}
			//if len(lights) > 0 {
			//sendMsg.Set("light", lights)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-5-5":
			//c := []rune(msg.GetStr("L"))
			//lights := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("star_led")
			//for k, v := range c {
			//if v == '1' {
			//if !m.starTower.ConstellationLight[k+29] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k + 29),
			//"status":  "0",
			//})
			//}
			//} else {
			//if m.starTower.ConstellationLight[k+29] {
			//lights = append(lights, map[string]string{
			//"light_n": strconv.Itoa(k + 29),
			//"status":  "1",
			//})
			//}
			//}
			//}
			//if len(lights) > 0 {
			//sendMsg.Set("light", lights)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-5-6":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_table")
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
			if len(sendMsg.Data) > 2 {
				m.srv.sendToOne(sendMsg, addr)
			}
			if !m.starTower.InAnimation {
				m.starTower.CurrentConstellationLight, _ = strconv.Atoi(msg.GetStr("S"))
			}
			m.dealStar(m.starTower.CurrentConstellationLight)
			m.magicLab.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.magicLab, m.magicLab.MagicWords)
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
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("power_point")
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				sendMsg.Set("type_power", strconv.Itoa(m.endRoom.CurrentSymbol))
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[0] {
				sendMsg.Set("useful", m.endRoom.PowerPointUseful[0])
			}
			power, _ := strconv.Atoi(msg.GetStr("F"))
			if power == m.endRoom.CurrentSymbol && power != 0 {
				m.endRoom.PowerPoint[0] = m.endRoom.CurrentSymbol
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("magic_table")
				sendMsg1.Set("power_done", "1")
				addr1 := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg1, addr1)
				m.endRoom.CurrentSymbol = 0
				m.broadSymbolToArduino(0)
			}
			if len(sendMsg.Data) > 1 {
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-6-2":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("power_point")
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				sendMsg.Set("type_power", strconv.Itoa(m.endRoom.CurrentSymbol))
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[1] {
				sendMsg.Set("useful", m.endRoom.PowerPointUseful[1])
			}
			power, _ := strconv.Atoi(msg.GetStr("F"))
			if power == m.endRoom.CurrentSymbol && power != 0 {
				m.endRoom.PowerPoint[1] = m.endRoom.CurrentSymbol
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("magic_table")
				sendMsg1.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg1, addr)
				m.endRoom.CurrentSymbol = 0
				m.broadSymbolToArduino(0)
			}
			if len(sendMsg.Data) > 1 {
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-6-3":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("power_point")
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				sendMsg.Set("type_power", strconv.Itoa(m.endRoom.CurrentSymbol))
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[2] {
				sendMsg.Set("useful", m.endRoom.PowerPointUseful[2])
			}
			power, _ := strconv.Atoi(msg.GetStr("F"))
			if power == m.endRoom.CurrentSymbol && power != 0 {
				m.endRoom.PowerPoint[2] = m.endRoom.CurrentSymbol
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("magic_table")
				sendMsg1.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg1, addr)
				m.endRoom.CurrentSymbol = 0
				m.broadSymbolToArduino(0)
			}
			if len(sendMsg.Data) > 1 {
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-6-4":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("power_point")
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				sendMsg.Set("type_power", strconv.Itoa(m.endRoom.CurrentSymbol))
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[3] {
				sendMsg.Set("useful", m.endRoom.PowerPointUseful[3])
			}
			power, _ := strconv.Atoi(msg.GetStr("F"))
			if power == m.endRoom.CurrentSymbol && power != 0 {
				m.endRoom.PowerPoint[4] = m.endRoom.CurrentSymbol
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("magic_table")
				sendMsg1.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg1, addr)
				m.endRoom.CurrentSymbol = 0
				m.broadSymbolToArduino(0)
			}
			if len(sendMsg.Data) > 1 {
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-6-5":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("power_point")
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				sendMsg.Set("type_power", strconv.Itoa(m.endRoom.CurrentSymbol))
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[4] {
				sendMsg.Set("useful", m.endRoom.PowerPointUseful[4])
			}
			power, _ := strconv.Atoi(msg.GetStr("F"))
			if power == m.endRoom.CurrentSymbol && power != 0 {
				m.endRoom.PowerPoint[5] = m.endRoom.CurrentSymbol
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("magic_table")
				sendMsg1.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg1, addr)
				m.endRoom.CurrentSymbol = 0
				m.broadSymbolToArduino(0)
			}
			if len(sendMsg.Data) > 1 {
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-6-6":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("power_point")
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				sendMsg.Set("type_power", strconv.Itoa(m.endRoom.CurrentSymbol))
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[5] {
				sendMsg.Set("useful", m.endRoom.PowerPointUseful[5])
			}
			power, _ := strconv.Atoi(msg.GetStr("F"))
			if power == m.endRoom.CurrentSymbol && power != 0 {
				m.endRoom.PowerPoint[6] = m.endRoom.CurrentSymbol
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("magic_table")
				sendMsg1.Set("power_done", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-6-7"}
				m.srv.sendToOne(sendMsg1, addr)
				m.endRoom.CurrentSymbol = 0
				m.broadSymbolToArduino(0)
			}
			if len(sendMsg.Data) > 1 {
				m.srv.sendToOne(sendMsg, addr)
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
			m.endRoom.CurrentSymbol, _ = strconv.Atoi(msg.GetStr("TY"))
			m.broadSymbolToArduino(m.endRoom.CurrentSymbol)
			m.endRoom.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.endRoom, m.endRoom.MagicWords)
		case "R-6-8":
			// candleMode := msg.GetStr("mode")
			//if candleMode != "1" {
			//return
			//}
			//c := []rune(msg.GetStr("C"))
			//candles := make([]map[string]string, 0)
			//sendMsg := NewInboxMessage()
			//sendMsg.SetCmd("led_candle")
			//for k, v := range c {
			//if int(v-'0') != m.endRoom.Candles[k] {
			//candles = append(candles, map[string]string{
			//"candle": strconv.Itoa(k),
			//"color":  strconv.Itoa(m.endRoom.Candles[k]),
			//})
			//}
			//}
			//if len(candles) > 0 {
			//sendMsg.Set("candle", candles)
			//m.srv.sendToOne(sendMsg, addr)
			//}
		case "R-6-9":
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
			if m.library.Step < 4 {
				m.library.Step++
			} else {
				m.setStage(StageRoom3)
			}
		case StageRoom3:
			if m.stairRoom.Step < 4 {
				m.library.Step++
			} else {
				m.setStage(StageRoom4)
			}
		case StageRoom4:
			if m.magicLab.Step < 4 {
				m.library.Step++
			} else {
				m.setStage(StageRoom5)
			}
		case StageRoom5:
			if m.starTower.Step < 4 {
				m.library.Step++
			} else {
				m.setStage(StageRoom6)
			}
		case StageRoom6:
			if m.endRoom.Step < 5 {
				m.library.Step++
			} else {
				m.setStage(StageEnd)
			}
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
	}
	log.Printf("game stage:%v\n", s)
	m.Stage = s
}

func (m *Match) gameStage(dt time.Duration) {
	if m.Stage == "" {
		log.Println("game stage error!")
		return
	}
	switch m.Stage {
	case READY:
		if m.CurrentBgm != 1 {
			m.CurrentBgm = 1
			m.bgmPlay(m.CurrentBgm)
		}
	case StageRoom1:
		if m.livingRoom.DoorMirror == DoorOpen {
			m.room1Animation()
			log.Println("room 1 finish!")
		}
	case StageRoom2:
		if m.library.Step == 1 {
			m.fakeBooksAnimation(dt)
			if m.fakeActNum() == 5 {
				if m.ensureFakeBooks() {
					m.fakeBooksAnimation(dt)
				} else {
					m.fakeBooksErrorAnimation(dt)
				}
			}
		} else if m.library.Step == 2 {
			if m.library.Table.IsFinish {
				m.library.Step = 3
				log.Println("room2 step 2 finish!")
			}
		} else if m.library.Step == 3 {
			if m.library.Table.IsDestroyed {
				m.endingAnimation(StageRoom2, dt)
			}
		}
	case StageRoom3:
		if m.stairRoom.Step == 1 {
			if m.ensureCandlesPoweroff() {
				m.magicTableAnimation(StageRoom3)
				m.stairRoom.Step = 2
				log.Println("room3 step 1 finish!")
			}
		} else if m.stairRoom.Step == 2 {
			if m.ensureCandlesColor() {
				m.stairRoom.Table.IsFinish = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("useful", "1")
				sendMsg.Set("finish", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-3-7"}
				m.srv.sendToOne(sendMsg, addr)
				m.stairRoom.Step = 3
				log.Println("room3 step 2 finish!")
			}
		} else if m.stairRoom.Step == 3 {
			if m.stairRoom.Table.IsDestroyed {
				m.endingAnimation(StageRoom3, dt)
				m.stairRoom.DoorExit = DoorOpen
				log.Println("room3 step 3 finish!")
			}
		}
	case StageRoom4:
		if m.magicLab.Step == 1 {
			if m.ensureMagicStandsPowerOn() {
				m.magicTableAnimation(StageRoom4)
				m.magicLab.Step = 2
				log.Println("room4 step 1 finish!")
			}
		} else if m.magicLab.Step == 2 {
			if m.ensureMagicStandsPowerFul() {
				m.magicLab.Table.IsFinish = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("useful", "1")
				sendMsg.Set("finish", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-5"}
				m.srv.sendToOne(sendMsg, addr)
				m.magicLab.Step = 3
				log.Println("room4 step 2 finish!")

			}
		} else if m.magicLab.Step == 3 {
			if m.magicLab.Table.IsDestroyed {
				m.endingAnimation(StageRoom4, dt)
				m.magicLab.DoorExit = DoorOpen
				log.Println("room4 step 3 finish!")
			}
		}
	case StageRoom5:
		if m.starTower.Step == 1 {
			if m.starTower.Table.IsUseful {
				m.starTower.Step = 2
				log.Println("room 5 step 1 finish!")
			}
		} else if m.starTower.Step == 2 {
			if m.ensureConstellationSymbol() || m.starTower.Table.IsFinish {
				m.starTower.Table.IsFinish = true
				m.starTower.Step = 3
				log.Println("room 5 step 2 finish!")
			}
		} else if m.starTower.Step == 3 {
			if m.starTower.Table.IsDestroyed {
				m.endingAnimation(StageRoom5, dt)
				m.starTower.DoorExit = DoorOpen
				m.starTower.DoorMagicRod = DoorOpen
				log.Println("room 5 step 3 finish!")
			} else if !m.starTower.Table.IsFinish {
				m.starTower.Step = 2
			}

		}
	case StageRoom6:
		if m.endRoom.Step == 1 {
			//if m.endRoom.NextStep == 2 {
			//m.amMagicAnimation()
			//m.endRoom.Step = 2
			//log.Println("room 6 step 1 finish!")
			//}
		} else if m.endRoom.Step == 2 {
			//if m.exitRoom.ButtonNextStage { //endroom 数据维护需要锁
			m.bgmPlay(8) //bgm
			m.endRoom.Table.IsUseful = true
			m.magicTableAnimation(StageRoom6)
			//m.endRoom.Step = 3
			//m.endRoom.NextStep = 3
			log.Println("room 6 step 2 finish!")
			//}
		} else if m.endRoom.Step == 3 {
			//if m.ensureElementSymbol() {
			//m.endRoom.Table.IsFinish = true
			//} else {
			//m.endRoom.Table.IsFinish = false
			//}
			//if m.endRoom.Table.IsFinish && m.endRoom.NextStep == 4 {
			m.endingAnimation(StageRoom6, dt)
			//m.endRoom.Table.IsDestroyed = true
			log.Println("room 6 step 3 finish!")
			//}
		} else if m.endRoom.Step == 4 {
			m.endRoom.DoorExit = DoorOpen
		}
	case StageEnd:
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("door_ctrl")
		sendMsg.Set("status", "1")
		sendMsg.Set("time", m.opt.Room2OpenDoorDelayTime)
		addr := InboxAddress{InboxAddressTypeDoorArduino, "D-6"}
		m.srv.sendToOne(sendMsg, addr)

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
	}
}

func (m *Match) reset() {
	m.initHardwareData()
	m.Stage = READY
	m.TotalTime = 0
	m.CurrentBgm = 0
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
				log.Println("room2 step 1 finish!")
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
	for k, v := range m.opt.FakeBooks {
		log.Println("k = ", k, "v = ", v)
		if !m.library.FakeBooks[v] {
			return false
		}
	}
	return true
}

//room3
func (m *Match) ensureCandlesPoweroff() bool {
	for _, v := range m.stairRoom.Candles {
		if v != 0 {
			return false
		}
	}
	return true
}

func (m *Match) ensureCandlesColor() bool {
	for k, v := range m.opt.CandlesColor {
		if m.stairRoom.Candles[k] != v {
			return false
		}
	}
	return true
}

//room4
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
	i := 1
	for _, v := range m.starTower.ConstellationSymbol {
		m.updateStarStatus(i, v)
		i++
	}
	m.starTower.InAnimation = false
}

func (m *Match) updateStarStatus(starNum int, isOpen bool) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("star_led")
	lights := make([]map[string]string, 0)
	leds := make([]map[string]string, 0)
	if isOpen {
		switch starNum {
		case 1:
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
		case 2:
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
		case 3:
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
		case 4:
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
		case 5:
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
		case 6:
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
		case 7:
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
		case 8:
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
		case 9:
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
		case 10:
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
		case 11:
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
		case 12:
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
		case 13:
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
		case 14:
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
		case 15:
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
	sendMsg.Set("useful", "1")
	sendMsg.Set("type_power", strconv.Itoa(symbol))
	m.srv.send(sendMsg, addrs)
}

func (m *Match) ensureElementSymbol() bool {
	for k, v := range m.opt.ElementSymbol {
		if m.endRoom.PowerPoint[k] != v {
			return false
		}
	}
	return true
}

func (m *Match) magicTableAnimation(s string) {
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
		m.magicLab.Table.IsUseful = true
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("magic_table")
		sendMsg.Set("useful", "1")
		sendMsg.Set("time", strconv.FormatFloat(GetOptions().FakeAnimationTime, 'f', 0, 64))
		sendMsg.Set("InAnimation", "1")
		addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-5"}
		m.srv.sendToOne(sendMsg, addr)
	case StageRoom5:
	case StageRoom6:
		time.Sleep(6 * time.Second)
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("magic_table")
		sendMsg.Set("useful", "1")
		sendMsg.Set("time", strconv.FormatFloat(GetOptions().FakeAnimationTime, 'f', 0, 64))
		addr := InboxAddress{InboxAddressTypeDoorArduino, "R-6-7"}
		m.srv.sendToOne(sendMsg, addr)
		time.Sleep(3 * time.Second)
		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("led_candle")
		sendMsg1.Set("mode", "2")
		addr1 := InboxAddress{InboxAddressTypeDoorArduino, "R-6-8"}
		m.srv.sendToOne(sendMsg1, addr1)

		sendMsg2 := NewInboxMessage()
		sendMsg2.SetCmd("water_light")
		sendMsg2.Set("status", "1")
		addr2 := InboxAddress{InboxAddressTypeDoorArduino, "R-6-9"}
		m.srv.sendToOne(sendMsg2, addr2)
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
		m.library.InAnimation = true
		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("door_ctrl")
		sendMsg1.Set("status", "1")
		sendMsg1.Set("time", m.opt.Room2OpenDoorDelayTime)
		log.Println(m.opt.Room2OpenDoorDelayTime)
		addr1 := InboxAddress{InboxAddressTypeDoorArduino, "D-2"}
		m.srv.sendToOne(sendMsg1, addr1)
		m.library.DoorExit = DoorOpen
		m.library.InAnimation = false
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
		sendMsg.Set("time", m.opt.Room2OpenDoorDelayTime)
		addr := InboxAddress{InboxAddressTypeDoorArduino, "D-3"}
		m.srv.sendToOne(sendMsg, addr)
	case StageRoom4:
		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("book_desk")
		sendMsg1.Set("status", "0")
		addr1 := InboxAddress{InboxAddressTypeDoorArduino, "R-4-6"}
		m.srv.sendToOne(sendMsg1, addr1)

		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("door_ctrl")
		sendMsg.Set("status", "1")
		sendMsg.Set("time", m.opt.Room2OpenDoorDelayTime)
		addr := InboxAddress{InboxAddressTypeDoorArduino, "D-4"}
		m.srv.sendToOne(sendMsg, addr)

	case StageRoom5:
		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("magic_rob")
		sendMsg1.Set("status", "1")
		addr1 := InboxAddress{InboxAddressTypeDoorArduino, "R-5-7"}
		m.srv.sendToOne(sendMsg1, addr1)

		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("door_ctrl")
		sendMsg.Set("status", "1")
		sendMsg.Set("time", m.opt.Room2OpenDoorDelayTime)
		addr := InboxAddress{InboxAddressTypeDoorArduino, "D-5"}
		m.srv.sendToOne(sendMsg, addr)
	case StageRoom6:
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("magic_table")
		sendMsg.Set("destroyed", "1")
		addr := InboxAddress{InboxAddressTypeDoorArduino, "R-6-7"}
		m.srv.sendToOne(sendMsg, addr)
		sendMsg1 := NewInboxMessage()
		sendMsg1.SetCmd("led_candle")
		sendMsg1.Set("mode", "0")
		addr1 := InboxAddress{InboxAddressTypeDoorArduino, "R-6-8"}
		m.srv.sendToOne(sendMsg1, addr1)
		sendMsg2 := NewInboxMessage()
		sendMsg2.SetCmd("water_light")
		sendMsg2.Set("status", "0")
		addr2 := InboxAddress{InboxAddressTypeDoorArduino, "R-6-9"}
		m.srv.sendToOne(sendMsg2, addr2)
	}
	//delay Xs
	//control door
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
				}
			case 5:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-10")
					m.library.FakeBooks[2] = true
					m.library.CurrentFakeBookLight++
				}
			case 6:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-11")
					m.library.FakeBooks[3] = true
					m.library.CurrentFakeBookLight++
				}
			case 7:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-12")
					m.library.FakeBooks[4] = true
					m.library.CurrentFakeBookLight++
				}
			case 8:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-13")
					m.library.FakeBooks[5] = true
					m.library.CurrentFakeBookLight++
				}
			case 9:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-14")
					m.library.FakeBooks[6] = true
					m.library.CurrentFakeBookLight++
				}
			case 10:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-15")
					m.library.FakeBooks[7] = true
					m.library.CurrentFakeBookLight++
				}
			case 11:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-16")
					m.library.FakeBooks[8] = true
					m.library.CurrentFakeBookLight++
				}
			case 12:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-17")
					m.library.FakeBooks[9] = true
					m.library.CurrentFakeBookLight++
				}
			case 13:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-18")
					m.library.FakeBooks[10] = true
					m.library.CurrentFakeBookLight++
				}
			case 14:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-19")
					m.library.FakeBooks[11] = true
					m.library.CurrentFakeBookLight++
				}
			case 15:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-20")
					m.library.FakeBooks[12] = true
					m.library.CurrentFakeBookLight++
				}
			case 16:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-21")
					m.library.FakeBooks[13] = true
					m.library.CurrentFakeBookLight++
				}
			case 17:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-22")
					m.library.FakeBooks[14] = true
					m.library.CurrentFakeBookLight++
				}
			case 18:
				if !m.library.Table.IsUseful && !m.library.FakeBooks[magicWords-3] {
					m.srv.fakeBooksControl("2", "1", "R-2-23")
					m.library.FakeBooks[15] = true
					m.library.CurrentFakeBookLight++
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
		} else if m.magicLab.Table.IsUseful && !m.magicLab.Table.IsDestroyed {
			switch magicWords {
			case 3:
				if m.magicLab.Table.IsFinish {
					m.magicLab.Table.IsDestroyed = true
				}
			}
		} else if magicWords == 4 {
			m.magicLab.Stands[0].IsPowerOn = true
			sendMsg.SetCmd("magic_desk")
			sendMsg.Set("useful", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-1"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 5 {
			m.magicLab.Stands[1].IsPowerOn = true
			sendMsg.SetCmd("magic_desk")
			sendMsg.Set("useful", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-2"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 6 {
			m.magicLab.Stands[2].IsPowerOn = true
			sendMsg.SetCmd("magic_desk")
			sendMsg.Set("useful", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-3"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 7 {
			m.magicLab.Stands[3].IsPowerOn = true
			sendMsg.SetCmd("magic_desk")
			sendMsg.Set("useful", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-4-4"}
			m.srv.sendToOne(sendMsg, addr)
		}
		m.magicLab.MagicWords = 0
	case *Room5:
		if magicWords == 1 {
			sendMsg.SetCmd("light_ctrl")
			sendMsg.Set("status", "1")
			addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-5-8"}
			m.srv.sendToOne(sendMsg, addr)
		} else if magicWords == 2 {
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
				for k, _ := range m.starTower.ConstellationSymbol {
					m.starTower.ConstellationSymbol[k] = false
				}
				lights := make([]map[string]string, 0)
				for i := 0; i < 37; i++ {
					m.starTower.ConstellationLight[i] = 0
					light := map[string]string{"light_n": strconv.Itoa(i), "status": strconv.Itoa(m.starTower.ConstellationLight[i])}
					lights = append(lights, light)
				}
				leds := make([]map[string]string, 0)
				for i := 0; i < 33; i++ {
					m.starTower.ConstellationLed[i] = 0
					led := map[string]string{"led_n": strconv.Itoa(i), "mode": strconv.Itoa(m.starTower.ConstellationLed[i])}
					leds = append(leds, led)
				}
				sendMsg.SetCmd("star_led")
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
