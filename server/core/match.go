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
	EventToDay = iota + 1
	EventToNight
	EventChallengeBilly
	EventRecoverDay
	EventRecoverNight
	EventRobBar
	CancleRobBar
)

type MatchEvent struct {
	Type MatchEventType
	ID   uint
	Data interface{}
}

type Match struct {
	opt *MatchOptions
	srv *Srv

	Event   int
	IsGoing bool

	LapseTime   float64
	TotalStep   int
	CurrentStep int
	CurrentBgm  int

	msgCh   chan *InboxMessage
	closeCh chan bool
}

func NewMatch(s *Srv, event int) *Match {
	m := Match{}
	m.CurrentBgm = 0
	m.srv = s
	m.opt = GetOptions()
	m.Event = event
	m.IsGoing = true
	m.LapseTime = 0
	m.TotalStep = 8
	m.CurrentStep = 0
	m.msgCh = make(chan *InboxMessage, 1000)
	m.closeCh = make(chan bool)
	log.Println("Event:", event, " has been ready!")
	return &m

}

func (m *Match) Run() {
	dt := 10 * time.Millisecond
	tickChan := time.Tick(dt)
	for {
		<-tickChan
		m.handleInputs()
		m.tick(dt)
		if !m.IsGoing {
			break
		}
	}
}

