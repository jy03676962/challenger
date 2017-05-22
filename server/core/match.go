package core

import (
	"log"
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

type MatchEvent struct {
	Type MatchEventType
	ID   uint
	Data interface{}
}

type Match struct {
	opt *MatchOptions
	srv *Srv

	Step              int
	TotalTime         float64
	OpenDoorDelayTime float64
	CurrentBgm        int

	msgCh   chan *InboxMessage
	closeCh chan bool
}

func NewMatch(s *Srv) *Match {
	m := Match{}
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

}

func (m *Match) setStage(s string) {

}

func (m *Match) gameStage(dt time.Duration) {

}

func (m *Match) updateStage() {

}

func (m *Match) bgmPlay(bgm int) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("mp3_ctrl")
	sendMsg.Set("music", strconv.Itoa(bgm))
	addr := InboxAddress{InboxAddressTypeMusicArduino, "B-1"}
	m.srv.sendToOne(sendMsg, addr)

}

func att(id string) InboxAddressType {
	if id == "" {
		return InboxAddressTypeUnknown
	} else if strings.HasPrefix(id, "R") {
		return InboxAddressTypeGameArduinoDevice
	} else if strings.HasPrefix(id, "L") {
		return InboxAddressTypeLightArduinoDevice
	} else if strings.HasPrefix(id, "B") {
		return InboxAddressTypeMusicArduino
	} else if strings.HasPrefix(id, "D") {
		return InboxAddressTypeDoorArduino
	}
	return InboxAddressTypeUnknown
}
