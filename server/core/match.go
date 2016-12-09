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
	StageRoom1 = "ready"
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
	CurrentBgm        string

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
	m.Stage = StageRoom1
	m.srv = s
	m.opt = GetOptions()
	m.TotalTime = 0 //config
	m.OpenDoorDelayTime = GetOptions().OpenDoorDelayTime
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
		case "B-1":
			if m.CurrentBgm != msg.GetStr("MP3") {
				m.srv.bgmControl(m.CurrentBgm)
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
			mode := msg.GetStr("MD")
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("fake_book")
			sendMsg.Set("time", strconv.FormatFloat(m.opt.FakeAnimationTime, 'f', 0, 64))
			books := make([]map[string]string, 0)
			if mode == "0" {
				sendMsg.Set("mode", "0")
				c := []rune(msg.GetStr("BK"))
				var open bool
				for k, v := range c {
					if v == '1' {
						open = true
					} else {
						open = false
					}
					if open != m.library.FakeBooks[k] {
						if m.library.FakeBooks[k] {
							books = append(books, map[string]string{
								"book_n": strconv.Itoa(k),
								"book_m": "1",
							})
						} else {
							books = append(books, map[string]string{
								"book_n": strconv.Itoa(k),
								"book_m": "0",
							})
						}
					}
				}
				if len(books) > 0 {
					sendMsg.Set("book", books)
					m.srv.sendToOne(sendMsg, addr)
				}
			}
			if mode == "1" {
				m.library.InAnimation = true
			} else {
				m.library.InAnimation = false
			}
		case "R-2-2":
			c := []rune(msg.GetStr("C"))
			candles := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("led_candle")
			for k, v := range c {
				if m.library.Candles[k] != int(v-'0') {
					candles = append(candles, map[string]string{
						"candle": strconv.Itoa(k),
						"color":  strconv.Itoa(m.library.Candles[k]),
					})
				}
			}
			if len(candles) > 0 {
				sendMsg.Set("candles", candles)
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-2-3":
			c := []rune(msg.GetStr("C"))
			candles := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("led_candle")
			for k, v := range c {
				if m.library.Candles[k] != int(v-'0') {
					candles = append(candles, map[string]string{
						"candle": strconv.Itoa(k + 3),
						"color":  strconv.Itoa(m.library.Candles[k+3]),
					})
				}
			}
			if len(candles) > 0 {
				sendMsg.Set("candles", candles)
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-2-4":
			c := []rune(msg.GetStr("C"))
			candles := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("led_candle")
			for k, v := range c {
				if m.library.Candles[k] != int(v-'0') {
					candles = append(candles, map[string]string{
						"candle": strconv.Itoa(k + 6),
						"color":  strconv.Itoa(m.library.Candles[k+6]),
					})
				}
			}
			if len(candles) > 0 {
				sendMsg.Set("candles", candles)
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-2-5":
			c := []rune(msg.GetStr("C"))
			candles := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("led_candle")
			for k, v := range c {
				if m.library.Candles[k] != int(v-'0') {
					candles = append(candles, map[string]string{
						"candle": strconv.Itoa(k + 9),
						"color":  strconv.Itoa(m.library.Candles[k+9]),
					})
				}
			}
			if len(candles) > 0 {
				sendMsg.Set("candles", candles)
				m.srv.sendToOne(sendMsg, addr)
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
			if msg.GetStr("D") == "1" {
				if m.library.Table.IsDestroyed != true {
					sendMsg.Set("destroyed", "0")
				}
			} else {
				if m.library.Table.IsDestroyed != false {
					sendMsg.Set("destroyed", "1")
				}
			}
			if len(sendMsg.Data) > 2 {
				m.srv.sendToOne(sendMsg, addr)
			}
			m.library.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.library.Table.CurrentAngle, _ = strconv.ParseFloat(msg.GetStr("A"), 64)
			m.dealAngle()
			if !m.library.InAnimation {
				m.dealMagicWords(m.library, m.library.MagicWords)
			}
		case "R-2-7":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_book")
			if msg.GetStr("ST") == "1" {
				if m.library.MagicBooksLightStatus[0] != true {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.library.MagicBooksLightStatus[0] != false {
					sendMsg.Set("status", "1")
					m.srv.sendToOne(sendMsg, addr)
				}
			}
		case "R-2-8":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_book")
			if msg.GetStr("ST") == "1" {
				if m.library.MagicBooksLightStatus[1] != true {
					sendMsg.Set("status", "0")
					m.srv.sendToOne(sendMsg, addr)
				}
			} else {
				if m.library.MagicBooksLightStatus[1] != false {
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
			if msg.GetStr("D") == "1" {
				if m.stairRoom.Table.IsDestroyed != true {
					sendMsg.Set("destroyed", "0")
				}
			} else {
				if m.stairRoom.Table.IsDestroyed != false {
					sendMsg.Set("destroyed", "1")
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
			if msg.GetStr("USF") == "1" {
				m.magicLab.Stands[0].IsPowerOn = true
			} else {
				m.magicLab.Stands[0].IsPowerOn = false
			}
			if msg.GetStr("P") == "1" {
				m.poweringAnimation()
			} else if msg.GetStr("P") == "2" {
				m.magicLab.Stands[0].IsPowerful = true
				m.powerDownAnimation()
			}
		case "R-4-2":
			if msg.GetStr("USF") == "1" {
				m.magicLab.Stands[1].IsPowerOn = true
			} else {
				m.magicLab.Stands[1].IsPowerOn = false
			}
			if msg.GetStr("P") == "1" {
				m.poweringAnimation()
			} else if msg.GetStr("P") == "2" {
				m.magicLab.Stands[1].IsPowerful = true
				m.powerDownAnimation()
			}
		case "R-4-3":
			if msg.GetStr("USF") == "1" {
				m.magicLab.Stands[2].IsPowerOn = true
			} else {
				m.magicLab.Stands[2].IsPowerOn = false
			}
			if msg.GetStr("P") == "1" {
				m.poweringAnimation()
			} else if msg.GetStr("P") == "2" {
				m.magicLab.Stands[2].IsPowerful = true
				m.powerDownAnimation()
			}
		case "R-4-4":
			if msg.GetStr("USF") == "1" {
				m.magicLab.Stands[3].IsPowerOn = true
			} else {
				m.magicLab.Stands[3].IsPowerOn = false
			}
			if msg.GetStr("P") == "1" {
				m.poweringAnimation()
			} else if msg.GetStr("P") == "2" {
				m.magicLab.Stands[3].IsPowerful = true
				m.powerDownAnimation()
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
			if msg.GetStr("D") == "1" {
				if m.magicLab.Table.IsDestroyed != true {
					sendMsg.Set("destroyed", "0")
				}
			} else {
				if m.magicLab.Table.IsDestroyed != false {
					sendMsg.Set("destroyed", "1")
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
			c := []rune(msg.GetStr("L"))
			lights := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("star_led")
			for k, v := range c {
				if v == '1' {
					if !m.starTower.ConstellationLight[k] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k),
							"status":  "0",
						})
					}
				} else {
					if m.starTower.ConstellationLight[k] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k),
							"status":  "1",
						})

					}
				}
			}
			if len(lights) > 0 {
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-5-2":
			c := []rune(msg.GetStr("L"))
			lights := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("star_led")
			for k, v := range c {
				if v == '1' {
					if !m.starTower.ConstellationLight[k+9] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k + 9),
							"status":  "0",
						})
					}
				} else {
					if m.starTower.ConstellationLight[k+9] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k + 9),
							"status":  "1",
						})
					}
				}
			}
			if len(lights) > 0 {
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-5-3":
			c := []rune(msg.GetStr("L"))
			lights := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("star_led")
			for k, v := range c {
				if v == '1' {
					if !m.starTower.ConstellationLight[k+18] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k + 18),
							"status":  "0",
						})
					}
				} else {
					if m.starTower.ConstellationLight[k+18] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k + 18),
							"status":  "1",
						})
					}
				}
			}
			if len(lights) > 0 {
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-5-4":
			c := []rune(msg.GetStr("L"))
			lights := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("star_led")
			for k, v := range c {
				if v == '1' {
					if !m.starTower.ConstellationLight[k+23] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k + 23),
							"status":  "0",
						})
					}
				} else {
					if m.starTower.ConstellationLight[k+23] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k + 23),
							"status":  "1",
						})
					}
				}
			}
			if len(lights) > 0 {
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-5-5":
			c := []rune(msg.GetStr("L"))
			lights := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("star_led")
			for k, v := range c {
				if v == '1' {
					if !m.starTower.ConstellationLight[k+29] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k + 29),
							"status":  "0",
						})
					}
				} else {
					if m.starTower.ConstellationLight[k+29] {
						lights = append(lights, map[string]string{
							"light_n": strconv.Itoa(k + 29),
							"status":  "1",
						})
					}
				}
			}
			if len(lights) > 0 {
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)
			}
		case "R-5-6":
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_table")
			if msg.GetStr("U") == "1" {
				if m.starTower.Table.IsUseful != true {
					sendMsg.Set("useful", "0")
				}
			} else {
				if m.starTower.Table.IsUseful != false {
					sendMsg.Set("useful", "1")
				}
			}
			if msg.GetStr("F") == "1" {
				if m.starTower.Table.IsFinish != true {
					sendMsg.Set("finish", "0")
				}
			} else {
				if m.starTower.Table.IsFinish != false {
					sendMsg.Set("finish", "1")
				}
			}
			if msg.GetStr("D") == "1" {
				if m.starTower.Table.IsDestroyed != true {
					sendMsg.Set("destroyed", "0")
				}
			} else {
				if m.starTower.Table.IsDestroyed != false {
					sendMsg.Set("destroyed", "1")
				}
			}
			if len(sendMsg.Data) > 2 {
				m.srv.sendToOne(sendMsg, addr)
			}
			star, _ := strconv.Atoi(msg.GetStr("S"))
			m.dealStar(star)
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
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[0] = m.endRoom.CurrentSymbol
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
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[1] = m.endRoom.CurrentSymbol
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
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[2] = m.endRoom.CurrentSymbol
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
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[4] = m.endRoom.CurrentSymbol
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
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[5] = m.endRoom.CurrentSymbol
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
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[6] = m.endRoom.CurrentSymbol
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
			candleMode := msg.GetStr("mode")
			if candleMode != "1" {
				return
			}
			c := []rune(msg.GetStr("C"))
			candles := make([]map[string]string, 0)
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("led_candle")
			for k, v := range c {
				if int(v-'0') != m.endRoom.Candles[k] {
					candles = append(candles, map[string]string{
						"candle": strconv.Itoa(k),
						"color":  strconv.Itoa(m.endRoom.Candles[k]),
					})
				}
			}
			if len(candles) > 0 {
				sendMsg.Set("candle", candles)
				m.srv.sendToOne(sendMsg, addr)
			}
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
		case StageRoom1:
			m.setStage(StageRoom2)
		case StageRoom2:
			if m.library.Step < 3 {
				m.library.Step++
			} else {
				m.setStage(StageRoom3)
			}
		case StageRoom3:
			if m.stairRoom.Step < 3 {
				m.library.Step++
			} else {
				m.setStage(StageRoom4)
			}
		case StageRoom4:
			if m.magicLab.Step < 3 {
				m.library.Step++
			} else {
				m.setStage(StageRoom5)
			}
		case StageRoom5:
			if m.starTower.Step < 3 {
				m.library.Step++
			} else {
				m.setStage(StageRoom6)
			}
		case StageRoom6:
			if m.endRoom.Step < 3 {
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
	case StageRoom2:
	case StageRoom3:
	case StageRoom4:
	case StageRoom5:
	case StageRoom6:
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
	case StageRoom1:
		if m.livingRoom.DoorMirror == DoorOpen {
			m.room1Animation()
			log.Println("room 1 finish!")
		}
	case StageRoom2:
		if m.library.Step == 1 {
			if m.fakeActNum() == 5 {
				if m.ensureFakeBooks() {
					m.fakeBooksAnimation(dt)
				} else {
					m.fakeBooksErrorAnimation()
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
				m.stairRoom.InAnimation = true
				m.magicTableAnimation(StageRoom3)
				m.stairRoom.Table.IsUseful = true
				m.stairRoom.Step = 2
				log.Println("room3 step 1 finish!")
			}
		} else if m.stairRoom.Step == 2 {
			if m.ensureCandlesColor() {
				m.stairRoom.Table.IsFinish = true
				m.stairRoom.Step = 3
				log.Println("room3 step 2 finish!")
			}
		} else if m.stairRoom.Step == 3 {
			if m.stairRoom.Table.IsDestroyed {
				m.stairRoom.InAnimation = true
				m.endingAnimation(StageRoom3, dt)
				m.stairRoom.DoorExit = DoorOpen
				log.Println("room3 step 3 finish!")
			}
		}
	case StageRoom4:
		if m.magicLab.Step == 1 {
			if m.ensureMagicStandsPowerOn() {
				m.magicLab.InAnimation = true
				m.magicTableAnimation(StageRoom4)
				m.magicLab.Table.IsUseful = true
				m.magicLab.Step = 2
				log.Println("room4 step 1 finish!")
			}
		} else if m.magicLab.Step == 2 {
			if m.ensureMagicStandsPowerFul() {
				m.magicLab.Table.IsFinish = true
				m.magicLab.Step = 3
				log.Println("room4 step 2 finish!")

			}
		} else if m.magicLab.Step == 3 {
			if m.magicLab.Table.IsDestroyed {
				m.magicLab.InAnimation = true
				m.endingAnimation(StageRoom4, dt)
				m.magicLab.DoorExit = DoorOpen
				log.Println("room4 step 3 finish!")
			}
		}
	case StageRoom5:
		if m.starTower.Step == 1 {
			if m.starTower.CurrentConstellationLight == 3 {
				m.starTower.InAnimation = true
				m.magicTableAnimation(StageRoom5)
				m.starTower.Table.IsUseful = true
				m.starTower.Step = 2
				log.Println("room 5 step 1 finish!")
			}
		} else if m.starTower.Step == 2 {
			if m.ensureConstellationSymbol() {
				m.starTower.Table.IsFinish = true
				m.starTower.Step = 3
				log.Println("room 5 step 2 finish!")
			}
		} else if m.starTower.Step == 3 {
			if m.starTower.Table.IsDestroyed {
				m.starTower.InAnimation = true
				m.endingAnimation(StageRoom5, dt)
				m.starTower.DoorExit = DoorOpen
				m.starTower.DoorMagicRod = DoorOpen
				log.Println("room 5 step 3 finish!")
			}

		}
	case StageRoom6:
		if m.endRoom.Step == 1 {
			if m.endRoom.NextStep == 2 {
				m.endRoom.InAnimation = true
				m.amMagicAnimation()
				m.endRoom.Step = 2
				log.Println("room 6 step 1 finish!")
			}
		} else if m.endRoom.Step == 2 {
			if m.exitRoom.ButtonNextStage { //endroom 数据维护需要锁
				m.endRoom.InAnimation = true
				m.voicePlay(m.endRoom.Step)
				m.endRoom.Table.IsUseful = true
				m.endRoom.InAnimation = true
				m.magicTableAnimation(StageRoom6)
				m.endRoom.Step = 3
				m.endRoom.NextStep = 3
				log.Println("room 6 step 2 finish!")
			}
		} else if m.endRoom.Step == 3 {
			if m.ensureElementSymbol() {
				m.endRoom.Table.IsFinish = true
				m.voicePlay(m.endRoom.Step)
			} else {
				m.endRoom.Table.IsFinish = false
			}
			if m.endRoom.Table.IsFinish && m.endRoom.NextStep == 4 {
				m.endRoom.InAnimation = true
				m.endingAnimation(StageRoom6, dt)
				m.endRoom.Table.IsDestroyed = true
				m.endRoom.DoorExit = DoorOpen
				log.Println("room 6 step 3 finish!")
			}
		}
	case StageEnd:
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
	initHardware()
	m.Stage = StageRoom1
	m.TotalTime = 0
	log.Println("game reset success!")
}

func initHardware() {

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
		m.library.InAnimation = true
		m.library.FakeAnimationStep = 1
		books := make([]map[string]string, 5)
		num := 0
		for k, v := range m.library.FakeBooks {
			if v {
				books[num] = map[string]string{"book_n": strconv.Itoa(k), "book_m": "0"}
				num++
			}
		}
		m.srv.fbControls(books, "3")
		addrs := []InboxAddress{
			{InboxAddressTypeRoomArduinoDevice, "R-2-7"},
			{InboxAddressTypeRoomArduinoDevice, "R-2-8"},
		}
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("magic_book")
		sendMsg.Set("status", "0")
		m.srv.send(sendMsg, addrs)
		//TODO
		//send step 1 灯光全灭，书发出3语音后全灭
	} else {
		m.library.FakeAnimationTime = math.Max(m.library.FakeAnimationTime-sec, 0)
		if m.library.FakeAnimationTime == 0 {
			m.library.FakeAnimationStep++
			switch m.library.FakeAnimationStep {
			case 2: //第一组开始 1
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("fake_book")
				sendMsg.Set("mode", "3")
				sendMsg.Set("time", strconv.FormatFloat(GetOptions().FakeAnimationTime, 'f', 0, 64))
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-2-1"}
				m.srv.sendToOne(sendMsg, addr)
				m.library.FakeAnimationTime = opt.FakeAnimationTime
			case 3: //第二组 2,3 并且蜡烛、灯箱开始
				//TODO
				m.library.FakeAnimationTime = opt.FakeAnimationTime
			case 4: //第三组 4,5,6
				m.library.Table.IsUseful = true
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("magic_table")
				sendMsg.Set("useful", "1")
				sendMsg.Set("time", strconv.FormatFloat(GetOptions().FakeAnimationTime, 'f', 0, 64))
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-2-1"}
				m.srv.sendToOne(sendMsg, addr)
				m.library.FakeAnimationTime = opt.FakeAnimationTime
			case 5: //第四组 7,8,9,10//记忆水晶
				m.library.FakeAnimationTime = opt.FakeAnimationTime
			case 6: //第五组 11,12,13,14,15 蜡烛、灯箱换颜色
				//TODO
				m.library.FakeAnimationTime = opt.FakeAnimationTime
			case 7:
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("fake_book")
				sendMsg.Set("mode", "1")
				addr := InboxAddress{InboxAddressTypeRoomArduinoDevice, "R-2-1"}
				m.srv.sendToOne(sendMsg, addr)
				m.library.InAnimation = false
				m.library.CurrentFakeBookLight = 15
				m.library.Table.MarkAngle = m.library.Table.CurrentAngle
				m.library.Step = 2
				log.Println("room2 step 1 finish!")

			}
		}
	}

}