func (m *Match) Stop() {
	m.IsGoing = false
	log.Println("event stop!")
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

func (m *Match) tick(dt time.Duration) {
	sec := dt.Seconds()
	switch m.Event {
	case CancleRobBar:
		addr := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-6-5"}
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("reset")
		m.srv.sendToOne(sendMsg, addr)
		m.srv.stopMatch()
	case EventRobBar:
		addr := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-6-5"}
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("loot")
		m.srv.sendToOne(sendMsg, addr)
		m.srv.stopMatch()
	case EventChallengeBilly:
		addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("mp3_ctrl")
		mp3 := make([]map[string]string, 0)
		mp3 = append(mp3,
			map[string]string{"mp3_n": "0", "music": "12"},
		)
		sendMsg.Set("mp3", mp3)
		m.srv.sendToOne(sendMsg, addr)
		m.srv.stopMatch()
	case EventToDay:
		m.LapseTime = math.Max(m.LapseTime-sec, 0)
		if m.LapseTime == 0 {
			log.Println("current Step:", m.CurrentStep)
			switch m.CurrentStep {
			case 0:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("mp3_ctrl")
				mp3 := make([]map[string]string, 0)
				mp3 = append(mp3,
					map[string]string{"mp3_n": "0", "music": "8"},
				)
				sendMsg.Set("mp3", mp3)
				m.srv.sendToOne(sendMsg, addr)

				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("led_ctrl")
				led := make([]map[string]string, 0)
				led = append(led,
					map[string]string{"led_n": "0", "mode": "0"},
				)
				sendMsg1.Set("led", led)
				m.srv.sends(sendMsg1, InboxAddressTypeNightArduino)
			case 1:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "0", "light_s": "1"},
					map[string]string{"light_n": "10", "light_s": "1"},
					map[string]string{"light_n": "11", "light_s": "1"},
					map[string]string{"light_n": "12", "light_s": "1"},
					map[string]string{"light_n": "19", "light_s": "1"},
					map[string]string{"light_n": "21", "light_s": "1"},
					map[string]string{"light_n": "22", "light_s": "1"},
					map[string]string{"light_n": "23", "light_s": "1"},
					map[string]string{"light_n": "25", "light_s": "1"},
					map[string]string{"light_n": "27", "light_s": "1"},
					map[string]string{"light_n": "28", "light_s": "1"},
					map[string]string{"light_n": "41", "light_s": "1"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-7-2"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "0", "light_s": "1"},
					map[string]string{"light_n": "1", "light_s": "1"},
					map[string]string{"light_n": "2", "light_s": "1"},
					map[string]string{"light_n": "3", "light_s": "1"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)
			case 2:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "3", "light_s": "1"},
					map[string]string{"light_n": "30", "light_s": "1"},
					map[string]string{"light_n": "42", "light_s": "1"},
					map[string]string{"light_n": "46", "light_s": "1"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)
			case 3:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "4", "light_s": "1"},
					map[string]string{"light_n": "6", "light_s": "1"},
					map[string]string{"light_n": "8", "light_s": "1"},
					map[string]string{"light_n": "9", "light_s": "1"},
					map[string]string{"light_n": "16", "light_s": "1"},
					map[string]string{"light_n": "43", "light_s": "1"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-4-1"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "0", "light_s": "1"},
					map[string]string{"light_n": "1", "light_s": "1"},
					map[string]string{"light_n": "2", "light_s": "1"},
					map[string]string{"light_n": "3", "light_s": "1"},
					map[string]string{"light_n": "4", "light_s": "1"},
					map[string]string{"light_n": "5", "light_s": "1"},
					map[string]string{"light_n": "9", "light_s": "1"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)
			case 4:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "17", "light_s": "1"},
					map[string]string{"light_n": "18", "light_s": "1"},
					map[string]string{"light_n": "35", "light_s": "1"},
					map[string]string{"light_n": "36", "light_s": "1"},
					map[string]string{"light_n": "39", "light_s": "1"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-2-1"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "2", "light_s": "1"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)
			case 5:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "13", "light_s": "1"},
					map[string]string{"light_n": "15", "light_s": "1"},
					map[string]string{"light_n": "31", "light_s": "1"},
					map[string]string{"light_n": "38", "light_s": "1"},
					map[string]string{"light_n": "40", "light_s": "1"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-2-1"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "0", "light_s": "1"},
					map[string]string{"light_n": "1", "light_s": "1"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)

				addr2 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-1-1"}
				sendMsg2 := NewInboxMessage()
				sendMsg2.SetCmd("light_ctrl")
				lights2 := make([]map[string]string, 0)
				lights2 = append(lights2,
					map[string]string{"light_n": "2", "light_s": "1"},
					map[string]string{"light_n": "3", "light_s": "1"},
					map[string]string{"light_n": "4", "light_s": "1"},
					map[string]string{"light_n": "5", "light_s": "1"},
					map[string]string{"light_n": "9", "light_s": "1"},
				)
				sendMsg2.Set("light", lights2)
				m.srv.sendToOne(sendMsg2, addr2)
			case 6:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "14", "light_s": "1"},
					map[string]string{"light_n": "24", "light_s": "1"},
					map[string]string{"light_n": "34", "light_s": "1"},
					map[string]string{"light_n": "45", "light_s": "1"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-1-1"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "0", "light_s": "1"},
					map[string]string{"light_n": "1", "light_s": "1"},
					map[string]string{"light_n": "6", "light_s": "1"},
					map[string]string{"light_n": "8", "light_s": "1"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)

				addr2 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-9-1"}
				sendMsg2 := NewInboxMessage()
				sendMsg2.SetCmd("light_ctrl")
				lights2 := make([]map[string]string, 0)
				lights2 = append(lights2,
					map[string]string{"light_n": "0", "light_s": "1"},
				)
				sendMsg2.Set("light", lights2)
				m.srv.sendToOne(sendMsg2, addr2)
				m.srv.stopMatch()
			}
			m.LapseTime = GetOptions().LapseTime
			m.CurrentStep++
		}
	case EventToNight:
		m.LapseTime = math.Max(m.LapseTime-sec, 0)
		if m.LapseTime == 0 {
			switch m.CurrentStep {
			case 0:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("mp3_ctrl")
				mp3 := make([]map[string]string, 0)
				mp3 = append(mp3,
					map[string]string{"mp3_n": "0", "music": "10"},
				)
				sendMsg.Set("mp3", mp3)
				m.srv.sendToOne(sendMsg, addr)

				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("led_ctrl")
				led := make([]map[string]string, 0)
				led = append(led,
					map[string]string{"led_n": "0", "mode": "1"},
				)
				sendMsg1.Set("led", led)
				m.srv.sends(sendMsg1, InboxAddressTypeNightArduino)
			case 1:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "0", "light_s": "0"},
					map[string]string{"light_n": "10", "light_s": "0"},
					map[string]string{"light_n": "11", "light_s": "0"},
					map[string]string{"light_n": "12", "light_s": "0"},
					map[string]string{"light_n": "19", "light_s": "0"},
					map[string]string{"light_n": "21", "light_s": "0"},
					map[string]string{"light_n": "22", "light_s": "0"},
					map[string]string{"light_n": "23", "light_s": "0"},
					map[string]string{"light_n": "25", "light_s": "0"},
					map[string]string{"light_n": "27", "light_s": "0"},
					map[string]string{"light_n": "28", "light_s": "0"},
					map[string]string{"light_n": "41", "light_s": "0"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-7-2"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "0", "light_s": "0"},
					map[string]string{"light_n": "1", "light_s": "0"},
					map[string]string{"light_n": "2", "light_s": "0"},
					map[string]string{"light_n": "3", "light_s": "0"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)
			case 2:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "3", "light_s": "0"},
					map[string]string{"light_n": "30", "light_s": "0"},
					map[string]string{"light_n": "42", "light_s": "0"},
					map[string]string{"light_n": "46", "light_s": "0"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)
			case 3:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "4", "light_s": "0"},
					map[string]string{"light_n": "6", "light_s": "0"},
					map[string]string{"light_n": "8", "light_s": "0"},
					map[string]string{"light_n": "9", "light_s": "0"},
					map[string]string{"light_n": "16", "light_s": "0"},
					map[string]string{"light_n": "43", "light_s": "0"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-4-1"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "0", "light_s": "0"},
					map[string]string{"light_n": "1", "light_s": "0"},
					map[string]string{"light_n": "2", "light_s": "0"},
					map[string]string{"light_n": "3", "light_s": "0"},
					map[string]string{"light_n": "4", "light_s": "0"},
					map[string]string{"light_n": "5", "light_s": "0"},
					map[string]string{"light_n": "9", "light_s": "0"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)
			case 4:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "17", "light_s": "0"},
					map[string]string{"light_n": "18", "light_s": "0"},
					map[string]string{"light_n": "35", "light_s": "0"},
					map[string]string{"light_n": "36", "light_s": "0"},
					map[string]string{"light_n": "39", "light_s": "0"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-2-1"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "2", "light_s": "0"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)
			case 5:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "13", "light_s": "0"},
					map[string]string{"light_n": "15", "light_s": "0"},
					map[string]string{"light_n": "31", "light_s": "0"},
					map[string]string{"light_n": "38", "light_s": "0"},
					map[string]string{"light_n": "40", "light_s": "0"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-2-1"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "0", "light_s": "0"},
					map[string]string{"light_n": "1", "light_s": "0"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)

				addr2 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-1-1"}
				sendMsg2 := NewInboxMessage()
				sendMsg2.SetCmd("light_ctrl")
				lights2 := make([]map[string]string, 0)
				lights2 = append(lights2,
					map[string]string{"light_n": "2", "light_s": "0"},
					map[string]string{"light_n": "3", "light_s": "0"},
					map[string]string{"light_n": "4", "light_s": "0"},
					map[string]string{"light_n": "5", "light_s": "0"},
					map[string]string{"light_n": "9", "light_s": "0"},
				)
				sendMsg2.Set("light", lights2)
				m.srv.sendToOne(sendMsg2, addr2)
			case 6:
				addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
				sendMsg := NewInboxMessage()
				sendMsg.SetCmd("light_ctrl")
				lights := make([]map[string]string, 0)
				lights = append(lights,
					map[string]string{"light_n": "14", "light_s": "0"},
					map[string]string{"light_n": "24", "light_s": "0"},
					map[string]string{"light_n": "34", "light_s": "0"},
					map[string]string{"light_n": "45", "light_s": "0"},
				)
				sendMsg.Set("light", lights)
				m.srv.sendToOne(sendMsg, addr)

				addr1 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-1-1"}
				sendMsg1 := NewInboxMessage()
				sendMsg1.SetCmd("light_ctrl")
				lights1 := make([]map[string]string, 0)
				lights1 = append(lights1,
					map[string]string{"light_n": "0", "light_s": "0"},
					map[string]string{"light_n": "1", "light_s": "0"},
					map[string]string{"light_n": "6", "light_s": "0"},
					map[string]string{"light_n": "8", "light_s": "0"},
				)
				sendMsg1.Set("light", lights1)
				m.srv.sendToOne(sendMsg1, addr1)

				addr2 := InboxAddress{InboxAddressTypeGameArduinoDevice, "G-9-1"}
				sendMsg2 := NewInboxMessage()
				sendMsg2.SetCmd("light_ctrl")
				lights2 := make([]map[string]string, 0)
				lights2 = append(lights2,
					map[string]string{"light_n": "0", "light_s": "0"},
				)
				sendMsg2.Set("light", lights2)
				m.srv.sendToOne(sendMsg2, addr2)
				m.srv.stopMatch()
			}
			m.LapseTime = GetOptions().LapseTime
			m.CurrentStep++
		}
	case EventRecoverDay:
		addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("mp3_ctrl")
		mp3 := make([]map[string]string, 0)
		mp3 = append(mp3,
			map[string]string{"mp3_n": "0", "music": "9"},
		)
		sendMsg.Set("mp3", mp3)
		m.srv.sendToOne(sendMsg, addr)
		m.srv.stopMatch()
	case EventRecoverNight:
		addr := InboxAddress{InboxAddressTypeDjArduino, "D-1"}
		sendMsg := NewInboxMessage()
		sendMsg.SetCmd("mp3_ctrl")
		mp3 := make([]map[string]string, 0)
		mp3 = append(mp3,
			map[string]string{"mp3_n": "0", "music": "11"},
		)
		sendMsg.Set("mp3", mp3)
		m.srv.sendToOne(sendMsg, addr)
		m.srv.stopMatch()
	}
}

func (m *Match) updateStage() {

}

func (m *Match) bgmPlay(bgm int) {
	sendMsg := NewInboxMessage()
	sendMsg.SetCmd("mp3_ctrl")
	sendMsg.Set("music", strconv.Itoa(bgm))
	addr := InboxAddress{InboxAddressTypeDjArduino, "B-1"}
	m.srv.sendToOne(sendMsg, addr)

}

func att(id string) InboxAddressType {
	if id == "" {
		return InboxAddressTypeUnknown
	} else if strings.HasPrefix(id, "G") {
		return InboxAddressTypeGameArduinoDevice
	} else if strings.HasPrefix(id, "N") {
		return InboxAddressTypeNightArduino
	} else if strings.HasPrefix(id, "B") {
		return InboxAddressTypeBoxArduinoDevice
	} else if strings.HasPrefix(id, "D") {
		return InboxAddressTypeDjArduino
	}
	return InboxAddressTypeUnknown
}
