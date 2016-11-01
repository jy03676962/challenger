package core

type ArduinoMode int

const ArduinoModeUnknown ArduinoMode = -1

const (
	ArduinoModeOff = iota
	ArduinoModeOn
	//ArduinoModeScan
	//ArduinoModeFree
)

type ArduinoController struct {
	Address InboxAddress `json:"address"`
	ID      string       `json:"id"`
	Mode    ArduinoMode  `json:"mode"`
	Online  bool         `json:"online"`
	//ScoreUpdated bool         `json:"scoreUpdated"`
}

func NewArduinoController(addr InboxAddress) *ArduinoController {
	a := ArduinoController{}
	a.Address = addr
	a.ID = addr.String()
	a.Mode = ArduinoModeUnknown
	a.Online = false
	//a.ScoreUpdated = false
	return &a
}
