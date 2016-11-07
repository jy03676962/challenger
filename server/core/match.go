package core

import (
	"log"
	"strconv"
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

	Stage      string
	TotalTime  float64
	CurrentBgm string

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
		m.gameStage()

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
		switch id {
		case "B-1":
			if m.CurrentBgm != msg.GetStr("MP3") {

			}
		case "L-1":
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

					}
				case 2:
					if m.library.LightStatus != open {

					}
				case 3:
					if m.stairRoom.LightStatus != open {

					}
				case 4:
					if m.magicLab.LightStatus != open {

					}
				case 5:
					if m.starTower.LightStatus != open {

					}
				case 6:
					if m.endRoom.LightStatus != open {

					}
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
			c := []rune(msg.GetStr("BK"))
			var open bool
			for k, v := range c {
				if v == '1' {
					open = true
				} else {
					open = false
				}
				if open != m.library.FakeBooks[k] {
					//不统一
				}
			}
			mode := msg.GetStr("MD")
			if mode == "1" {
				m.library.InAnimation = true
			} else {
				m.library.InAnimation = false
			}
		case "R-2-2":
			c := []rune(msg.GetStr("C"))
			var s int
			for k, v := range c {
				if v == '1' {
					s = 1
				} else {
					s = 0
				}
				if m.library.Candles[k] != s {
					//TODO
				}
			}
		case "R-2-3":
			c := []rune(msg.GetStr("C"))
			var s int
			for k, v := range c {
				if v == '1' {
					s = 1
				} else {
					s = 0
				}
				if m.library.Candles[k+3] != s {
					//TODO
				}
			}
		case "R-2-4":
			c := []rune(msg.GetStr("C"))
			var s int
			for k, v := range c {
				if v == '1' {
					s = 1
				} else {
					s = 0
				}
				if m.library.Candles[k+6] != s {
					//TODO
				}
			}
		case "R-2-5":
			c := []rune(msg.GetStr("C"))
			var s int
			for k, v := range c {
				if v == '1' {
					s = 1
				} else {
					s = 0
				}
				if m.library.Candles[k+9] != s {
					//TODO
				}
			}
		case "R-2-6":
			if msg.GetStr("U") == "1" {
				if m.library.Table.IsUseful != true {
					//TODO
				}
			} else {
				if m.library.Table.IsUseful != false {
					//TODO
				}
			}
			if msg.GetStr("F") == "1" {
				if m.library.Table.IsFinish != true {
					//TODO
				}
			} else {
				if m.library.Table.IsFinish != false {
					//TODO
				}
			}
			if msg.GetStr("D") == "1" {
				if m.library.Table.IsDestroyed != true {
					//TODO
				}
			} else {
				if m.library.Table.IsDestroyed != false {
					//TODO
				}
			}
			m.library.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.library.Table.CurrentAngle, _ = strconv.Atoi(msg.GetStr("A"))
			m.dealAngle()
			m.dealMagicWords(m.library, m.library.MagicWords)
		case "R-2-7":
			if msg.GetStr("ST") == "1" {
				if m.library.MagicBooksLightStatus[0] != true {
					//TODO
				}
			} else {
				if m.library.MagicBooksLightStatus[0] != false {
					//TODO
				}
			}
			if msg.GetStr("LED") == "1" {
				if m.library.MagicBooksLEDStatus[0] != true {
					//TODO
				}
			} else {
				if m.library.MagicBooksLEDStatus[0] != false {
					//TODO
				}
			}
		case "R-2-8":
			if msg.GetStr("ST") == "1" {
				if m.library.MagicBooksLightStatus[1] != true {
					//TODO
				}
			} else {
				if m.library.MagicBooksLightStatus[1] != false {
					//TODO
				}
			}
			if msg.GetStr("LED") == "1" {
				if m.library.MagicBooksLEDStatus[1] != true {
					//TODO
				}
			} else {
				if m.library.MagicBooksLEDStatus[1] != false {
					//TODO
				}
			}
		case "D-2":
			if msg.GetStr("ST") == "1" {
				if m.library.DoorExit != DoorOpen {

				}
			} else {
				if m.library.DoorExit != DoorClose {

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
			if msg.GetStr("U") == "1" {
				if m.stairRoom.Table.IsUseful != true {
					//TODO
				}
			} else {
				if m.stairRoom.Table.IsUseful != false {
					//TODO
				}
			}
			if msg.GetStr("F") == "1" {
				if m.stairRoom.Table.IsFinish != true {
					//TODO
				}
			} else {
				if m.stairRoom.Table.IsFinish != false {
					//TODO
				}
			}
			if msg.GetStr("D") == "1" {
				if m.stairRoom.Table.IsDestroyed != true {
					//TODO
				}
			} else {
				if m.stairRoom.Table.IsDestroyed != false {
					//TODO
				}
			}
			m.library.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.stairRoom, m.stairRoom.MagicWords)
		case "D-3":
			if msg.GetStr("ST") == "1" {
				if m.library.DoorExit != DoorOpen {

				}
			} else {
				if m.library.DoorExit != DoorClose {

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
				m.poweDownAnimation()
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
				m.poweDownAnimation()
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
				m.poweDownAnimation()
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
				m.poweDownAnimation()
			}
		case "R-4-5":
			if msg.GetStr("U") == "1" {
				if m.magicLab.Table.IsUseful != true {
					//TODO
				}
			} else {
				if m.magicLab.Table.IsUseful != false {
					//TODO
				}
			}
			if msg.GetStr("F") == "1" {
				if m.magicLab.Table.IsFinish != true {
					//TODO
				}
			} else {
				if m.magicLab.Table.IsFinish != false {
					//TODO
				}
			}
			if msg.GetStr("D") == "1" {
				if m.magicLab.Table.IsDestroyed != true {
					//TODO
				}
			} else {
				if m.magicLab.Table.IsDestroyed != false {
					//TODO
				}
			}
			m.magicLab.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.magicLab, m.magicLab.MagicWords)

		case "R-4-6":
			if msg.GetStr("ST") == "1" {
				if !m.magicLab.DeskLight {

				}
			} else {
				if m.magicLab.DeskLight {

				}
			}
		case "D-4":
			if msg.GetStr("ST") == "1" {
				if m.magicLab.DoorExit != DoorOpen {

				}
			} else {
				if m.magicLab.DoorExit != DoorClose {

				}
			}
		case "R-5-1":
			c := []rune(msg.GetStr("L"))
			for k, v := range c {
				if v == '1' {
					if m.starTower.ConstellationLight[k] {

					}
				} else {
					if !m.starTower.ConstellationLight[k] {

					}
				}
			}
		case "R-5-2":
			c := []rune(msg.GetStr("L"))
			for k, v := range c {
				if v == '1' {
					if m.starTower.ConstellationLight[k+5] {

					}
				} else {
					if !m.starTower.ConstellationLight[k+5] {

					}
				}
			}
		case "R-5-3":
			c := []rune(msg.GetStr("L"))
			for k, v := range c {
				if v == '1' {
					if m.starTower.ConstellationLight[k+10] {

					}
				} else {
					if !m.starTower.ConstellationLight[k+10] {

					}
				}
			}
		case "R-5-4":
			c := []rune(msg.GetStr("L"))
			for k, v := range c {
				if v == '1' {
					if m.starTower.ConstellationLight[k+15] {

					}
				} else {
					if !m.starTower.ConstellationLight[k+15] {

					}
				}
			}
		case "R-5-5":
			c := []rune(msg.GetStr("L"))
			for k, v := range c {
				if v == '1' {
					if m.starTower.ConstellationLight[k+20] {

					}
				} else {
					if !m.starTower.ConstellationLight[k+20] {

					}
				}
			}
		case "R-5-6":
			if msg.GetStr("U") == "1" {
				if m.starTower.Table.IsUseful != true {
					//TODO
				}
			} else {
				if m.starTower.Table.IsUseful != false {
					//TODO
				}
			}
			if msg.GetStr("F") == "1" {
				if m.starTower.Table.IsFinish != true {
					//TODO
				}
			} else {
				if m.starTower.Table.IsFinish != false {
					//TODO
				}
			}
			if msg.GetStr("D") == "1" {
				if m.starTower.Table.IsDestroyed != true {
					//TODO
				}
			} else {
				if m.starTower.Table.IsDestroyed != false {
					//TODO
				}
			}
			m.magicLab.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.magicLab, m.magicLab.MagicWords)
		case "R-5-7":
			if msg.GetStr("ST") == "1" {
				if m.starTower.DoorMagicRod != DoorOpen {
					//TODO
				}
			} else {
				if m.starTower.DoorMagicRod != DoorClose {
					//TODO
				}
			}
		case "D-5":
			if msg.GetStr("ST") == "1" {
				if m.starTower.DoorExit != DoorOpen {

				}
			} else {
				if m.starTower.DoorExit != DoorClose {

				}
			}
		case "R-6-1":
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				//TODO
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[0] {
				//TODO
			}
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[0] = ty
			}
		case "R-6-2":
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				//TODO
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[1] {
				//TODO
			}
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[1] = ty
			}
		case "R-6-3":
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				//TODO
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[2] {
				//TODO
			}
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[2] = ty
			}
		case "R-6-4":
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				//TODO
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[3] {
				//TODO
			}
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[4] = ty
			}
		case "R-6-5":
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				//TODO
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[4] {
				//TODO
			}
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[5] = ty
			}
		case "R-6-6":
			ty, _ := strconv.Atoi(msg.GetStr("TY"))
			if ty != m.endRoom.CurrentSymbol {
				//TODO
			}
			useful, _ := strconv.Atoi(msg.GetStr("U"))
			if useful != m.endRoom.PowerPointUseful[5] {
				//TODO
			}
			if msg.GetStr("F") == "1" {
				m.endRoom.PowerPoint[6] = ty
			}
		case "R-6-7":
			if msg.GetStr("U") == "1" {
				if m.endRoom.Table.IsUseful != true {
					//TODO
				}
			} else {
				if m.endRoom.Table.IsUseful != false {
					//TODO
				}
			}
			if msg.GetStr("F") == "1" {
				if m.endRoom.Table.IsFinish != true {
					//TODO
				}
			} else {
				if m.endRoom.Table.IsFinish != false {
					//TODO
				}
			}
			if msg.GetStr("D") == "1" {
				if m.endRoom.Table.IsDestroyed != true {
					//TODO
				}
			} else {
				if m.endRoom.Table.IsDestroyed != false {
					//TODO
				}
			}
			m.endRoom.MagicWords, _ = strconv.Atoi(msg.GetStr("W"))
			m.dealMagicWords(m.magicLab, m.endRoom.MagicWords)
		case "R-6-8":
			c := []rune(msg.GetStr("C"))
			for k, v := range c {
				if v == '1' {
					if m.endRoom.Candles[k] != 1 {
						//TODO
					}
				} else {
					if m.endRoom.Candles[k] != 0 {
						//TODO
					}
				}
			}
		case "R-6-9":
			if msg.GetStr("ST") == "1" {
				if !m.endRoom.WaterLight {

				}
			} else {
				if m.endRoom.WaterLight {

				}
			}
		case "D-6":
			if msg.GetStr("ST") == "1" {
				if m.endRoom.DoorExit != DoorOpen {

				}
			} else {
				if m.endRoom.DoorExit != DoorClose {

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

func (m *Match) gameStage() {
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
			if m.library.MagicBooksLEDStatus[0] != true {
				m.library.MagicBooksLEDStatus[0] = true
				//TODO
			}
			if m.library.MagicBooksLEDStatus[1] != true {
				m.library.MagicBooksLEDStatus[1] = true
				//TODO
			}
			if m.fakeActNum() == 5 {
				if m.ensureFakeBooks() {
					m.library.InAnimation = true
					m.fakeBooksAnimation()
					m.library.CurrentFakeBookLight = 15
					m.library.Table.IsUseful = true
					m.library.Table.MarkAngle = m.library.Table.CurrentAngle
					m.library.Step = 2
					log.Println("room2 step 1 finish!")
				} else {
					m.library.CurrentFakeBookLight = 0
					m.fakeBooksErrorAnimation()
				}
			}
		} else if m.library.Step == 2 {
			if m.ensurePowerPointOrder() {
				m.library.Table.IsFinish = true
				m.library.Step = 3
				log.Println("room2 step 2 finish!")
			}
		} else if m.library.Step == 3 {
			if m.library.Table.IsDestroyed {
				m.library.InAnimation = true
				m.endingAnimation(StageRoom2)
				m.library.DoorExit = DoorOpen
				log.Println("room2 step 3 finish!")
			}
		}
	case StageRoom3:
		if m.stairRoom.Step == 1 {
			if m.stairRoom.CurrentCandlesLight == 0 {
				m.stairRoom.InAnimation = true
				m.magicTableAnimation(StageRoom3)
				m.stairRoom.Table.IsUseful = true
				m.stairRoom.Step = 2
				log.Println("room3 step 1 finish!")
			}
		} else if m.stairRoom.Step == 2 {
			if m.stairRoom.CurrentCandlesLight == 6 {
				if m.ensureCandlesColor() {
					m.stairRoom.Table.IsFinish = true
					m.stairRoom.Step = 3
					log.Println("room3 step 2 finish!")
				}
			}
		} else if m.stairRoom.Step == 3 {
			if m.stairRoom.Table.IsDestroyed {
				m.stairRoom.InAnimation = true
				m.endingAnimation(StageRoom3)
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
				m.endingAnimation(StageRoom4)
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
				m.endingAnimation(StageRoom5)
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
				m.endingAnimation(StageRoom6)
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
func (m *Match) fakeBooksAnimation() {

}

func (m *Match) fakeBooksErrorAnimation() {

}

func (m *Match) dealAngle() {
	if !m.library.Table.IsUseful {
		return
	}
}

func (m *Match) ensureFakeBooks() bool {
	for _, v := range m.opt.FakeBooks {
		if !m.library.FakeBooks[v] {
			return false
		}
	}
	return true
}

func (m *Match) ensurePowerPointOrder() bool {
	for k, v := range m.opt.PowerPoints {
		if m.library.Table.ButtonStatus[v] != k {
			return false
		}
	}
	return true
}

//room3
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

func (m *Match) poweringAnimation() {

}

func (m *Match) poweDownAnimation() {

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

//room6 animation
func (m *Match) amMagicAnimation() {

}

func (m *Match) voicePlay(step int) {
	switch step {
	case 2:
	case 3:
	}

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

func (m *Match) endingAnimation(s string) {
	switch s {
	case StageRoom2:
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
	switch room.(type) {
	case *Room2:
		if m.library.Table.IsUseful && !m.library.Table.IsDestroyed {

		}
	case *Room3:
	case *Room4:
	case *Room5:
	case *Room6:
	}

}
