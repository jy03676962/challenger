package core

import (
	"log"
	"sync"
)

var _ = log.Println

type Inbox struct {
	srv   *Srv
	cdict map[int]*InboxClient
	curID int
	l     *sync.RWMutex
}

type p struct {
	msg  *InboxMessage
	addr InboxAddress
}

func NewInbox(srv *Srv) *Inbox {
	inbox := Inbox{}
	inbox.srv = srv
	inbox.cdict = make(map[int]*InboxClient)
	inbox.curID = 1
	inbox.l = new(sync.RWMutex)
	return &inbox
}

func (inbox *Inbox) ListenConnection(conn InboxConnection) {
	inbox.l.Lock()
	c := NewInboxClient(conn, inbox, inbox.curID)
	inbox.cdict[inbox.curID] = c
	inbox.curID += 1
	log.Printf("inbox got connection, current:%v\n", len(inbox.cdict))
	inbox.l.Unlock()
	c.Listen()
}

func (inbox *Inbox) RemoveClient(id int) {
	inbox.l.Lock()
	defer inbox.l.Unlock()
	log.Printf("inbox remove connection:%v, current:%v\n", id, len(inbox.cdict))
	delete(inbox.cdict, id)
}

func (inbox *Inbox) ReceiveMessage(m *InboxMessage) {
	inbox.srv.onInboxMessageArrived(m)
}

func (inbox *Inbox) Send(msg *InboxMessage, addrs []InboxAddress) {
	inbox.l.RLock()
	defer inbox.l.RUnlock()
	for _, cli := range inbox.cdict {
		for _, addr := range addrs {
			if cli.Accept(addr) {
				cli.Write(msg)
			}
		}
	}
}
