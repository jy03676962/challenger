package core

import (
	"fmt"
)

type InboxAddressType int

const (
	InboxAddressTypeUnknown           = 0
	InboxAddressTypeAdminDevice       = 1 // 管理员屏幕
	InboxAddressTypeGameArduinoDevice = 2 // 游戏 Arduino
	InboxAddressTypeBoxArduinoDevice  = 3 // 箱子 Arduino
	InboxAddressTypeNightArduino      = 4 // 垃圾桶 arduino
	InboxAddressTypeDjArduino         = 5 // dj台 arduino
)

func (t InboxAddressType) IsArduinoControllerType() bool {
	return t == InboxAddressTypeGameArduinoDevice || t == InboxAddressTypeBoxArduinoDevice ||
		t == InboxAddressTypeNightArduino || t == InboxAddressTypeDjArduino
}

type InboxAddress struct {
	Type InboxAddressType `json:"type"`
	ID   string           `json:"id"`
}

func (addr InboxAddress) String() string {
	return fmt.Sprintf("%v:%v", addr.Type, addr.ID)
}