func (m *Match) fakeBooksErrorAnimation() {
	books := make([]map[string]string, 5)
	num := 0
	m.library.CurrentFakeBookLight = 0
	for k, v := range m.library.FakeBooks {
		if v {
			books[num] = map[string]string{"book_n": strconv.Itoa(k), "book_m": "0"}
			num++
		}
	}
	m.srv.fbControls(books, "2")
	for i := 0; i < 15; i++ {
		m.library.FakeBooks[i] = false
	}

}

func (m *Match) dealAngle() {
	if !m.library.Table.IsUseful {
		return
	}
	addrs := []InboxAddress{{InboxAddressTypeRoomArduinoDevice, "R-2-2"}, {InboxAddressTypeRoomArduinoDevice, "R-2-3"}, {InboxAddressTypeRoomArduinoDevice, "R-2-4"}, {InboxAddressTypeRoomArduinoDevice, "R-2-5"}}
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("led_candle")
	sendMsg.Set("angle", strconv.FormatFloat(m.library.Table.CurrentAngle-m.library.Table.MarkAngle, 'f', -1, 64))
	m.srv.send(sendMsg, addrs)
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
	for _, v := range m.opt.Constellations {
		if !m.starTower.ConstellationSymbol[v] {
			return false
		}
	}
	return true
}

