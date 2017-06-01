package core

const (
	api                       = "http://172.16.10.7/gsaleapi/"
	AuthorityGet              = api + "card_list.php"
	GameDataAdivinacionCreate = api + "gamedata_adivinacion.php"
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
	Data     string
	JsonData map[string]interface{}
	Api      string
	Msg      *InboxMessage
	//ArduinoId string //该消息收到后应该反馈的arduinoID
	//CardId    string //请求该消息的CardId
}

func NewHttpResponse() *HttpResponse {
	res := HttpResponse{}
	res.JsonData = make(map[string]interface{})
	return &res
}

func (res *HttpResponse) Get(key string) interface{} {
	if v, ok := res.JsonData[key]; ok {
		return v
	}
	return nil
}

func (res *HttpResponse) Set(key string, value interface{}) {
	res.JsonData[key] = value
}

func (res *HttpResponse) GetStr(key string) string {
	if v, ok := res.JsonData[key]; ok {
		return v.(string)
	}
	return ""
}
