package core

import (
	"encoding/json"
)

type InboxMessage struct {
	Data                  map[string]interface{}
	Address               *InboxAddress
	RemoveAddress         *InboxAddress
	AddAddress            *InboxAddress
	ShouldCloseConnection bool
}

func NewInboxMessage() *InboxMessage {
	msg := InboxMessage{}
	msg.Data = make(map[string]interface{})
	return &msg
}

func NewErrorInboxMessage(errMsg string) *InboxMessage {
	msg := NewInboxMessage()
	msg.SetCmd("error")
	msg.Set("msg", errMsg)
	return msg
}

func (message *InboxMessage) Get(key string) interface{} {
	if v, ok := message.Data[key]; ok {
		return v
	}
	return nil
}

func (message *InboxMessage) Set(key string, value interface{}) {
	message.Data[key] = value
}

func (message *InboxMessage) GetStr(key string) string {
	if v, ok := message.Data[key]; ok {
		return v.(string)
	}
	return ""
}

func (message *InboxMessage) GetCmd() string {
	return message.GetStr("cmd")
}

func (message *InboxMessage) SetCmd(v string) {
	message.Set("cmd", v)
}

func (message *InboxMessage) Marshal() (b []byte, e error) {
	b, e = json.Marshal(message.Data)
	return
}

func (message *InboxMessage) Empty() bool {
	return len(message.Data) == 0
}
