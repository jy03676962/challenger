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
	magicLab     *Room3
	starTower    *Room4
	endRoom      *Room5
	exitRoom     *ExitRoom
}

func NewMatch(s *Srv) *Match {
	m := Match{}
	m.Stage = "ready"
	m.srv = s
	m.opt = GetOptions()
	m.TotalTime = 3600000.00 //config
	m.msgCh = make(chan *InboxMessage, 1000)
	m.closeCh = make(chan bool)
	m.InitHardwareData()
	return &m

}

func (m *Match) InitHardwareData() {
	m.livingRoom = NewRoom1()
	m.library = NewRoom2()
	m.magicLab = NewRoom3()
	m.starTower = NewRoom4()
	m.endRoom = NewRoom5()
	m.entranceRoom = NewEntranceRoom()
	m.exitRoom = NewExitRoom()

}

func (m *Match) Run() {
	dt := 33 * time.Millisecond
	tickChan := time.Tick(dt)
	for {
		<-tickChan
		m.handleInputs()
		if m.Stage == "end" || m.Stage == "maintenance" {
			break
		}

	}
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
	case "A":
	case "B":
	}
}
