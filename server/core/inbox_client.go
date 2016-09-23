package core

import (
	"log"
)

var _ = log.Println

type InboxClient struct {
	conn  InboxConnection
	id    int
	inbox *Inbox
}

func NewInboxClient(conn InboxConnection, inbox *Inbox, id int) *InboxClient {
	client := InboxClient{}
	client.conn = conn
	client.id = id
	client.inbox = inbox
	return &client
}

func (c *InboxClient) Listen() {
	c.listenRead()
	c.conn.Close()
	c.inbox.RemoveClient(c.id)
}

func (c *InboxClient) Accept(addr InboxAddress) bool {
	return c.conn.Accept(addr)
}

func (c *InboxClient) Write(msg *InboxMessage) {
	go func() {
		e := c.conn.WriteJSON(msg)
		if e != nil {
			log.Printf("send message error:%v\n", e.Error())
		}
	}()
}

func (c *InboxClient) listenRead() {
	for {
		m := NewInboxMessage()
		e := c.conn.ReadJSON(m)
		if e != nil {
			log.Printf("read message error:%v\n", e.Error())
		}
		if !m.Empty() || m.RemoveAddress != nil || m.AddAddress != nil {
			c.inbox.ReceiveMessage(m)
		}
		if m.ShouldCloseConnection {
			return
		}
	}

}
