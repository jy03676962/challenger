package core

const (
	api                       = ""
	AuthorityGet              = api + "1"
	GameDataAdivinacionCreate = api + "2"
	GameDataAdivinacionModify = api + "3"
	GameDataBangCreate        = api + "4"
	GameDataBangModify        = api + "5"
	GameDataFollowCreate      = api + "6"
	GameDataFollowModify      = api + "7"
	GameDataGreetingCreate    = api + "8"
	GameDataGreetingModify    = api + "9"
	GameDataHighnoonCreate    = api + "10"
	GameDataHighnoonModify    = api + "11"
	GameDataHunterCreate      = api + "12"
	GameDataHunterModify      = api + "13"
	GameDataHunterBoxCreate   = api + "14"
	GameDataHunterBoxModify   = api + "15"
	GameDataMarksmanCreate    = api + "16"
	GameDataMarksmanModify    = api + "17"
	GameDataMinerCreate       = api + "18"
	GameDataMinerModify       = api + "19"
	GameDataPrivityCreate     = api + "20"
	GameDataPrivityModify     = api + "21"
	GameDataRussianCreate     = api + "22"
	GameDataRussianModify     = api + "23"
	TicketUse                 = api + "24"
	TicketCheck               = api + "25"
)

type HttpResponse struct {
	Data      string
	Api       string
	ArduinoId string
}
