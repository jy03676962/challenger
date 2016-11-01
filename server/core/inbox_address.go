package core

import (
	"fmt"
)

type InboxAddressType int

const (
	InboxAddressTypeUnknown            = 0
	InboxAddressTypeAdminDevice        = 1 // 管理员屏幕
	InboxAddressTypeRoomArduinoDevice  = 2 // 房间Arduino
	InboxAddressTypeLightArduinoDevice = 3 // 环境灯光Arduino
	InboxAddressTypeMusicArduino       = 4 // music arduino
	InboxAddressTypeDoorArduino        = 5 // door arduino
)

func (t InboxAddressType) IsArduinoControllerType() bool {
	return t == InboxAddressTypeRoomArduinoDevice || t == InboxAddressTypeLightArduinoDevice ||
		t == InboxAddressTypeMusicArduino || t == InboxAddressTypeDoorArduino
}

type InboxAddress struct {
	Type InboxAddressType `json:"type"`
	ID   string           `json:"id"`
}

func (addr InboxAddress) String() string {
	return fmt.Sprintf("%v:%v", addr.Type, addr.ID)
}
