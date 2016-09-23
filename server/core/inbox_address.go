package core

import (
	"fmt"
)

type InboxAddressType int

const (
	InboxAddressTypeUnknown           = 0
	InboxAddressTypeAdminDevice       = 1  // 管理员iPad
	InboxAddressTypeSimulatorDevice   = 2  // 模拟器
	InboxAddressTypeArduinoTestDevice = 3  //测试Arduino设备
	InboxAddressTypePostgameDevice    = 4  // 出口处iPad
	InboxAddressTypeWearableDevice    = 5  // 穿戴设备
	InboxAddressTypeMainArduinoDevice = 6  // Arduino主墙设备
	InboxAddressTypeSubArduinoDevice  = 7  // Arduino小墙设备
	InboxAddressTypeQueueDevice       = 8  // 叫号屏幕
	InboxAddressTypeIngameDevice      = 9  // 游戏内屏幕
	InboxAddressTypeMusicArduino      = 10 // music arduino
	InboxAddressTypeDoorArduino       = 10 // door arduino
)

func (t InboxAddressType) IsPlayerControllerType() bool {
	return t == InboxAddressTypeSimulatorDevice || t == InboxAddressTypeWearableDevice
}

func (t InboxAddressType) IsArduinoControllerType() bool {
	return t == InboxAddressTypeMainArduinoDevice || t == InboxAddressTypeSubArduinoDevice ||
		t == InboxAddressTypeMusicArduino || t == InboxAddressTypeDoorArduino
}

type InboxAddress struct {
	Type InboxAddressType `json:"type"`
	ID   string           `json:"id"`
}

func (addr InboxAddress) String() string {
	return fmt.Sprintf("%v:%v", addr.Type, addr.ID)
}