func (m *Match) starControl(starNum int, isOpen bool) { //TODO
	if isOpen {
		switch starNum {
		case 1:
		case 2:
		case 3:
		case 4:
		case 5:
		case 6:
		case 7:
		case 8:
		case 9:
		case 10:
		case 11:
		case 12:
		case 13:
		case 14:
		case 15:
		}
	} else {
		switch starNum {
		case 1:
		case 2:
		case 3:
		case 4:
		case 5:
		case 6:
		case 7:
		case 8:
		case 9:
		case 10:
		case 11:
		case 12:
		case 13:
		case 14:
		case 15:
		}

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
}

//room6 animation
func (m *Match) amMagicAnimation() {

}

func (m *Match) voicePlay(step int) {
	switch step {
	case 2:
	case 3:
	}

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
	case StageRoom4:
	case StageRoom5:
	case StageRoom6:
	}
}

func (m *Match) endingAnimation(s string, dt time.Duration) {
	sec := dt.Seconds()
	switch s {
	case StageRoom2:
		if m.library.InAnimation != true {
			//animation
			addrs := []InboxAddress{
				{InboxAddressTypeRoomArduinoDevice, "R-2-7"},
				{InboxAddressTypeRoomArduinoDevice, "R-2-8"},
			}
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("magic_book")
			sendMsg.Set("status", "0")
			m.srv.send(sendMsg, addrs)

			sendMsg1 := NewInboxMessage()
			sendMsg1.SetCmd("fake_book")
			sendMsg1.Set("mode", "0")
			addr := InboxAddress{InboxAddressTypeDoorArduino, "R-2-1"}
			m.srv.sendToOne(sendMsg1, addr)
			//TODO 蜡烛
			m.library.InAnimation = true
		}
		m.OpenDoorDelayTime = math.Max(m.OpenDoorDelayTime-sec, 0)
		if m.OpenDoorDelayTime == 0 {
			sendMsg := NewInboxMessage()
			sendMsg.SetCmd("door_ctrl")
			sendMsg.Set("status", "1")
			addr := InboxAddress{InboxAddressTypeDoorArduino, "D-2"}
			m.srv.sendToOne(sendMsg, addr)
			m.library.DoorExit = 1
			m.library.InAnimation = false
			m.OpenDoorDelayTime = m.opt.OpenDoorDelayTime
			log.Println("room2 finish!")

		}
	case StageRoom3:
	case StageRoom4:
	case StageRoom5:
	case StageRoom6:
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
			m.library.MagicBooksLightStatus[0] = true
			m.library.MagicBooksLightStatus[1] = true
			sendMsg.SetCmd("magic_book")
			sendMsg.Set("status", "1")
			addrs := []InboxAddress{{InboxAddressTypeRoomArduinoDevice, "R-2-7"}, {InboxAddressTypeRoomArduinoDevice, "R-2-8"}}
			m.srv.send(sendMsg, addrs)
		} else if magicWords == 2 {
			m.library.MagicBooksLightStatus[0] = false
			m.library.MagicBooksLightStatus[1] = false
			sendMsg.SetCmd("magic_book")
			sendMsg.Set("status", "0")
			addrs := []InboxAddress{{InboxAddressTypeRoomArduinoDevice, "R-2-7"}, {InboxAddressTypeRoomArduinoDevice, "R-2-8"}}
			m.srv.send(sendMsg, addrs)
		} else if !m.library.Table.IsDestroyed {
			switch magicWords {
			case 3:
				if m.library.Table.IsFinish {
					m.library.Table.IsDestroyed = true
				}
			case 4:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("1", "1", "R-2-1")
					m.library.FakeBooks[1] = true
					m.library.CurrentFakeBookLight++
				}
			case 5:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("2", "1", "R-2-1")
					m.library.FakeBooks[2] = true
					m.library.CurrentFakeBookLight++
				}
			case 6:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("3", "1", "R-2-1")
					m.library.FakeBooks[3] = true
					m.library.CurrentFakeBookLight++
				}
			case 7:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("4", "1", "R-2-1")
					m.library.FakeBooks[4] = true
					m.library.CurrentFakeBookLight++
				}
			case 8:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("5", "1", "R-2-1")
					m.library.FakeBooks[5] = true
					m.library.CurrentFakeBookLight++
				}
			case 9:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("6", "1", "R-2-1")
					m.library.FakeBooks[6] = true
					m.library.CurrentFakeBookLight++
				}
			case 10:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("7", "1", "R-2-1")
					m.library.FakeBooks[7] = true
					m.library.CurrentFakeBookLight++
				}
			case 11:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("8", "1", "R-2-1")
					m.library.FakeBooks[8] = true
					m.library.CurrentFakeBookLight++
				}
			case 12:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("9", "1", "R-2-1")
					m.library.FakeBooks[9] = true
					m.library.CurrentFakeBookLight++
				}
			case 13:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("10", "1", "R-2-1")
					m.library.FakeBooks[10] = true
					m.library.CurrentFakeBookLight++
				}
			case 14:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("11", "1", "R-2-1")
					m.library.FakeBooks[11] = true
					m.library.CurrentFakeBookLight++
				}
			case 15:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("12", "1", "R-2-1")
					m.library.FakeBooks[12] = true
					m.library.CurrentFakeBookLight++
				}
			case 16:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("13", "1", "R-2-1")
					m.library.FakeBooks[13] = true
					m.library.CurrentFakeBookLight++
				}
			case 17:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("14", "1", "R-2-1")
					m.library.FakeBooks[14] = true
					m.library.CurrentFakeBookLight++
				}
			case 18:
				if !m.library.Table.IsUseful {
					m.srv.fakeBooksControl("15", "1", "R-2-1")
					m.library.FakeBooks[15] = true
					m.library.CurrentFakeBookLight++
				}
			}
			m.library.MagicWords = 0
		}
	case *Room3:
		if m.stairRoom.Table.IsUseful && !m.stairRoom.Table.IsDestroyed {
			switch magicWords {
			case 3:
				if m.stairRoom.Table.IsFinish {
					m.stairRoom.Table.IsDestroyed = true
				}
			}
		}
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
		}
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
			}
		}
	case *Room6:
		if m.endRoom.Table.IsUseful && !m.endRoom.Table.IsDestroyed {
			switch magicWords {
			case 3:
			case 20:
			}
		}
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
