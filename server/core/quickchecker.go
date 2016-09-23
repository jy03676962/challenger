package core

import (
	"log"
	"strconv"
	"strings"
	"time"
)

var _ = log.Println

type ReceiverStatus int

const (
	ReceiverStatusUnknown           = 0
	ReceiverStatusBroken            = 1
	ReceiverStatusBrokenButReceived = 2
	ReceiverStatusNotReceived       = 3
	ReceiverStatusNormal            = 4
)

type QuickChecker struct {
	srv       *Srv
	receivers map[string]bool
	statusMap map[string]ReceiverStatus
	enterCh   chan string
	leaveCh   chan string
	closeCh   chan struct{}
}

func NewQuickChecker(srv *Srv) *QuickChecker {
	qc := QuickChecker{}
	qc.srv = srv
	qc.receivers = GetLaserPair().GetValidReceivers(true)
	qc.statusMap = make(map[string]ReceiverStatus)
	qc.enterCh = make(chan string)
	qc.leaveCh = make(chan string)
	qc.closeCh = make(chan struct{})
	for k, _ := range qc.receivers {
		qc.statusMap[k] = ReceiverStatusUnknown
	}
	qc.toggleAllLasers("1")
	go qc.blinkLoop()
	return &qc
}

func (qc *QuickChecker) OnArduinoHeartBeat(hb *InboxMessage) {
	ur := hb.GetStr("UR")
	id := hb.Address.ID
	for i, r := range ur {
		c := string(r)
		key := id + ":" + strconv.Itoa(i)
		if c == "1" {
			if _, ok := qc.receivers[key]; ok {
				if qc.statusMap[key] == ReceiverStatusNotReceived {
					qc.stopBlink(key)
				}
				qc.statusMap[key] = ReceiverStatusNormal
			} else {
				qc.statusMap[key] = ReceiverStatusBrokenButReceived
			}
		} else {
			if _, ok := qc.receivers[key]; ok {
				if qc.statusMap[key] != ReceiverStatusNotReceived {
					qc.blink(key)
				}
				qc.statusMap[key] = ReceiverStatusNotReceived
			} else {
				qc.statusMap[key] = ReceiverStatusBroken
			}
		}
	}
}

func (qc *QuickChecker) Query() {
	msg := NewInboxMessage()
	msg.SetCmd("QuickCheck")
	data := make(map[string]ReceiverStatus)
	for k, v := range qc.statusMap {
		data[k] = v
	}
	msg.Set("data", data)
	qc.srv.sends(msg, InboxAddressTypeAdminDevice)
}

func (qc *QuickChecker) Stop(save bool) {
	if save {
		qc.record()
	}
	close(qc.closeCh)
}

func (qc *QuickChecker) record() {
	ret := make([]string, 0)
	for k, v := range qc.statusMap {
		if v == ReceiverStatusNotReceived {
			ret = append(ret, k)
		}
	}
	GetLaserPair().RecordBrokens(ret)
}

func (qc *QuickChecker) toggleAllLasers(status string) {
	senders := GetLaserPair().GetValidSenders()
	for id, li := range senders {
		info := arduinoInfoFromID(id)
		msg := NewInboxMessage()
		msg.SetCmd("laser_ctrl")
		laserList := make([]map[string]string, len(li))
		for i, v := range li {
			v += 1
			if info.LaserNum == 5 {
				v += 5
			}
			laser := make(map[string]string)
			laser["laser_s"] = status
			laser["laser_n"] = strconv.Itoa(v)
			laserList[i] = laser
		}
		msg.Set("laser", laserList)
		qc.srv.sendToOne(msg, InboxAddress{InboxAddressTypeMainArduinoDevice, id})
	}
}

func (qc *QuickChecker) blinkLoop() {
	blinkSenders := make(map[string]bool)
	isOpen := false
	tickCh := time.Tick(1000 * time.Millisecond)
	for {
		select {
		case sender := <-qc.enterCh:
			blinkSenders[sender] = true
		case sender := <-qc.leaveCh:
			delete(blinkSenders, sender)
			li := strings.Split(sender, ":")
			idx, _ := strconv.Atoi(li[1])
			qc.srv.laserControl(li[0], idx, true)
		case <-tickCh:
			laser := make(map[string][]string)
			for sender, _ := range blinkSenders {
				id, idx := parseSender(sender)
				if v, ok := laser[id]; ok {
					laser[id] = append(v, idx)
				} else {
					laser[id] = []string{idx}
				}
			}
			for id, li := range laser {
				msg := NewInboxMessage()
				msg.SetCmd("laser_ctrl")
				laserList := make([]map[string]string, len(li))
				for i, v := range li {
					laser := make(map[string]string)
					if isOpen {
						laser["laser_s"] = "1"
					} else {
						laser["laser_s"] = "0"
					}
					laser["laser_n"] = v
					laserList[i] = laser
				}
				msg.Set("laser", laserList)
				qc.srv.sendToOne(msg, InboxAddress{InboxAddressTypeMainArduinoDevice, id})
			}
			isOpen = !isOpen
		case <-qc.closeCh:
			qc.toggleAllLasers("0")
			return
		}
	}
}

func (qc *QuickChecker) blink(receiver string) {
	if info, sender := GetLaserPair().FindByReceiver(receiver); info != nil {
		qc.enterCh <- sender
	}
}

func (qc *QuickChecker) stopBlink(receiver string) {
	if info, sender := GetLaserPair().FindByReceiver(receiver); info != nil {
		qc.leaveCh <- sender
	}
}

func parseSender(sender string) (i string, ix string) {
	li := strings.Split(sender, ":")
	id := li[0]
	info := arduinoInfoFromID(id)
	idx, _ := strconv.Atoi(li[1])
	idx += 1
	if info.LaserNum == 5 {
		idx += 5
	}
	return id, strconv.Itoa(idx)
}
