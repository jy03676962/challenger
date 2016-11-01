package core

import (
	"log"
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

	Stage     string
	TotalTime float64

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
	switch cmd {
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
			if m.library.CurrentFakeBookLight == 5 {
				if m.ensureFakeBooks() {
					m.library.InAnimation = true
					m.fakeBooksAnimation()
					m.library.CurrentFakeBookLight = 15
					m.library.Table.IsUseful = true
					m.library.Step = 2
					log.Println("room2 step 1 finish!")
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
				m.voicePlay()
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

func (m *Match) fakeBooksAnimation() {

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

func (m *Match) voicePlay() {

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
}
